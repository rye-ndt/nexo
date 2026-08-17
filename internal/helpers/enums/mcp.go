package enums

import "slices"

type MCPAuthFlow string

const (
	// dynamic client registration, then authorization code against a loopback redirect
	MCPAuthFlowDCR MCPAuthFlow = "dcr"
	// a code the operator types into the server's own page
	MCPAuthFlowDevice MCPAuthFlow = "device"
	// no flow: the operator pastes a token in settings
	MCPAuthFlowToken MCPAuthFlow = "token"
	// no credential at all: the operator switches a local capability on
	MCPAuthFlowEnable MCPAuthFlow = "enable"
)

var mcpAuthFlows = []MCPAuthFlow{
	MCPAuthFlowDCR,
	MCPAuthFlowDevice,
	MCPAuthFlowToken,
	MCPAuthFlowEnable,
}

func (f MCPAuthFlow) Valid() bool {
	return slices.Contains(mcpAuthFlows, f)
}

func (f MCPAuthFlow) String() string {
	return string(f)
}
