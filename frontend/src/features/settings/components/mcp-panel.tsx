import {useState, type FormEvent} from 'react'

import {Button} from '@/shared/ui/button'
import {ConfirmDialog} from '@/shared/components/confirm-dialog'
import {Input} from '@/shared/ui/input'
import {HelpTip} from '@/shared/components/help-tip'
import {StatusChip} from '@/shared/components/status-chip'
import {MCPAuthKind} from '@/shared/lib/enums'
import {useMCPServers} from '@/features/settings/use-mcp'
import {useToggle} from '@/shared/hooks/use-toggle'
import {formatRelative} from '@/shared/lib/format'
import type {MCPServer} from '@/features/settings/types'

type KindCopy = {
    connected: string
    connect: string
    connecting: string
    disconnect: string
    disconnecting: string
    destructive: boolean
    title: (server: MCPServer) => string
    description: (server: MCPServer) => string
}

function revokable(regain: string): KindCopy {
    return {
        connected: 'Authorized',
        connect: 'Authorize',
        connecting: 'Authorizing',
        disconnect: 'Revoke',
        disconnecting: 'Revoking',
        destructive: true,
        title: (server) => `Revoke ${server.name} access`,
        description: (server) =>
            `Agents lose access to ${server.name} ${regain}. This deletes ${
                server.account
                    ? `the credential stored for ${server.account}`
                    : 'the stored credential'
            }.`,
    }
}

const KIND_COPY: Record<MCPAuthKind, KindCopy> = {
    [MCPAuthKind.DynamicRegistration]: revokable('until you authorize it again'),
    [MCPAuthKind.Device]: revokable('until you authorize it again'),
    [MCPAuthKind.Token]: {
        ...revokable('until you save a new token'),
        connected: 'Connected',
        connect: 'Save token',
        connecting: 'Saving',
    },
    [MCPAuthKind.Enable]: {
        connected: 'Enabled',
        connect: 'Enable',
        connecting: 'Enabling',
        disconnect: 'Disable',
        disconnecting: 'Disabling',
        destructive: false,
        title: (server) => `Disable ${server.name}`,
        description: () =>
            'Nexo quits the Chrome it started, and agents stop driving your browser. You stay signed in to every site, and turning it back on takes one click.',
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
                    <h3 className="text-lg font-medium">Servers agents can call</h3>
                    <HelpTip term="mcp" />
                </div>
                <p className="text-sm text-muted-foreground">
                    The list comes from config.yaml. Authorize a server once and every agent reaches
                    it through the proxy.
                </p>
            </div>

            <div className="divide-y divide-border border-t border-border">
                {loading && (
                    <p className="px-4 py-3 text-base text-muted-foreground">
                        Loading MCP servers…
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
                        No MCP servers in config.yaml. Add one there and reopen this panel.
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
                    <StatusChip tone="done">{copy.connected}</StatusChip>
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
                        {revoking ? copy.disconnecting : copy.disconnect}
                    </Button>

                    {confirmingRevoke.on && (
                        <ConfirmDialog
                            destructive={copy.destructive}
                            title={copy.title(server)}
                            description={copy.description(server)}
                            confirmLabel={copy.disconnect}
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
    const label = pending ? copy.connecting : copy.connect

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
                placeholder="Paste access token"
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
