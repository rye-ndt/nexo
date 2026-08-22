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

type ControlPorts struct {
	Control     WorkflowControl
	Workflows   WorkflowManager
	Coordinator Coordinator
	Roles       RoleManager
}

type MCPProxyServer interface {
	// TrackRoleHelper hands the proxy the place a drafted role lands. It is
	// set after construction because the helper is built on top of the agent manager,
	// which is itself built on top of the gateway this proxy serves.
	TrackRoleHelper(helper RoleHelper)
	TrackWorkflowControl(ports *ControlPorts)
	List() ([]*MCPAuthInfo, error)
	Authorize(server string) error // rfc 8252
	Revoke(server string) error
	SetCredential(server, secret string) error
	Serve() (*MCPGateway, error)
	Close() error
}
