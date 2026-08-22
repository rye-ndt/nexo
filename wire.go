package main

import (
	_ "embed"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	"hexago/internal/implementation/core/agent_manager"
	"hexago/internal/implementation/core/coordinator"
	"hexago/internal/implementation/core/helper_agent"
	"hexago/internal/implementation/core/manual_approval_broker"
	"hexago/internal/implementation/core/mcp_proxy"
	"hexago/internal/implementation/core/role_manager"
	"hexago/internal/implementation/core/workflow_control"
	"hexago/internal/implementation/core/workflow_manager"
	"hexago/internal/implementation/input/config"
	"hexago/internal/implementation/input/harness/claude_code"
	"hexago/internal/implementation/input/harness/codex"
	"hexago/internal/implementation/input/harness/open_code"
	"hexago/internal/implementation/input/http_cli"
	"hexago/internal/implementation/input/role_archive"
	"hexago/internal/implementation/input/storage"
	"hexago/internal/implementation/input/workflow_archive"
	"hexago/internal/implementation/input/workspace_history"
	"hexago/internal/implementation/output/app_builder"
	wails_api "hexago/internal/implementation/output/fe_api"
	"hexago/internal/implementation/output/logger"
	"hexago/internal/implementation/output/user_config"
	core_itf "hexago/internal/interface/core"
	input_itf "hexago/internal/interface/input"
	output_itf "hexago/internal/interface/output"
)

const httpTimeout = 30 * time.Second

//go:embed config.yaml
var builtinConfig []byte

type App struct {
	Config     input_itf.Config
	Logger     output_itf.Logger
	AppBuilder output_itf.AppBuilder
}

type harnessBuilder func(
	cfg input_itf.Config,
	httpCli input_itf.HttpCli,
	store input_itf.HarnessStorage,
	gateway *core_itf.MCPGateway,
	raw map[string]any,
) (input_itf.AgentHarness, error)

var harnessBuilders = map[enums.AgentHarness]harnessBuilder{
	enums.ClaudeCode: claude_code.New,
	enums.Codex:      codex.New,
	enums.OpenCode:   open_code.New,
}

func wire(assets fs.FS) (*App, error) {
	cfg, err := config.New(builtinConfig)
	if err != nil {
		return nil, err
	}

	appLogger := logger.New(cfg)

	httpCli := http_cli.New(&http_cli.BasicHttpCliCfg{Timeout: httpTimeout})

	dataDir, err := appDataDir(cfg)
	if err != nil {
		return nil, err
	}

	store, err := storage.New(filepath.Join(dataDir, "harness.db"))
	if err != nil {
		return nil, err
	}

	history, err := workspace_history.InitV1(filepath.Join(dataDir, "workflows"))
	if err != nil {
		return nil, err
	}

	workflowManager, err := workflow_manager.InitV1(cfg.Read().Workflow, store.StepStore())
	if err != nil {
		return nil, err
	}

	approvalBroker, err := manual_approval_broker.InitV1(cfg.Read().ApprovalBroker)
	if err != nil {
		return nil, err
	}

	mcpProxy, err := mcp_proxy.InitV1(
		cfg.Read().MCPServers,
		dataDir,
		store.MCPStore(),
		httpCli,
		approvalBroker,
		workflowManager,
	)
	if err != nil {
		return nil, err
	}

	mcpGateway, err := mcpProxy.Serve()
	if err != nil {
		return nil, err
	}

	wired := false

	defer func() {
		if !wired {
			mcpProxy.Close()
		}
	}()

	harnesses, err := buildHarnesses(cfg, httpCli, store, mcpGateway)
	if err != nil {
		return nil, err
	}

	userCfg, err := user_config.InitV1(
		filepath.Join(dataDir, "user_config.json"),
		cfg.Read().AgentDefaults,
		runnable(harnesses),
		cfg.Read().AgentManager.MaxRunningAgents,
	)
	if err != nil {
		return nil, err
	}

	agentManager, err := agent_manager.InitV1(cfg.Read().AgentManager, httpCli, harnesses, approvalBroker, userCfg)
	if err != nil {
		return nil, err
	}

	workflowManager.TrackLiveAgents(agentManager)

	if err := workflowManager.Restore(); err != nil {
		return nil, err
	}

	workflowCoordinator, err := coordinator.InitV1(cfg.Read().Workflow, workflowManager, agentManager, history, appLogger)
	if err != nil {
		return nil, err
	}

	roleManager, err := role_manager.InitV1(store.RoleStore(), role_archive.InitV1())
	if err != nil {
		return nil, err
	}

	roleHelper, err := helper_agent.InitV1(agentManager, roleManager, userCfg, appLogger)
	if err != nil {
		return nil, err
	}

	mcpProxy.TrackRoleHelper(roleHelper)

	workflowControl, err := workflow_control.InitV1(
		cfg.Read().MCPServers.Control,
		workflowManager,
		workflowCoordinator,
		roleManager,
		userCfg,
		store.StepStore(),
	)
	if err != nil {
		return nil, err
	}

	mcpProxy.TrackWorkflowControl(&core_itf.ControlPorts{
		Control:     workflowControl,
		Workflows:   workflowManager,
		Coordinator: workflowCoordinator,
		Roles:       roleManager,
	})

	feAPI := wails_api.New(&wails_api.Deps{
		AgentManager:    agentManager,
		MCPProxy:        mcpProxy,
		Approvals:       approvalBroker,
		Roles:           roleManager,
		Workflows:       workflowManager,
		Control:         workflowControl,
		Coordinator:     workflowCoordinator,
		History:         history,
		UserConfig:      userCfg,
		Drafts:          store.DraftStore(),
		WorkflowArchive: workflow_archive.InitV1(),
		RoleHelper:      roleHelper,
	})

	wired = true

	return &App{
		Config:     cfg,
		Logger:     appLogger,
		AppBuilder: app_builder.New(cfg, feAPI, assets),
	}, nil
}

func appDataDir(cfg input_itf.Config) (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(base, cfg.Read().App.DataDir), nil
}

func buildHarnesses(
	cfg input_itf.Config,
	httpCli input_itf.HttpCli,
	store input_itf.HarnessStorage,
	gateway *core_itf.MCPGateway,
) (map[enums.AgentHarness]input_itf.AgentHarness, error) {
	harnesses := map[enums.AgentHarness]input_itf.AgentHarness{}

	for name, raw := range cfg.Read().AgentHarness {
		build, found := harnessBuilders[name]
		if !found {
			return nil, custom_error.Critical("agent harness %s has no initializer", name)
		}

		harness, err := build(cfg, httpCli, store, gateway, raw)
		if err != nil {
			return nil, err
		}

		harnesses[name] = harness
	}

	return harnesses, nil
}

func runnable(harnesses map[enums.AgentHarness]input_itf.AgentHarness) output_itf.ModelReady {
	return func(model enums.ModelName) bool {
		for _, harness := range harnesses {
			if !harness.Support(model) {
				continue
			}

			if status, err := harness.Status(); err == nil && status.LoggedIn {
				return true
			}
		}

		return false
	}
}
