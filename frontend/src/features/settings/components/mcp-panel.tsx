import {useState, type FormEvent} from 'react'

import {Button} from '@/shared/ui/button'
import {ConfirmDialog} from '@/shared/components/confirm-dialog'
import {Input} from '@/shared/ui/input'
import {HelpTip} from '@/shared/components/help-tip'
import {MCPAuthKind} from '@/shared/lib/enums'
import {useMCPServers} from '@/features/settings/use-mcp'
import {useToggle} from '@/shared/hooks/use-toggle'
import {formatRelative} from '@/shared/lib/format'
import {t, type MessageKey} from '@/shared/lib/i18n'
import type {MCPServer} from '@/features/settings/types'

type KindCopy = {
    connect: MessageKey
    connecting: MessageKey
    disconnect: MessageKey
    disconnecting: MessageKey
    destructive: boolean
    title: MessageKey
    describe: (server: MCPServer) => string
}

function revokable(withAccount: MessageKey, withoutAccount: MessageKey): KindCopy {
    return {
        connect: 'settings.mcp.authorize',
        connecting: 'settings.mcp.authorizing',
        disconnect: 'settings.mcp.revoke',
        disconnecting: 'settings.mcp.revoking',
        destructive: true,
        title: 'settings.mcp.revokeTitle',
        describe: (server) =>
            server.account
                ? t(withAccount, {name: server.name, account: server.account})
                : t(withoutAccount, {name: server.name}),
    }
}

const KIND_COPY: Record<MCPAuthKind, KindCopy> = {
    [MCPAuthKind.DynamicRegistration]: revokable(
        'settings.mcp.revokeAuthAccount',
        'settings.mcp.revokeAuth',
    ),
    [MCPAuthKind.Device]: revokable('settings.mcp.revokeAuthAccount', 'settings.mcp.revokeAuth'),
    [MCPAuthKind.Token]: {
        ...revokable('settings.mcp.revokeTokenAccount', 'settings.mcp.revokeToken'),
        connect: 'settings.mcp.saveToken',
        connecting: 'settings.mcp.savingToken',
    },
    [MCPAuthKind.Enable]: {
        connect: 'settings.mcp.enable',
        connecting: 'settings.mcp.enabling',
        disconnect: 'settings.mcp.disable',
        disconnecting: 'settings.mcp.disabling',
        destructive: false,
        title: 'settings.mcp.disableTitle',
        describe: () => t('settings.mcp.disableDescription'),
    },
}

export function MCPPanel() {
    const {servers, loading, pendingId, authorize, setToken, savingId, revoke, revokingId} =
        useMCPServers()

    const isEmpty = !loading && servers.length === 0

    return (
        <section className="flex flex-col">
            <div className="flex flex-col gap-1 px-4 pt-4 pb-3">
                <div className="flex items-center gap-2">
                    <h3 className="text-lg font-medium">{t('settings.mcp.title')}</h3>
                    <HelpTip term="mcp" />
                </div>
                <p className="text-sm text-muted-foreground">{t('settings.mcp.hint')}</p>
            </div>

            <div className="divide-y divide-border border-t border-border">
                {loading && (
                    <p className="px-4 py-3 text-base text-muted-foreground">
                        {t('settings.mcp.loading')}
                    </p>
                )}

                {!loading &&
                    servers.map((server) => (
                        <MCPServerRow
                            key={server.id}
                            server={server}
                            pending={pendingId === server.id || savingId === server.id}
                            revoking={revokingId === server.id}
                            onAuthorize={authorize}
                            onSetToken={setToken}
                            onRevoke={revoke}
                        />
                    ))}

                {isEmpty && (
                    <p className="px-4 py-3 text-base text-muted-foreground">
                        {t('settings.mcp.empty')}
                    </p>
                )}
            </div>
        </section>
    )
}

function MCPServerRow({
    server,
    pending,
    revoking,
    onAuthorize,
    onSetToken,
    onRevoke,
}: {
    server: MCPServer
    pending: boolean
    revoking: boolean
    onAuthorize: (serverId: string) => void
    onSetToken: (input: {serverId: string; token: string}) => void
    onRevoke: (serverId: string) => void
}) {
    const confirmingRevoke = useToggle()
    const copy = KIND_COPY[server.kind]
    const busy = pending || revoking

    const revoke = () => {
        onRevoke(server.id)
        confirmingRevoke.close()
    }

    return (
        <div className="flex items-center gap-3 px-4 py-3">
            <div className="flex min-w-0 flex-1 flex-col gap-1">
                <p className="truncate text-base font-medium">{server.name}</p>
                <p className="flex min-w-0 items-center gap-1.5 text-sm text-muted-foreground">
                    {server.account && (
                        <>
                            <span className="max-w-56 shrink-0 truncate">{server.account}</span>
                            <span aria-hidden>·</span>
                        </>
                    )}
                    <span className="truncate font-mono">{server.url}</span>
                </p>
            </div>

            {server.authorized ? (
                <span className="flex shrink-0 items-center gap-2">
                    {server.authorizedAt && (
                        <span className="text-sm text-muted-foreground">
                            {formatRelative(server.authorizedAt)}
                        </span>
                    )}
                    <Button
                        variant="ghost"
                        size="sm"
                        className={
                            copy.destructive
                                ? 'text-muted-foreground hover:text-destructive'
                                : 'text-muted-foreground'
                        }
                        disabled={busy}
                        onClick={confirmingRevoke.open}
                    >
                        {t(revoking ? copy.disconnecting : copy.disconnect)}
                    </Button>

                    {confirmingRevoke.on && (
                        <ConfirmDialog
                            destructive={copy.destructive}
                            title={t(copy.title, {name: server.name})}
                            description={copy.describe(server)}
                            confirmLabel={t(copy.disconnect)}
                            onConfirm={revoke}
                            onClose={confirmingRevoke.close}
                        />
                    )}
                </span>
            ) : (
                <MCPConnectAction
                    server={server}
                    busy={busy}
                    pending={pending}
                    onAuthorize={onAuthorize}
                    onSetToken={onSetToken}
                />
            )}
        </div>
    )
}

function MCPConnectAction({
    server,
    busy,
    pending,
    onAuthorize,
    onSetToken,
}: {
    server: MCPServer
    busy: boolean
    pending: boolean
    onAuthorize: (serverId: string) => void
    onSetToken: (input: {serverId: string; token: string}) => void
}) {
    const copy = KIND_COPY[server.kind]
    const label = t(pending ? copy.connecting : copy.connect)

    if (server.kind === MCPAuthKind.Token)
        return (
            <MCPTokenForm
                busy={busy}
                label={label}
                onSubmit={(token) => onSetToken({serverId: server.id, token})}
            />
        )

    return (
        <Button
            size="sm"
            className="shrink-0"
            disabled={busy}
            onClick={() => onAuthorize(server.id)}
        >
            {label}
        </Button>
    )
}

function MCPTokenForm({
    busy,
    label,
    onSubmit,
}: {
    busy: boolean
    label: string
    onSubmit: (token: string) => void
}) {
    const [token, setToken] = useState('')
    const trimmed = token.trim()

    const submit = (event: FormEvent) => {
        event.preventDefault()
        if (!trimmed) return
        onSubmit(trimmed)
        setToken('')
    }

    return (
        <form className="flex shrink-0 items-center gap-2" onSubmit={submit}>
            <Input
                type="password"
                className="h-8 w-48"
                placeholder={t('settings.mcp.tokenPlaceholder')}
                value={token}
                autoComplete="off"
                disabled={busy}
                onChange={(event) => setToken(event.target.value)}
            />
            <Button size="sm" type="submit" disabled={busy || !trimmed}>
                {label}
            </Button>
        </form>
    )
}
