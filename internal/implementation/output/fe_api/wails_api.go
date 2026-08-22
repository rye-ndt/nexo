package wails_api

import (
	"context"
	"encoding/json"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/google/uuid"

	"hexago/internal/helpers"
	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	"hexago/internal/helpers/pricing"
	core_itf "hexago/internal/interface/core"
	input_itf "hexago/internal/interface/input"
	output_itf "hexago/internal/interface/output"
)

const installProgressEvent = "harness:install:progress"

const (
	idApproval = "approval"
	idRole     = "role"
	idWorkflow = "workflow"
	idStep     = "step"
	idDraft    = "workflow draft"
)

type API struct {
	ctx             context.Context
	agentManager    core_itf.AgentManager
	mcpProxy        core_itf.MCPProxyServer
	approvals       core_itf.ApprovalBroker
	roles           core_itf.RoleManager
	workflows       core_itf.WorkflowManager
	control         core_itf.WorkflowControl
	coordinator     core_itf.Coordinator
	history         input_itf.WorkspaceHistory
	userConfig      output_itf.UserConfig
	drafts          input_itf.DraftStorage
	workflowArchive input_itf.WorkflowArchive
	roleHelp        core_itf.RoleHelper
}

var _ output_itf.FEAPI = (*API)(nil)

type Deps struct {
	AgentManager    core_itf.AgentManager
	MCPProxy        core_itf.MCPProxyServer
	Approvals       core_itf.ApprovalBroker
	Roles           core_itf.RoleManager
	Workflows       core_itf.WorkflowManager
	Control         core_itf.WorkflowControl
	Coordinator     core_itf.Coordinator
	History         input_itf.WorkspaceHistory
	UserConfig      output_itf.UserConfig
	Drafts          input_itf.DraftStorage
	WorkflowArchive input_itf.WorkflowArchive
	RoleHelper      core_itf.RoleHelper
}

func New(deps *Deps) *API {
	return &API{
		agentManager:    deps.AgentManager,
		mcpProxy:        deps.MCPProxy,
		approvals:       deps.Approvals,
		roles:           deps.Roles,
		workflows:       deps.Workflows,
		control:         deps.Control,
		coordinator:     deps.Coordinator,
		history:         deps.History,
		userConfig:      deps.UserConfig,
		drafts:          deps.Drafts,
		workflowArchive: deps.WorkflowArchive,
		roleHelp:        deps.RoleHelper,
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

	if a.workflows != nil {
		a.workflows.Stop()
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

func withWorkflowStep(workflowID, stepID string, do func(workflow, step uuid.UUID) error) error {
	return withID(idWorkflow, workflowID, func(workflow uuid.UUID) error {
		return withID(idStep, stepID, func(step uuid.UUID) error {
			return do(workflow, step)
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

func (a *API) LogoutAgent(id string) error {
	h, err := a.admin(id)
	if err != nil {
		return err
	}

	return h.Logout()
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
		StepID:      request.StepID.String(),
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

func (a *API) Roles() ([]*output_itf.RoleInfo, error) {
	stored, err := a.roles.List()
	if err != nil {
		return nil, err
	}

	infos := make([]*output_itf.RoleInfo, 0, len(stored))

	for _, role := range stored {
		infos = append(infos, roleInfo(role))
	}

	return infos, nil
}

func (a *API) Role(id string) (*output_itf.RoleInfo, error) {
	return fromID(idRole, id, func(parsed uuid.UUID) (*output_itf.RoleInfo, error) {
		role, err := a.roles.Get(parsed)
		if err != nil {
			return nil, err
		}

		return roleInfo(role), nil
	})
}

func (a *API) UpsertRole(role *output_itf.RoleInfo) (string, error) {
	if role == nil {
		return "", custom_error.Critical("role is empty")
	}

	id := uuid.Nil

	if role.ID != "" {
		parsed, err := parseID(idRole, role.ID)
		if err != nil {
			return "", err
		}

		id = parsed
	}

	inputs := map[string]*core_itf.RoleInputs{}

	for key, input := range role.Inputs {
		if input == nil {
			inputs[key] = nil
			continue
		}

		inputs[key] = &core_itf.RoleInputs{
			Description: input.Description,
			Required:    input.Required,
			Type:        input.Type,
			Default:     input.Default,
			Options:     input.Options,
		}
	}

	saved, err := a.roles.Upsert(&core_itf.Role{
		ID:              id,
		Name:            role.Name,
		Description:     role.Description,
		Effort:          enums.Effort(role.Effort),
		Retryable:       role.Retryable,
		PauseForReview:  role.PauseForReview,
		Inputs:          inputs,
		Instructions:    role.Instructions,
		OutputStructure: role.OutputStructure,
	})
	if err != nil {
		return "", err
	}

	return saved.String(), nil
}

func (a *API) RemoveRole(id string) error {
	return withID(idRole, id, a.roles.Remove)
}

func (a *API) ExportRoles(ids []string, path string) (int, error) {
	parsed := make([]uuid.UUID, 0, len(ids))

	for _, id := range ids {
		roleID, err := parseID(idRole, id)
		if err != nil {
			return 0, err
		}

		parsed = append(parsed, roleID)
	}

	return a.roles.Export(parsed, path)
}

func (a *API) ImportRoles(path string) (int, error) {
	return a.roles.Import(path)
}

func roleInfo(role *core_itf.Role) *output_itf.RoleInfo {
	if role == nil {
		return nil
	}

	inputs := map[string]*output_itf.RoleInputInfo{}

	for key, input := range role.Inputs {
		inputs[key] = &output_itf.RoleInputInfo{
			Description: input.Description,
			Required:    input.Required,
			Type:        input.Type,
			Default:     input.Default,
			Options:     input.Options,
		}
	}

	return &output_itf.RoleInfo{
		ID:              role.ID.String(),
		Name:            role.Name,
		Description:     role.Description,
		Effort:          role.Effort.String(),
		Retryable:       role.Retryable,
		PauseForReview:  role.PauseForReview,
		Inputs:          inputs,
		Instructions:    role.Instructions,
		OutputStructure: role.OutputStructure,
	}
}

// RoleHelperBlocked is empty when a role can be filled in, and otherwise
// says why it cannot.
func (a *API) RoleHelperBlocked() string {
	return a.roleHelp.Blocked()
}

func (a *API) RefineRole(req *core_itf.DraftRequest) (*output_itf.RoleInfo, error) {
	role, err := a.roleHelp.Draft(req)
	if err != nil {
		return nil, err
	}

	return roleInfo(role), nil
}

func (a *API) UninstallAgent(id string) error {
	h, err := a.admin(id)
	if err != nil {
		return err
	}

	return h.Uninstall()
}

func (a *API) RunWorkflow(spec *output_itf.RunWorkflowSpec) (*output_itf.RunWorkflowResult, error) {
	if spec == nil || len(spec.Steps) == 0 {
		return nil, custom_error.Critical("run workflow spec has no steps")
	}

	autostart := true
	plan := &core_itf.ControlWorkflowSpec{
		ProjectDirPath: spec.ProjectDirPath,
		Autostart:      &autostart,
		Steps:          make([]*core_itf.ControlStepSpec, 0, len(spec.Steps)),
	}

	for _, step := range spec.Steps {
		if step == nil {
			return nil, custom_error.Critical("run workflow spec has an empty step")
		}

		plan.Steps = append(plan.Steps, &core_itf.ControlStepSpec{
			ClientID:        step.ClientID,
			Name:            step.Name,
			Prompt:          step.Prompt,
			Effort:          step.Effort,
			Instructions:    step.Instructions,
			OutputStructure: step.OutputStructure,
			DependsOn:       step.DependsOn,
			AutoRetry:       step.AutoRetry,
			PauseForReview:  step.PauseForReview,
		})
	}

	created, err := a.control.CreateWorkflow(plan)
	if err != nil {
		return nil, err
	}

	stepIDs := make(map[string]string, len(created.StepIDs))
	for clientID, stepID := range created.StepIDs {
		stepIDs[clientID] = stepID.String()
	}

	return &output_itf.RunWorkflowResult{
		WorkflowID: created.WorkflowID.String(),
		StepIDs:    stepIDs,
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

func (a *API) ExportWorkflow(path string, doc string) error {
	return a.workflowArchive.Write(path, &input_itf.WorkflowExport{
		Version:    input_itf.ArchiveVersion,
		ExportedAt: helpers.NewUTC(),
		Workflow:   json.RawMessage(doc),
	})
}

func (a *API) ImportWorkflow(path string) (string, error) {
	read, err := a.workflowArchive.Read(path)
	if err != nil {
		return "", err
	}

	return string(read.Workflow), nil
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

	for _, level := range enums.Efforts() {
		agentDefault := stored[level]
		if agentDefault == nil {
			continue
		}

		infos = append(infos, &output_itf.AgentDefaultInfo{
			Effort:        level.String(),
			Model:         agentDefault.Model.String(),
			ModelLabel:    agentDefault.Model.DisplayName(),
			ThinkingLevel: agentDefault.ThinkingLevel.String(),
		})
	}

	return infos, nil
}

func (a *API) ModelPrices() ([]*output_itf.ModelPriceInfo, error) {
	models := enums.ModelNames()
	infos := make([]*output_itf.ModelPriceInfo, 0, len(models))

	for _, model := range models {
		prices := a.userConfig.ModelPrice(model)
		if prices == nil {
			prices = &output_itf.TokenPrices{}
		}

		infos = append(infos, &output_itf.ModelPriceInfo{
			Model:            model.String(),
			ModelLabel:       model.DisplayName(),
			InputPrice:       priceText(prices.Input),
			CachedInputPrice: priceText(prices.CachedInput),
			OutputPrice:      priceText(prices.Output),
		})
	}

	return infos, nil
}

func priceText(price *float64) string {
	if price == nil {
		return ""
	}

	return strconv.FormatFloat(*price, 'f', -1, 64)
}

func parsePrice(raw string) (*float64, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	price, err := strconv.ParseFloat(trimmed, 64)
	if err != nil || price < 0 || math.IsNaN(price) || math.IsInf(price, 0) {
		return nil, custom_error.Critical("%q is not a price", raw)
	}

	return &price, nil
}

func (a *API) SetModelPrices(model string, input string, cachedInput string, output string) error {
	inputPrice, err := parsePrice(input)
	if err != nil {
		return err
	}

	cachedPrice, err := parsePrice(cachedInput)
	if err != nil {
		return err
	}

	outputPrice, err := parsePrice(output)
	if err != nil {
		return err
	}

	if inputPrice == nil && cachedPrice == nil && outputPrice == nil {
		return a.userConfig.SetModelPrices(enums.ModelName(model), nil)
	}

	return a.userConfig.SetModelPrices(enums.ModelName(model), &output_itf.TokenPrices{
		Input:       inputPrice,
		CachedInput: cachedPrice,
		Output:      outputPrice,
	})
}

func (a *API) SetAgentDefault(effort string, model string, thinkingLevel string) error {
	return a.userConfig.SetAgentDefault(enums.Effort(effort), &output_itf.AgentDefault{
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

func (a *API) Language() string {
	return a.userConfig.Language().String()
}

func (a *API) SetLanguage(language string) error {
	return a.userConfig.SetLanguage(enums.Language(language))
}

func (a *API) Autopilot() bool {
	return a.userConfig.Autopilot()
}

func (a *API) SetAutopilot(on bool) error {
	return a.userConfig.SetAutopilot(on)
}

func (a *API) MaxRunningAgents() int {
	return a.userConfig.MaxRunningAgents()
}

func (a *API) SetMaxRunningAgents(limit int) error {
	return a.userConfig.SetMaxRunningAgents(limit)
}

func (a *API) AgentDefaultOptions() (*output_itf.AgentDefaultOptionsInfo, error) {
	agents, err := a.agentManager.SupportedAgents()
	if err != nil {
		return nil, err
	}

	harnessOf := map[enums.ModelName]enums.AgentHarness{}

	for agent, names := range agents {
		h, err := a.agentManager.Admin(agent)
		if err != nil {
			return nil, err
		}

		status, err := h.Status()
		if err != nil {
			return nil, err
		}

		if !status.LoggedIn {
			continue
		}

		for _, name := range names {
			harnessOf[name] = agent
		}
	}

	models := make([]*output_itf.ModelOptionInfo, 0, len(harnessOf))

	for _, name := range enums.ModelNames() {
		agent, ready := harnessOf[name]
		if !ready {
			continue
		}

		models = append(models, &output_itf.ModelOptionInfo{
			Model:   name.String(),
			Label:   name.DisplayName(),
			Harness: agent.String(),
		})
	}

	return &output_itf.AgentDefaultOptionsInfo{
		Efforts:        helpers.Labels(enums.Efforts()),
		Models:         models,
		ThinkingLevels: helpers.Labels(enums.ThinkingLevels()),
	}, nil
}

func (a *API) WorkflowStatus(workflowID string) (*output_itf.WorkflowStatusInfo, error) {
	return fromID(idWorkflow, workflowID, func(parsed uuid.UUID) (*output_itf.WorkflowStatusInfo, error) {
		status, err := a.workflows.Status(parsed)
		if err != nil {
			return nil, err
		}

		return a.workflowStatusInfo(status), nil
	})
}

func (a *API) ResumeWorkflow(workflowID string) error {
	return withID(idWorkflow, workflowID, a.coordinator.Run)
}

func (a *API) PauseWorkflow(workflowID string) error {
	return withID(idWorkflow, workflowID, a.coordinator.Pause)
}

func (a *API) CancelWorkflow(workflowID string) error {
	return withID(idWorkflow, workflowID, a.coordinator.Cancel)
}

func (a *API) StepDiff(workflowID, stepID string) ([]*output_itf.FileChangeInfo, error) {
	infos := []*output_itf.FileChangeInfo{}

	err := withWorkflowStep(workflowID, stepID, func(workflow, step uuid.UUID) error {
		changes, err := a.history.Diff(workflow, step)
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

func (a *API) RevertWorkflowTo(workflowID, stepID string) error {
	return withWorkflowStep(workflowID, stepID, a.coordinator.RevertTo)
}

func (a *API) DiscardWorkflowRun(workflowID string) error {
	return withID(idWorkflow, workflowID, a.coordinator.DiscardRun)
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

func (a *API) RetryWorkflowStep(stepID string) error {
	return withID(idStep, stepID, a.workflows.RetryStep)
}

func (a *API) AnswerStepReview(stepID string, accepted bool) error {
	return withID(idStep, stepID, func(parsed uuid.UUID) error {
		return a.workflows.AnswerReview(parsed, accepted)
	})
}

func (a *API) WorkflowDrafts() ([]*output_itf.WorkflowDraftInfo, error) {
	stored, err := a.drafts.List()
	if err != nil {
		return nil, err
	}

	infos := make([]*output_itf.WorkflowDraftInfo, 0, len(stored))

	for _, draft := range stored {
		infos = append(infos, &output_itf.WorkflowDraftInfo{
			ID:        draft.ID.String(),
			Doc:       draft.Doc,
			UpdatedAt: draft.UpdatedAt.Format(time.RFC3339),
		})
	}

	return infos, nil
}

func (a *API) SaveWorkflowDraft(id string, doc string) error {
	return withID(idDraft, id, func(parsed uuid.UUID) error {
		return a.drafts.Save(&input_itf.WorkflowDraftEntity{
			ID:        parsed,
			Doc:       doc,
			UpdatedAt: helpers.NewUTC(),
		})
	})
}

func (a *API) DeleteWorkflowDraft(id string) error {
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

func momentInfo(at time.Time) string {
	if at.IsZero() {
		return ""
	}

	return at.Format(time.RFC3339)
}

func (a *API) workflowStatusInfo(status *core_itf.WorkflowStatus) *output_itf.WorkflowStatusInfo {
	if status == nil {
		return nil
	}

	steps := make([]*output_itf.WorkflowStepInfo, 0, len(status.Steps))
	cost := 0.0
	priced := true

	for stepID, report := range status.Steps {
		step := a.workflowStepInfo(stepID, report)
		steps = append(steps, step)

		switch {
		case step.Priced:
			cost += step.CostUSD
		case spentAnything(step.Spent):
			priced = false
		}
	}

	sort.Slice(steps, func(i, j int) bool { return steps[i].StepID < steps[j].StepID })

	return &output_itf.WorkflowStatusInfo{
		WorkflowID:   status.ID.String(),
		Status:       string(status.Status),
		Steps:        steps,
		TokensBilled: status.TokensBilled,
		TokensInput:  status.TokensInput,
		TokensCached: status.TokensCached,
		CostUSD:      cost,
		Priced:       priced,
		StartedAt:    momentInfo(status.StartedAt),
		CompletedAt:  momentInfo(status.CompletedAt),
	}
}

func spentAnything(spent *input_itf.ContextUsage) bool {
	return spent != nil && (spent.Input > 0 || spent.Cached > 0 || spent.Billed > 0)
}

func (a *API) stepCost(model enums.ModelName, spent *input_itf.ContextUsage) (float64, bool) {
	if spent == nil {
		return 0, false
	}

	prices := a.userConfig.ModelPrice(model)
	if prices == nil {
		return 0, false
	}

	rates, ok := pricing.NewRates(prices.Input, prices.CachedInput, prices.Output)
	if !ok {
		return 0, false
	}

	return rates.Cost(pricing.Tokens{Input: spent.Input, Cached: spent.Cached, Output: spent.Billed}), true
}

func (a *API) workflowStepInfo(stepID uuid.UUID, report *core_itf.StepResult) *output_itf.WorkflowStepInfo {
	info := &output_itf.WorkflowStepInfo{
		StepID:   stepID.String(),
		Handoffs: []*output_itf.HandoffInfo{},
		Activity: []*output_itf.StepActivityInfo{},
	}

	if report == nil {
		return info
	}

	info.Status = string(report.Status)
	info.ContextUsage = report.ContextUsage
	info.Effort = report.Effort.String()
	info.Spent = report.Spent
	info.CostUSD, info.Priced = a.stepCost(report.Model, report.Spent)

	if report.AgentID != uuid.Nil {
		info.AgentID = report.AgentID.String()
	}

	for _, doc := range report.Handoffs {
		info.Handoffs = append(info.Handoffs, handoffInfo(doc))
	}

	for _, line := range report.Activity {
		info.Activity = append(info.Activity, &output_itf.StepActivityInfo{
			Seq:  line.Seq,
			At:   line.At.Format(time.RFC3339),
			Text: line.Text,
		})
	}

	return info
}

func handoffInfo(doc *core_itf.Handoff) *output_itf.HandoffInfo {
	if doc == nil {
		return nil
	}

	return &output_itf.HandoffInfo{
		Step:              doc.Step,
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
