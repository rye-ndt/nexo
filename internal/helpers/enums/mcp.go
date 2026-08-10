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
)

var mcpAuthFlows = []MCPAuthFlow{
	MCPAuthFlowDCR,
	MCPAuthFlowDevice,
	MCPAuthFlowToken,
}

func MCPAuthFlows() []MCPAuthFlow {
	return slices.Clone(mcpAuthFlows)
}

func (f MCPAuthFlow) Valid() bool {
	return slices.Contains(mcpAuthFlows, f)
}

func (f MCPAuthFlow) String() string {
	return string(f)
}
