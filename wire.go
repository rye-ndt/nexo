package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"hexago/internal/helpers/custom_error"
	"hexago/internal/helpers/enums"
	"hexago/internal/implementation/core/agent_manager"
	"hexago/internal/implementation/core/coordinator"
	"hexago/internal/implementation/core/manual_approval_broker"
	"hexago/internal/implementation/core/mcp_proxy"
	"hexago/internal/implementation/core/session_manager"
	"hexago/internal/implementation/core/template_manager"
	"hexago/internal/implementation/core/wal_sync"
	"hexago/internal/implementation/input/config"
	"hexago/internal/implementation/input/harness/claude_code"
	"hexago/internal/implementation/input/harness/open_code"
	"hexago/internal/implementation/input/http_cli"
	"hexago/internal/implementation/input/storage"
	"hexago/internal/implementation/input/wal"
	"hexago/internal/implementation/input/workspace_history"
	"hexago/internal/implementation/output/app_builder"
	wails_api "hexago/internal/implementation/output/fe_api"
	"hexago/internal/implementation/output/logger"
	"hexago/internal/implementation/output/message_queue"
	"hexago/internal/implementation/output/user_config"
	core_itf "hexago/internal/interface/core"
	input_itf "hexago/internal/interface/input"
	output_itf "hexago/internal/interface/output"
)

const httpTimeout = 30 * time.Second

const dataCorruptedWarning = "Task history from the previous session is corrupted or could not be saved. " +
	"The app will continue without it."

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
	enums.OpenCode:   open_code.New,
}

func wire(assets fs.FS) (*App, error) {
	cfg, err := config.New("config.yaml")
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

	taskWAL, err := wal.New(filepath.Join(dataDir, "task.wal"))
	if err != nil {
		return nil, err
	}

	history, err := workspace_history.InitV1(filepath.Join(dataDir, "sessions"))
	if err != nil {
		return nil, err
	}

	userCfg, err := user_config.InitV1(filepath.Join(dataDir, "user_config.json"))
	if err != nil {
		return nil, err
	}

	dataWarning := ""
	if err := wal_sync.Run(taskWAL, store.TaskStore()); err != nil {
		appLogger.Error("wal sync", "err", err)
		dataWarning = dataCorruptedWarning
	}

	sessionManager, err := session_manager.InitV1(cfg.Read().Session, taskWAL, message_queue.InitV1())
	if err != nil {
		return nil, err
	}

	approvalBroker, err := manual_approval_broker.InitV1(cfg.Read().ApprovalBroker)
	if err != nil {
		return nil, err
	}

	mcpProxy, err := mcp_proxy.InitV1(cfg.Read().MCPServers, store.MCPStore(), httpCli, approvalBroker, sessionManager)
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

	agentManager, err := agent_manager.InitV1(cfg.Read().AgentManager, httpCli, harnesses, approvalBroker)
	if err != nil {
		return nil, err
	}

	sessionManager.TrackLiveAgents(agentManager)

	sessionCoordinator, err := coordinator.InitV1(cfg.Read().Session, sessionManager, agentManager, history, appLogger)
	if err != nil {
		return nil, err
	}

	templateManager, err := template_manager.InitV1(store.TemplateStore())
	if err != nil {
		return nil, err
	}

	feAPI := wails_api.New(&wails_api.Deps{
		AgentManager: agentManager,
		MCPProxy:     mcpProxy,
		Approvals:    approvalBroker,
		Templates:    templateManager,
		Sessions:     sessionManager,
		Coordinator:  sessionCoordinator,
		History:      history,
		UserConfig:   userCfg,
		Drafts:       store.DraftStore(),
		DataWarning:  dataWarning,
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
