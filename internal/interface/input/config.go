package input_itf

import (
	"time"

	"hexago/internal/helpers/enums"
)

type AppConfig struct {
	Name    string `mapstructure:"name" validate:"required"`
	DataDir string `mapstructure:"data_dir" validate:"required"`
	W       int    `mapstructure:"w" validate:"gt=0"`
	H       int    `mapstructure:"h" validate:"gt=0"`
	Bg      string `mapstructure:"bg" validate:"required,hexcolor"`
}

type WorkflowConfig struct {
	HeartbeatTimeout       time.Duration `mapstructure:"heartbeat_timeout" validate:"gt=0"`
	HeartbeatScanInterval  time.Duration `mapstructure:"heartbeat_scan_interval" validate:"gt=0,ltefield=HeartbeatTimeout"`
	AgentHeartbeatInterval time.Duration `mapstructure:"agent_heartbeat_interval" validate:"gt=0,ltefield=HeartbeatTimeout"`
}

type ApprovalBrokerConfig struct {
	Timeout time.Duration `mapstructure:"timeout" validate:"gt=0"`
}

type AgentManagerConfig struct {
	FreezeTimeout             time.Duration `mapstructure:"freeze_timeout" validate:"gt=0"`
	MaxRunningAgents          int           `mapstructure:"max_running_agents" validate:"gt=0"`
	AllowedAgentContextWindow int           `mapstructure:"allowed_agent_context_window" validate:"gte=0"`
	ConnectivityProbeURL      string        `mapstructure:"connectivity_probe_url" validate:"required,http_url"`
	ConnectivityCacheTTL      time.Duration `mapstructure:"connectivity_cache_ttl" validate:"gt=0,ltefield=FreezeTimeout"`
}

type MCPAccountConfig struct {
	URL    string   `mapstructure:"url" validate:"required,http_url"`
	Header string   `mapstructure:"header" validate:"required"`
	Scheme string   `mapstructure:"scheme"`
	Fields []string `mapstructure:"fields" validate:"required,min=1"`
}

type MCPServerConfig struct {
	Name          string            `mapstructure:"name" validate:"required"`
	AuthKeyName   string            `mapstructure:"auth_key_name"`
	URL           string            `mapstructure:"url" validate:"required,http_url"`
	AuthFlow      enums.MCPAuthFlow `mapstructure:"auth_flow" validate:"required,mcp_auth_flow"`
	ClientID      string            `mapstructure:"client_id" validate:"required_if=AuthFlow device"`
	DeviceAuthURL string            `mapstructure:"device_auth_url" validate:"required_if=AuthFlow device,omitempty,http_url"`
	TokenTTL      time.Duration     `mapstructure:"token_ttl" validate:"omitempty,gt=0"`
	Account       *MCPAccountConfig `mapstructure:"account" validate:"omitempty"`
}

type MCPChromeConfig struct {
	DebugPort     int           `mapstructure:"debug_port" validate:"gt=0"`
	ProfileDir    string        `mapstructure:"profile_dir" validate:"required"`
	LaunchTimeout time.Duration `mapstructure:"launch_timeout" validate:"gt=0"`
	CallTimeout   time.Duration `mapstructure:"call_timeout" validate:"gt=0"`
	MaxPageChars  int           `mapstructure:"max_page_chars" validate:"gt=0"`
}

type MCPServersConfig struct {
	SupportedServers map[string]*MCPServerConfig `mapstructure:"supported_servers" validate:"dive,required"`
	Chrome           *MCPChromeConfig            `mapstructure:"chrome" validate:"required"`
	Control          *ControlConfig              `mapstructure:"control" validate:"required"`
	EncodeKey        string                      `mapstructure:"encode_key" validate:"required"`
	AuthTimeout      time.Duration               `mapstructure:"auth_timeout" validate:"gt=0"`
	ClientName       string                      `mapstructure:"client_name" validate:"required"`
	CallbackPath     string                      `mapstructure:"callback_path" validate:"required,startswith=/"`
	ShutdownGrace    time.Duration               `mapstructure:"shutdown_grace" validate:"gt=0"`
	VerifierBytes    int                         `mapstructure:"verifier_bytes"`
	StateBytes       int                         `mapstructure:"state_bytes"`
	DefaultTokenTTL  time.Duration               `mapstructure:"default_token_ttl" validate:"gt=0"`
	ChallengeMethod  string                      `mapstructure:"challenge_method"`
}

type ControlConfig struct {
	Enabled             bool     `mapstructure:"enabled"`
	EndpointFile        string   `mapstructure:"endpoint_file" validate:"required"`
	AllowAnyWorkspace   bool     `mapstructure:"allow_any_workspace"`
	AllowedRoots        []string `mapstructure:"allowed_roots"`
	MaxStepsPerWorkflow int      `mapstructure:"max_steps_per_workflow" validate:"gt=0"`
	MaxWorkflowsListed  int      `mapstructure:"max_workflows_listed" validate:"gt=0"`
	AutostartDefault    bool     `mapstructure:"autostart_default"`
}

type EffortDefault struct {
	Model         enums.ModelName     `mapstructure:"model" validate:"required,model_name"`
	ThinkingLevel enums.ThinkingLevel `mapstructure:"thinking_level" validate:"required,thinking_level"`
}

type HarnessDefaults struct {
	Harness enums.AgentHarness              `mapstructure:"harness" validate:"required"`
	Models  map[enums.Effort]*EffortDefault `mapstructure:"models" validate:"required,gt=0,dive,keys,effort,endkeys,required"`
}

type ConfigStruct struct {
	App            *AppConfig                            `mapstructure:"app" validate:"required"`
	Version        string                                `mapstructure:"version" validate:"required"`
	LogLevel       string                                `mapstructure:"log_level" validate:"required,oneof=debug info warn error"`
	Workflow       *WorkflowConfig                       `mapstructure:"workflow" validate:"required"`
	ApprovalBroker *ApprovalBrokerConfig                 `mapstructure:"approval_broker" validate:"required"`
	AgentManager   *AgentManagerConfig                   `mapstructure:"agent_manager" validate:"required"`
	MCPServers     *MCPServersConfig                     `mapstructure:"mcp_servers" validate:"required"`
	AgentHarness   map[enums.AgentHarness]map[string]any `mapstructure:"agent_harness" validate:"required,gt=0"`
	AgentDefaults  []*HarnessDefaults                    `mapstructure:"agent_defaults" validate:"required,gt=0,dive,required"`
}

type Config interface {
	Read() *ConfigStruct
}
