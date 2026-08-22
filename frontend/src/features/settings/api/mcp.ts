/**
 * The only place the generated Wails bindings for MCP are touched. The roster
 * is declared in config.yaml, so the Go side lists servers, authorizes one and
 * revokes one — there is no add, edit or delete. Under the plain vite dev
 * server the same contract is answered from src/lib/mock-mcp.ts.
 */

import {MCPAuthKind} from '@/shared/lib/enums'
import {t} from '@/shared/lib/i18n'
import {bridge, hasWailsRuntime} from '@/shared/api/bridge'
import {
    mockAccount,
    mockAuthorizeFailure,
    mockTokenFailure,
    MOCK_MCP_SERVERS,
} from '@/features/settings/mock-mcp'
import type {MCPServer} from '@/features/settings/types'
import {
    AuthorizeMCPServer,
    MCPServers,
    RevokeMCPServer,
    SetMCPCredential,
} from '@wailsjs/go/wails_api/API'

const ROUNDTRIP_MS = 700

type MCPBackend = {
    list(): Promise<MCPServer[]>
    authorize(serverId: string): Promise<void>
    setCredential(serverId: string, token: string): Promise<void>
    revoke(serverId: string): Promise<void>
}

let servers: MCPServer[] = structuredClone(MOCK_MCP_SERVERS)

async function roundtrip() {
    await new Promise((resolve) => setTimeout(resolve, ROUNDTRIP_MS))
}

const wailsMCP: MCPBackend = {
    list: async () => {
        const infos = await bridge(MCPServers)

        return infos.map((info) => ({
            id: info.name,
            name: info.name,
            url: info.url,
            authorized: info.authorized,
            authorizedAt: info.authorized_at || undefined,
            account: info.account || undefined,
            kind: info.kind as MCPServer['kind'],
        }))
    },
    authorize: async (serverId) => {
        await bridge(() => AuthorizeMCPServer(serverId))
    },
    setCredential: async (serverId, token) => {
        await bridge(() => SetMCPCredential(serverId, token))
    },
    revoke: async (serverId) => {
        await bridge(() => RevokeMCPServer(serverId))
    },
}

function signIn(serverId: string) {
    servers = servers.map((server) =>
        server.id === serverId
            ? {
                  ...server,
                  authorized: true,
                  authorizedAt: new Date().toISOString(),
                  account: mockAccount(serverId),
              }
            : server,
    )
}

const mockMCP: MCPBackend = {
    list: async () => structuredClone(servers),
    authorize: async (serverId) => {
        if (!servers.some((server) => server.id === serverId))
            throw new Error(t('settings.error.serverGone'))

        await roundtrip()

        const failure = mockAuthorizeFailure(serverId)
        if (failure) throw failure

        signIn(serverId)
    },
    setCredential: async (serverId, token) => {
        const target = servers.find((server) => server.id === serverId)
        if (!target) throw new Error(t('settings.error.serverGone'))
        if (target.kind !== MCPAuthKind.Token)
            throw new Error(t('settings.error.serverTakesNoToken'))

        await roundtrip()

        const failure = mockTokenFailure(target.name, token)
        if (failure) throw failure

        signIn(serverId)
    },
    revoke: async (serverId) => {
        if (!servers.some((server) => server.id === serverId))
            throw new Error(t('settings.error.serverGone'))

        await roundtrip()

        servers = servers.map((server) =>
            server.id === serverId
                ? {...server, authorized: false, authorizedAt: undefined, account: undefined}
                : server,
        )
    },
}

const backend: MCPBackend = hasWailsRuntime() ? wailsMCP : mockMCP

export async function listMCPServers(): Promise<MCPServer[]> {
    return backend.list()
}

export async function authorizeMCPServer(serverId: string): Promise<void> {
    return backend.authorize(serverId)
}

export async function setMCPCredential(serverId: string, token: string): Promise<void> {
    return backend.setCredential(serverId, token)
}

export async function revokeMCPServer(serverId: string): Promise<void> {
    return backend.revoke(serverId)
}
