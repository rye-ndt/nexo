package helper_agent

import (
	"strings"
	"sync"
	"time"

	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	core_itf "hexago/internal/interface/core"
	output_itf "hexago/internal/interface/output"

	"github.com/google/uuid"
)

const (
	draftTimeout    = 5 * time.Minute
	livenessCheck   = 15 * time.Second
	maxParams       = 12
	maxPromptLength = 20000
)

type v1 struct {
	locker  sync.Mutex
	agents  core_itf.AgentManager
	userCfg output_itf.UserConfig
	logger  output_itf.Logger

	// One draft at a time, so an abandoned dialog cannot hold a harness slot for
	// the whole of draftTimeout.
	drafting uuid.UUID
	inbox    chan *core_itf.Template
}

func InitV1(
	agents core_itf.AgentManager,
	userCfg output_itf.UserConfig,
	logger output_itf.Logger,
) (core_itf.TemplateHelper, error) {
	if agents == nil || userCfg == nil || logger == nil {
		return nil, custom_error.Critical("template helper needs an agent manager, a user config and a logger")
	}

	return &v1{agents: agents, userCfg: userCfg, logger: logger}, nil
}

// Drafting a template is heavy work, so it borrows the heavy task level's model
// rather than introducing a setting of its own.
func (s *v1) Blocked() string {
	_, reason := s.heavyAgent()

	return reason
}

func (s *v1) heavyAgent() (*output_itf.AgentDefault, string) {
	agentDefault, err := s.userCfg.AgentDefault(enums.HeavyTask)
	if err != nil || agentDefault == nil {
		return nil, "No model is set for heavy tasks yet. Pick one in Settings."
	}

	harness := agentDefault.Model.HarnessTool()

	admin, err := s.agents.Admin(harness)
	if err != nil {
		return nil, "Heavy tasks run on a tool this build does not have."
	}

	installed, err := admin.Status()
	if err != nil || installed == nil {
		return nil, "Could not tell whether " + harness.DisplayName() + " is ready."
	}

	switch {
	case !installed.Installed:
		return nil, "Install " + harness.DisplayName() + " to use this."
	case !installed.LoggedIn:
		return nil, "Log in to " + harness.DisplayName() + " to use this."
	default:
		return agentDefault, ""
	}
}

func (s *v1) Draft(name string, role string) (*core_itf.Template, error) {
	if strings.TrimSpace(name) == "" {
		return nil, custom_error.Critical("a template needs a name before it can be filled in")
	}

	agentDefault, reason := s.heavyAgent()
	if reason != "" {
		return nil, custom_error.Critical("%s", reason)
	}

	// No working directory: the agent lands in its own throwaway workspace instead
	// of whatever project the user happens to have open. Writing a template needs
	// no files.
	agent, err := s.agents.RequestInstance(&core_itf.AgentRequest{
		Name:          agentDefault.Model,
		ThinkingLevel: agentDefault.ThinkingLevel,
		SystemPrompts: []string{systemPrompt},
	})
	if err != nil {
		return nil, err
	}

	defer s.kill(agent.ID)

	// Registered before the prompt is sent, so the agent cannot report into a gap.
	delivered := s.await(agent.ID)

	if err := s.agents.Send(agent.ID, buildPrompt(name, role)); err != nil {
		return nil, custom_error.Critical("cannot reach the assistant filling this in: %v", err)
	}

	return s.wait(agent.ID, delivered)
}

func (s *v1) wait(
	agentID uuid.UUID,
	delivered <-chan *core_itf.Template,
) (*core_itf.Template, error) {
	deadline := time.NewTimer(draftTimeout)
	defer deadline.Stop()

	ticker := time.NewTicker(livenessCheck)
	defer ticker.Stop()

	for {
		select {
		case template := <-delivered:
			return template, nil
		case <-deadline.C:
			return nil, custom_error.Critical(
				"the assistant took too long to fill this in, so nothing was changed",
			)
		case <-ticker.C:
			if err := s.agents.HeartBeat(agentID); err != nil {
				return nil, custom_error.Critical(
					"the assistant stopped responding before it finished, so nothing was changed",
				)
			}
		}
	}
}

func (s *v1) Drafting(agentID uuid.UUID) bool {
	s.locker.Lock()
	defer s.locker.Unlock()

	return agentID != uuid.Nil && s.drafting == agentID
}

// Deliver is what the report_template MCP tool calls. A template that does not pass
// validateDraft is refused, and the refusal travels back to the agent as a tool
// error naming what is wrong, so it gets to correct itself rather than handing the
// user something half-written.
func (s *v1) Deliver(agentID uuid.UUID, template *core_itf.Template) error {
	if template == nil {
		return custom_error.Critical("the template is empty")
	}

	if err := validateDraft(template); err != nil {
		return err
	}

	s.locker.Lock()
	defer s.locker.Unlock()

	if agentID == uuid.Nil || s.drafting != agentID {
		return custom_error.Critical("nobody is waiting for a template from this agent")
	}

	select {
	case s.inbox <- template:
		return nil
	default:
		return custom_error.Critical("a template was already submitted for this request")
	}
}

func (s *v1) await(agentID uuid.UUID) <-chan *core_itf.Template {
	inbox := make(chan *core_itf.Template, 1)

	s.locker.Lock()
	previous := s.drafting
	s.drafting, s.inbox = agentID, inbox
	s.locker.Unlock()

	if previous != uuid.Nil {
		s.kill(previous)
	}

	return inbox
}

func (s *v1) kill(agentID uuid.UUID) {
	s.locker.Lock()
	if s.drafting == agentID {
		s.drafting, s.inbox = uuid.Nil, nil
	}
	s.locker.Unlock()

	if err := s.agents.Kill(agentID); err != nil {
		s.logger.Warn("template helper cleanup", "agent", agentID, "err", err)
	}
}
