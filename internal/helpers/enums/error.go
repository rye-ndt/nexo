package enums

type Severity string

const (
	Critical Severity = "critical"
	Bypass   Severity = "bypass"
)

func (c Severity) String() string {
	return string(c)
}

type ErrorType string

const (
	ErrCannotGetAuthInfo     ErrorType = "err_cannot_get_auth_info"
	ErrMcpNotFound           ErrorType = "err_mcp_not_found"
	ErrMcpDiscoveryFailed    ErrorType = "err_mcp_discovery_failed"
	ErrMcpRegistrationFailed ErrorType = "err_mcp_registration_failed"
	ErrMcpAuthorizeFailed    ErrorType = "err_mcp_authorize_failed"
	ErrMcpAuthorizeTimeout   ErrorType = "err_mcp_authorize_timeout"
	ErrMcpTokenExchange      ErrorType = "err_mcp_token_exchange"
	ErrMcpStoreCredentials   ErrorType = "err_mcp_store_credentials"
	ErrMcpNotAuthenticated   ErrorType = "err_mcp_not_authenticated"
	ErrMcpCredentialsExpired ErrorType = "err_mcp_credentials_expired"
	ErrMcpForbiddenRequest   ErrorType = "err_mcp_forbidden_request"
	ErrMcpRequestFailed      ErrorType = "err_mcp_request_failed"
	ErrMcpTokenRequired      ErrorType = "err_mcp_token_required"

	ErrChromeNotFound     ErrorType = "err_chrome_not_found"
	ErrChromeLaunchFailed ErrorType = "err_chrome_launch_failed"
	ErrChromeNotConnected ErrorType = "err_chrome_not_connected"

	ErrTemplateFileInvalid ErrorType = "err_template_file_invalid"
	ErrTemplateConflict    ErrorType = "err_template_conflict"

	ErrSessionFileInvalid ErrorType = "err_session_file_invalid"
)
