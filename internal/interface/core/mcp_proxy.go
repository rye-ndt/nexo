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
	List() ([]*MCPAuthInfo, error)
	Authorize(server string) error // rfc 8252
	Revoke(server string) error
	SetCredential(server, secret string) error
	Serve() (*MCPGateway, error)
	Close() error
}
