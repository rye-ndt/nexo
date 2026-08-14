package wails_api

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/google/uuid"

	"hexago/internal/helpers"
	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	core_itf "hexago/internal/interface/core"
	input_itf "hexago/internal/interface/input"
	output_itf "hexago/internal/interface/output"
)

const installProgressEvent = "harness:install:progress"

const (
	idApproval = "approval"
	idTemplate = "template"
	idSession  = "session"
	idTask     = "task"
	idDraft    = "session draft"
)

type API struct {
	ctx          context.Context
	agentManager core_itf.AgentManager
	mcpProxy     core_itf.MCPProxyServer
	approvals    core_itf.ApprovalBroker
	templates    core_itf.AgentTemplateManager
	sessions     core_itf.SessionManager
	coordinator  core_itf.Coordinator
	history      input_itf.WorkspaceHistory
	userConfig   output_itf.UserConfig
	drafts       input_itf.DraftStorage
	templateHelp core_itf.TemplateHelper
}

var _ output_itf.FEAPI = (*API)(nil)

type Deps struct {
	AgentManager   core_itf.AgentManager
	MCPProxy       core_itf.MCPProxyServer
	Approvals      core_itf.ApprovalBroker
	Templates      core_itf.AgentTemplateManager
	Sessions       core_itf.SessionManager
	Coordinator    core_itf.Coordinator
	History        input_itf.WorkspaceHistory
	UserConfig     output_itf.UserConfig
	Drafts         input_itf.DraftStorage
	TemplateHelper core_itf.TemplateHelper
}

func New(deps *Deps) *API {
	return &API{
		agentManager: deps.AgentManager,
		mcpProxy:     deps.MCPProxy,
		approvals:    deps.Approvals,
		templates:    deps.Templates,
		sessions:     deps.Sessions,
		coordinator:  deps.Coordinator,
		history:      deps.History,
		userConfig:   deps.UserConfig,
		drafts:       deps.Drafts,
		templateHelp: deps.TemplateHelper,
	}
}

// Startup is wired to Wails OnStartup; it is not meant to be called from JS.
func (a *API) Startup(ctx context.Context) {
	a.ctx = ctx
}

// Shutdown is wired to Wails OnShutdown; it is not meant to be called from JS.
func (a *API) Shutdown(ctx context.Context) {
	if a.coordinator != nil {
		a.coordinator.Stop()
	}

	if a.sessions != nil {
		a.sessions.Stop()
	}

	if a.approvals != nil {
		a.approvals.Stop()
	}

	agents, err := a.agentManager.SupportedAgents()
	if err != nil {
		return
	}

	for agent := range agents {
		h, err := a.agentManager.Admin(agent)
		if err != nil {
			continue
		}
		h.Shutdown()
	}

	a.mcpProxy.Close()
}

func parseID(kind, id string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, custom_error.Critical("invalid %s id %s: %v", kind, id, err)
	}

	return parsed, nil
}

func withID(kind, id string, do func(uuid.UUID) error) error {
	parsed, err := parseID(kind, id)
	if err != nil {
		return err
	}

	return do(parsed)
}

func fromID[T any](kind, id string, do func(uuid.UUID) (T, error)) (T, error) {
	parsed, err := parseID(kind, id)
	if err != nil {
		var zero T

		return zero, err
	}

	return do(parsed)
}

func withSessionTask(sessionID, taskID string, do func(session, task uuid.UUID) error) error {
	return withID(idSession, sessionID, func(session uuid.UUID) error {
		return withID(idTask, taskID, func(task uuid.UUID) error {
			return do(session, task)
		})
	})
}

func (a *API) admin(id string) (input_itf.AgentAdmin, error) {
	return a.agentManager.Admin(enums.AgentHarness(id))
}

func (a *API) AgentStatuses() ([]output_itf.AgentInfo, error) {
	agents, err := a.agentManager.SupportedAgents()
	if err != nil {
		return nil, err
	}

	infos := make([]output_itf.AgentInfo, 0, len(agents))
	for agent := range agents {
		h, err := a.agentManager.Admin(agent)
		if err != nil {
			return nil, err
		}

		status, err := h.Status()
		if err != nil {
			return nil, err
		}

		infos = append(infos, output_itf.AgentInfo{ID: agent.String(), Status: status})
	}

	sort.Slice(infos, func(i, j int) bool { return infos[i].ID < infos[j].ID })

	return infos, nil
}

func (a *API) InstallAgent(id string) error {
	h, err := a.admin(id)
	if err != nil {
		return err
	}

	return h.Install(func(p input_itf.InstallProgress) {
		runtime.EventsEmit(a.ctx, installProgressEvent, id, p)
	})
}

func (a *API) AuthAgent(id string) (string, error) {
	h, err := a.admin(id)
	if err != nil {
		return "", err
	}

	return h.Auth()
}

func (a *API) SubmitAuthCode(id string, code string) error {
	h, err := a.admin(id)
	if err != nil {
		return err
	}

	return h.SubmitAuthCode(code)
}

func (a *API) PendingApprovals() ([]*output_itf.ApprovalInfo, error) {
	pending := a.approvals.Pending()

	infos := make([]*output_itf.ApprovalInfo, 0, len(pending))

	for _, request := range pending {
		item := approvalInfo(request)
		infos = append(infos, item)
	}

	return infos, nil
}

func approvalInfo(request *core_itf.ApprovalRequest) *output_itf.ApprovalInfo {
	if request == nil {
		return nil
	}

	return &output_itf.ApprovalInfo{
		ID:          request.ID.String(),
		AgentID:     request.AgentID.String(),
		TaskID:      request.TaskID.String(),
		Kind:        request.Kind.String(),
		Question:    request.Question,
		Detail:      request.Detail,
		Options:     request.Options,
		MultiSelect: request.MultiSelect,
		RequestedAt: request.RequestedAt.Format(time.RFC3339),
	}
}

func (a *API) AnswerApproval(requestID string, approved bool, optionIDs []string, guidance string) error {
	return withID(idApproval, requestID, func(parsed uuid.UUID) error {
		var sendErr error

		if !approved && guidance != "" {
			agentID := a.approvalAgent(parsed)
			if agentID == uuid.Nil {
				return custom_error.Critical("approval %s has no agent to send the guidance to", requestID)
			}

			sendErr = a.agentManager.Send(agentID, guidance)
		}

		if err := a.approvals.Answer(&core_itf.ApprovalAnswer{
			RequestID: parsed,
			Approved:  approved,
			OptionIDs: optionIDs,
			Guidance:  guidance,
		}); err != nil {
			return err
		}

		return sendErr
	})
}

func (a *API) approvalAgent(requestID uuid.UUID) uuid.UUID {
	for _, request := range a.approvals.Pending() {
		if request.ID == requestID {
			return request.AgentID
		}
	}

	return uuid.Nil
}

func (a *API) Templates() ([]*output_itf.TemplateInfo, error) {
	stored, err := a.templates.List()
	if err != nil {
		return nil, err
	}

	infos := make([]*output_itf.TemplateInfo, 0, len(stored))

	for _, template := range stored {
		infos = append(infos, templateInfo(template))
	}

	return infos, nil
}

func (a *API) Template(id string) (*output_itf.TemplateInfo, error) {
	return fromID(idTemplate, id, func(parsed uuid.UUID) (*output_itf.TemplateInfo, error) {
		template, err := a.templates.Get(parsed)
		if err != nil {
			return nil, err
		}

		return templateInfo(template), nil
	})
}

func (a *API) UpsertTemplate(template *output_itf.TemplateInfo) (string, error) {
	if template == nil {
		return "", custom_error.Critical("template is empty")
	}

	id := uuid.Nil

	if template.ID != "" {
		parsed, err := parseID(idTemplate, template.ID)
		if err != nil {
			return "", err
		}

		id = parsed
	}

	params := map[string]*core_itf.TemplateParams{}

	for key, param := range template.Params {
		if param == nil {
			params[key] = nil
			continue
		}

		params[key] = &core_itf.TemplateParams{
			Description: param.Description,
			Required:    param.Required,
			Type:        param.Type,
			Default:     param.Default,
			Options:     param.Options,
		}
	}

	saved, err := a.templates.Upsert(&core_itf.Template{
		ID:                   id,
		Name:                 template.Name,
		Role:                 template.Role,
		TaskLevel:            enums.TaskLevel(template.TaskLevel),
		Retryable:            template.Retryable,
		ManualAcceptRequired: template.ManualAcceptRequired,
		Params:               params,
		SystemPrompts:        template.SystemPrompts,
		OutputStructure:      template.OutputStructure,
	})
	if err != nil {
		return "", err
	}

	return saved.String(), nil
}

func (a *API) RemoveTemplate(id string) error {
	return withID(idTemplate, id, a.templates.Remove)
}

func (a *API) ExportTemplates(ids []string, path string) (int, error) {
	parsed := make([]uuid.UUID, 0, len(ids))

	for _, id := range ids {
		templateID, err := parseID(idTemplate, id)
		if err != nil {
			return 0, err
		}

		parsed = append(parsed, templateID)
	}

	return a.templates.Export(parsed, path)
}

func (a *API) ImportTemplates(path string) (int, error) {
	return a.templates.Import(path)
}

func templateInfo(template *core_itf.Template) *output_itf.TemplateInfo {
	if template == nil {
		return nil
	}

	params := map[string]*output_itf.TemplateParamInfo{}

	for key, param := range template.Params {
		params[key] = &output_itf.TemplateParamInfo{
			Description: param.Description,
			Required:    param.Required,
			Type:        param.Type,
			Default:     param.Default,
			Options:     param.Options,
		}
	}

	return &output_itf.TemplateInfo{
		ID:                   template.ID.String(),
		Name:                 template.Name,
		Role:                 template.Role,
		TaskLevel:            template.TaskLevel.String(),
		Retryable:            template.Retryable,
		ManualAcceptRequired: template.ManualAcceptRequired,
		Params:               params,
		SystemPrompts:        template.SystemPrompts,
		OutputStructure:      template.OutputStructure,
	}
}

// TemplateHelperBlocked is empty when a template can be filled in, and otherwise
// says why it cannot.
func (a *API) TemplateHelperBlocked() string {
	return a.templateHelp.Blocked()
}

func (a *API) RefineTemplate(req *core_itf.DraftRequest) (*output_itf.TemplateInfo, error) {
	template, err := a.templateHelp.Draft(req)
	if err != nil {
		return nil, err
	}

	return templateInfo(template), nil
}

func (a *API) UninstallAgent(id string) error {
	h, err := a.admin(id)
	if err != nil {
		return err
	}

	return h.Uninstall()
}

func (a *API) RunSession(spec *output_itf.RunSessionSpec) (*output_itf.RunSessionResult, error) {
	if spec == nil || len(spec.Tasks) == 0 {
		return nil, custom_error.Critical("run session spec has no tasks")
	}

	sessionID, err := a.sessions.NewSession(&core_itf.InitSession{
		WorkingDirPath: spec.WorkingDirPath,
		ContextDirPath: spec.ContextDirPath,
	})
	if err != nil {
		return nil, err
	}

	autopilot := a.userConfig.Autopilot()

	clientToTask := map[string]uuid.UUID{}
	remaining := append([]*output_itf.RunTaskSpec{}, spec.Tasks...)

	for len(remaining) > 0 {
		next := make([]*output_itf.RunTaskSpec, 0, len(remaining))

		for _, task := range remaining {
			deps, resolved := resolveDeps(task.DependsOn, clientToTask)
			if !resolved {
				next = append(next, task)
				continue
			}

			taskID, err := a.addSessionTask(sessionID, task, deps, autopilot)
			if err != nil {
				return nil, err
			}

			clientToTask[task.ClientID] = taskID
		}

		if len(next) == len(remaining) {
			return nil, custom_error.Critical("task dependencies are unresolvable or cyclic")
		}

		remaining = next
	}

	if err := a.coordinator.Run(sessionID); err != nil {
		return nil, err
	}

	taskIDs := make(map[string]string, len(clientToTask))
	for clientID, taskID := range clientToTask {
		taskIDs[clientID] = taskID.String()
	}

	return &output_itf.RunSessionResult{
		SessionID: sessionID.String(),
		TaskIDs:   taskIDs,
	}, nil
}

func (a *API) ChooseDirectory(title string) (string, error) {
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:                title,
		CanCreateDirectories: true,
	})
}

func (a *API) ChooseFile(title string, pattern string) (string, error) {
	return runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   title,
		Filters: fileFilters(pattern),
	})
}

func (a *API) ChooseSaveFile(title string, defaultName string, pattern string) (string, error) {
	return runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:                title,
		DefaultFilename:      defaultName,
		Filters:              fileFilters(pattern),
		CanCreateDirectories: true,
	})
}

func fileFilters(pattern string) []runtime.FileFilter {
	if pattern == "" {
		return nil
	}

	return []runtime.FileFilter{{
		DisplayName: strings.ToUpper(strings.TrimPrefix(pattern, "*.")) + " files",
		Pattern:     pattern,
	}}
}

func (a *API) AgentDefaults() ([]*output_itf.AgentDefaultInfo, error) {
	stored := a.userConfig.AgentDefaults()

	infos := make([]*output_itf.AgentDefaultInfo, 0, len(stored))

	for _, level := range enums.TaskLevels() {
		agentDefault := stored[level]
		if agentDefault == nil {
			continue
		}

		infos = append(infos, &output_itf.AgentDefaultInfo{
			TaskLevel:     level.String(),
			Model:         agentDefault.Model.String(),
			ModelLabel:    agentDefault.Model.DisplayName(),
			ThinkingLevel: agentDefault.ThinkingLevel.String(),
		})
	}

	return infos, nil
}

func (a *API) SetAgentDefault(taskLevel string, model string, thinkingLevel string) error {
	return a.userConfig.SetAgentDefault(enums.TaskLevel(taskLevel), &output_itf.AgentDefault{
		Model:         enums.ModelName(model),
		ThinkingLevel: enums.ThinkingLevel(thinkingLevel),
	})
}

func (a *API) Onboarded() bool {
	return a.userConfig.Onboarded()
}

func (a *API) CompleteOnboarding() error {
	return a.userConfig.CompleteOnboarding()
}

func (a *API) Autopilot() bool {
	return a.userConfig.Autopilot()
}

func (a *API) SetAutopilot(on bool) error {
	return a.userConfig.SetAutopilot(on)
}

func (a *API) AgentDefaultOptions() (*output_itf.AgentDefaultOptionsInfo, error) {
	names := enums.ModelNames()
	models := make([]*output_itf.ModelOptionInfo, 0, len(names))

	for _, name := range names {
		models = append(models, &output_itf.ModelOptionInfo{
			Model:   name.String(),
			Label:   name.DisplayName(),
			Harness: name.HarnessTool().String(),
		})
	}

	return &output_itf.AgentDefaultOptionsInfo{
		TaskLevels:     helpers.Labels(enums.TaskLevels()),
		Models:         models,
		ThinkingLevels: helpers.Labels(enums.ThinkingLevels()),
	}, nil
}

func (a *API) addSessionTask(sessionID uuid.UUID, task *output_itf.RunTaskSpec, deps []uuid.UUID, autopilot bool) (uuid.UUID, error) {
	agentDefault, err := a.userConfig.AgentDefault(enums.TaskLevel(task.TaskLevel))
	if err != nil {
		return uuid.Nil, err
	}

	return a.sessions.AddTask(sessionID, &core_itf.AddTask{
		Name:                 task.Name,
		AutoRetry:            task.AutoRetry,
		ManualAcceptRequired: task.ManualAcceptRequired && !autopilot,
		ExtraGuidance:        task.Prompt,
		OutputStructure:      task.OutputStructure,
		DependsOn:            deps,
		AgentSpecs: &core_itf.AgentRequest{
			Name:          agentDefault.Model,
			ThinkingLevel: agentDefault.ThinkingLevel,
			SystemPrompts: task.SystemPrompts,
		},
	})
}

func resolveDeps(dependsOn []string, clientToTask map[string]uuid.UUID) ([]uuid.UUID, bool) {
	deps := make([]uuid.UUID, 0, len(dependsOn))

	for _, clientID := range dependsOn {
		taskID, found := clientToTask[clientID]
		if !found {
			return nil, false
		}

		deps = append(deps, taskID)
	}

	return deps, true
}

func (a *API) SessionStatus(sessionID string) (*output_itf.SessionStatusInfo, error) {
	return fromID(idSession, sessionID, func(parsed uuid.UUID) (*output_itf.SessionStatusInfo, error) {
		status, err := a.sessions.Status(parsed)
		if err != nil {
			return nil, err
		}

		return sessionStatusInfo(status), nil
	})
}

func (a *API) ResumeSession(sessionID string) error {
	return withID(idSession, sessionID, a.coordinator.Run)
}

func (a *API) CancelSession(sessionID string) error {
	return withID(idSession, sessionID, a.coordinator.Cancel)
}

func (a *API) TaskDiff(sessionID, taskID string) ([]*output_itf.FileChangeInfo, error) {
	infos := []*output_itf.FileChangeInfo{}

	err := withSessionTask(sessionID, taskID, func(session, task uuid.UUID) error {
		changes, err := a.history.Diff(session, task)
		if err != nil {
			return err
		}

		infos = fileChangeInfos(changes)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return infos, nil
}

func (a *API) RevertSessionTo(sessionID, taskID string) error {
	return withSessionTask(sessionID, taskID, a.coordinator.RevertTo)
}

func fileChangeInfos(changes []*input_itf.FileChange) []*output_itf.FileChangeInfo {
	infos := make([]*output_itf.FileChangeInfo, 0, len(changes))

	for _, change := range changes {
		infos = append(infos, &output_itf.FileChangeInfo{
			Path:        change.Path,
			OldPath:     change.OldPath,
			ChangeType:  change.ChangeType.String(),
			Additions:   change.Additions,
			Deletions:   change.Deletions,
			UnifiedDiff: change.UnifiedDiff,
		})
	}

	return infos
}

func (a *API) RetrySessionTask(taskID string) error {
	return withID(idTask, taskID, a.sessions.RetryTask)
}

func (a *API) AnswerTaskAcceptance(taskID string, accepted bool) error {
	return withID(idTask, taskID, func(parsed uuid.UUID) error {
		return a.sessions.AnswerAcceptance(parsed, accepted)
	})
}

func (a *API) SessionDrafts() ([]*output_itf.SessionDraftInfo, error) {
	stored, err := a.drafts.List()
	if err != nil {
		return nil, err
	}

	infos := make([]*output_itf.SessionDraftInfo, 0, len(stored))

	for _, draft := range stored {
		infos = append(infos, &output_itf.SessionDraftInfo{
			ID:        draft.ID.String(),
			Doc:       draft.Doc,
			UpdatedAt: draft.UpdatedAt.Format(time.RFC3339),
		})
	}

	return infos, nil
}

func (a *API) SaveSessionDraft(id string, doc string) error {
	return withID(idDraft, id, func(parsed uuid.UUID) error {
		return a.drafts.Save(&input_itf.SessionDraftEntity{
			ID:        parsed,
			Doc:       doc,
			UpdatedAt: helpers.NewUTC(),
		})
	})
}

func (a *API) DeleteSessionDraft(id string) error {
	return withID(idDraft, id, a.drafts.Delete)
}

func (a *API) MCPServers() ([]*output_itf.MCPServerInfo, error) {
	servers, err := a.mcpProxy.List()
	if err != nil {
		return nil, err
	}

	infos := make([]*output_itf.MCPServerInfo, 0, len(servers))

	for _, server := range servers {
		authorizedAt := ""
		if !server.InitializedAt.IsZero() {
			authorizedAt = server.InitializedAt.Format(time.RFC3339)
		}

		infos = append(infos, &output_itf.MCPServerInfo{
			Name:         server.ServerName,
			URL:          server.URL,
			Authorized:   server.Authenticated,
			AuthorizedAt: authorizedAt,
			Account:      server.Account,
			Kind:         server.Kind.String(),
		})
	}

	sort.Slice(infos, func(i, j int) bool { return infos[i].Name < infos[j].Name })

	return infos, nil
}

func (a *API) AuthorizeMCPServer(name string) error {
	return a.mcpProxy.Authorize(name)
}

func (a *API) SetMCPCredential(name, secret string) error {
	return a.mcpProxy.SetCredential(name, secret)
}

func (a *API) RevokeMCPServer(name string) error {
	return a.mcpProxy.Revoke(name)
}

func sessionStatusInfo(status *core_itf.SessionStatus) *output_itf.SessionStatusInfo {
	if status == nil {
		return nil
	}

	tasks := make([]*output_itf.SessionTaskInfo, 0, len(status.Tasks))

	for taskID, report := range status.Tasks {
		tasks = append(tasks, sessionTaskInfo(taskID, report))
	}

	sort.Slice(tasks, func(i, j int) bool { return tasks[i].TaskID < tasks[j].TaskID })

	return &output_itf.SessionStatusInfo{
		SessionID: status.ID.String(),
		Status:    string(status.Status),
		Tasks:     tasks,
	}
}

func sessionTaskInfo(taskID uuid.UUID, report *core_itf.TaskReport) *output_itf.SessionTaskInfo {
	info := &output_itf.SessionTaskInfo{
		TaskID:       taskID.String(),
		HandoverDocs: []*output_itf.HandoverDocInfo{},
		Activity:     []*output_itf.TaskActivityInfo{},
	}

	if report == nil {
		return info
	}

	info.Status = string(report.Status)
	info.ContextUsage = report.ContextUsage

	if report.AgentID != uuid.Nil {
		info.AgentID = report.AgentID.String()
	}

	for _, doc := range report.HandoverDocs {
		info.HandoverDocs = append(info.HandoverDocs, handoverDocInfo(doc))
	}

	for _, line := range report.Activity {
		info.Activity = append(info.Activity, &output_itf.TaskActivityInfo{
			Seq:  line.Seq,
			At:   line.At.Format(time.RFC3339),
			Text: line.Text,
		})
	}

	return info
}

func handoverDocInfo(doc *core_itf.HandoverDoc) *output_itf.HandoverDocInfo {
	if doc == nil {
		return nil
	}

	return &output_itf.HandoverDocInfo{
		Task:              doc.Task,
		TLDR:              doc.TLDR,
		Outcome:           doc.Outcome,
		Blockers:          doc.Blockers,
		ApprovedDecisions: doc.ApprovedDecisions,
		RejectedDecisions: doc.RejectedDecisions,
		CurrentBehaviors:  doc.CurrentBehaviors,
		ChangedBehaviors:  doc.ChangedBehaviors,
		MustAvoid:         doc.MustAvoid,
		Nuances:           doc.Nuances,
		KnownGaps:         doc.KnownGaps,
	}
}
