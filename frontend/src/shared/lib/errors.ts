import {t, type MessageKey} from '@/shared/lib/i18n'

type ErrorSeverity = 'critical' | 'bypass' | 'unknown'

export type AppError = {
    id: string
    code: string
    severity: ErrorSeverity
    title: string
    message: string
    hint: string
    detail: string
    at: string
}

type Reading = {title: MessageKey; hint: MessageKey}

const GO_ENVELOPE = /^\[Err\](?: Type: (\S+) -)? Message: ([\s\S]*) - Critical: (critical|bypass)$/

const FALLBACK_TITLE: MessageKey = 'error.fallbackTitle'

const READINGS: Record<string, Reading> = {
    err_cannot_get_auth_info: {
        title: 'error.cannotGetAuthInfo.title',
        hint: 'error.cannotGetAuthInfo.hint',
    },
    err_mcp_not_found: {title: 'error.mcpNotFound.title', hint: 'error.mcpNotFound.hint'},
    err_mcp_discovery_failed: {
        title: 'error.mcpDiscoveryFailed.title',
        hint: 'error.mcpDiscoveryFailed.hint',
    },
    err_mcp_registration_failed: {
        title: 'error.mcpRegistrationFailed.title',
        hint: 'error.mcpRegistrationFailed.hint',
    },
    err_mcp_authorize_failed: {
        title: 'error.mcpAuthorizeFailed.title',
        hint: 'error.mcpAuthorizeFailed.hint',
    },
    err_mcp_authorize_timeout: {
        title: 'error.mcpAuthorizeTimeout.title',
        hint: 'error.mcpAuthorizeTimeout.hint',
    },
    err_mcp_token_exchange: {
        title: 'error.mcpTokenExchange.title',
        hint: 'error.mcpTokenExchange.hint',
    },
    err_mcp_store_credentials: {
        title: 'error.mcpStoreCredentials.title',
        hint: 'error.mcpStoreCredentials.hint',
    },
    err_mcp_not_authenticated: {
        title: 'error.mcpNotAuthenticated.title',
        hint: 'error.mcpNotAuthenticated.hint',
    },
    err_mcp_credentials_expired: {
        title: 'error.mcpCredentialsExpired.title',
        hint: 'error.mcpCredentialsExpired.hint',
    },
    err_mcp_forbidden_request: {
        title: 'error.mcpForbiddenRequest.title',
        hint: 'error.mcpForbiddenRequest.hint',
    },
    err_mcp_request_failed: {
        title: 'error.mcpRequestFailed.title',
        hint: 'error.mcpRequestFailed.hint',
    },
    err_role_file_invalid: {
        title: 'error.roleFileInvalid.title',
        hint: 'error.roleFileInvalid.hint',
    },
    err_workflow_file_invalid: {
        title: 'error.workflowFileInvalid.title',
        hint: 'error.workflowFileInvalid.hint',
    },
    err_role_conflict: {title: 'error.roleConflict.title', hint: 'error.roleConflict.hint'},
    err_mcp_token_required: {
        title: 'error.mcpTokenRequired.title',
        hint: 'error.mcpTokenRequired.hint',
    },
    err_chrome_not_found: {title: 'error.chromeNotFound.title', hint: 'error.chromeNotFound.hint'},
    err_chrome_launch_failed: {
        title: 'error.chromeLaunchFailed.title',
        hint: 'error.chromeLaunchFailed.hint',
    },
    err_chrome_not_connected: {
        title: 'error.chromeNotConnected.title',
        hint: 'error.chromeNotConnected.hint',
    },
}

function errorMessage(error: unknown): string {
    if (!error) return ''
    return error instanceof Error ? error.message : String(error)
}

function readable(text: string) {
    const trimmed = text.trim()
    if (!trimmed) return ''

    const opened = /^[a-z]+\b/.test(trimmed)
        ? trimmed.charAt(0).toUpperCase() + trimmed.slice(1)
        : trimmed

    return /[.!?)]$/.test(opened) ? opened : `${opened}.`
}

function framesOf(error: unknown) {
    if (!(error instanceof Error) || !error.stack) return ''

    return error.stack
        .split('\n')
        .filter((line) => line.trim().startsWith('at '))
        .join('\n')
}

export function toAppError(cause: unknown, action = ''): AppError {
    const raw = errorMessage(cause) || String(cause)
    const envelope = GO_ENVELOPE.exec(raw)

    const code = envelope?.[1] ?? ''
    const message = envelope?.[2] ?? raw
    const severity = (envelope?.[3] as ErrorSeverity) ?? 'unknown'
    const reading = READINGS[code]

    return {
        id: crypto.randomUUID(),
        code,
        severity,
        title: reading ? t(reading.title) : action || t(FALLBACK_TITLE),
        message: readable(message),
        hint: reading ? t(reading.hint) : '',
        detail: [raw, framesOf(cause)].filter(Boolean).join('\n\n'),
        at: new Date().toISOString(),
    }
}

export function errorReport(error: AppError): string {
    const rows = [
        `Title: ${error.title}`,
        `Message: ${error.message}`,
        error.code ? `Code: ${error.code}` : '',
        `Severity: ${error.severity}`,
        `Time: ${error.at}`,
    ].filter(Boolean)

    return `${rows.join('\n')}\n\n${error.detail}\n`
}
