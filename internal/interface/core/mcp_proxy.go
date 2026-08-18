package core_itf

import (
	"time"

	"hexago/internal/helpers/enums"
)

type MCPAuthInfo struct {
	ServerName    string
	Kind          enums.MCPAuthFlow
	URL           string
	Authenticated bool
	Account       string
	InitializedAt time.Time
}

type MCPGatewayServer struct {
	Name        string
	AuthKeyName string
}

type MCPGateway struct {
	BaseURL     string
	Token       string
	TokenHeader string
	Servers     []MCPGatewayServer
}

type MCPProxyServer interface {
	// TrackTemplateHelper hands the proxy the place a drafted template lands. It is
	// set after construction because the helper is built on top of the agent manager,
	// which is itself built on top of the gateway this proxy serves.
	TrackTemplateHelper(helper TemplateHelper)
	TrackSessionControl(control SessionControl)
	List() ([]*MCPAuthInfo, error)
	Authorize(server string) error // rfc 8252
	Revoke(server string) error
	SetCredential(server, secret string) error
	Serve() (*MCPGateway, error)
	Close() error
}
