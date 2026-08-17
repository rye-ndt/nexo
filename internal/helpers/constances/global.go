package constances

const GlobalLocalHost = "127.0.0.1"

const (
	GatewayAgentHeader      = "X-Harness-Agent-Id"
	GatewayAgentPlaceholder = "__HARNESS_AGENT_ID__"
	GatewayLocalServer      = "harness"
	FigmaLocalServer        = "figma"
	ChromeLocalServer       = "chrome-devtools"
	GatewayMCPPath          = "/mcp/"
)

const (
	PKCEMinVerifierBytes         = 32
	PKCEMinStateBytes            = 16
	PKCESupportedChallengeMethod = "S256"
)
