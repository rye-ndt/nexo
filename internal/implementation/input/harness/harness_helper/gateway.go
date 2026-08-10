package harness_helper

import (
	"hexago/internal/helpers/constances"
	core_itf "hexago/internal/interface/core"
)

func GatewayURL(gateway *core_itf.MCPGateway, server core_itf.MCPGatewayServer) string {
	return gateway.BaseURL + constances.GatewayMCPPath + server.Name
}

func GatewayHeaders(gateway *core_itf.MCPGateway, server core_itf.MCPGatewayServer) map[string]string {
	headers := map[string]string{
		gateway.TokenHeader:           gateway.Token,
		constances.GatewayAgentHeader: constances.GatewayAgentPlaceholder,
	}

	if server.AuthKeyName != "" {
		headers["Authorization"] = "Bearer " + server.AuthKeyName
	}

	return headers
}
