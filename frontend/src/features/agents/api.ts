/**
 * The only place the generated Wails bindings for agents are touched. They
 * dereference window.go synchronously, so outside the Wails webview they throw
 * instead of rejecting — every call goes through `bridge` to land as a rejected
 * promise. Under the plain vite dev server there is no Go side at all, so the
 * roster and installs fall back to src/lib/mock-agents.ts and progress is
 * emitted locally instead of over the Wails event bus.
 */

import {formatAgentName} from '@/shared/lib/format'
import {bridge, hasWailsRuntime} from '@/shared/api/bridge'
import {
    mockInstallAgent,
    mockListAgents,
    mockLogoutAgent,
    mockStartLogin,
    mockSubmitAuthCode,
    mockUninstallAgent,
} from '@/features/agents/mock-agents'
import type {Agent, InstallProgress} from '@/features/agents/types'
import {
    AgentStatuses,
    AuthAgent,
    InstallAgent,
    LogoutAgent,
    SubmitAuthCode,
    UninstallAgent,
} from '@wailsjs/go/wails_api/API'
import {BrowserOpenURL, EventsOn} from '@wailsjs/runtime/runtime'

const INSTALL_PROGRESS_EVENT = 'harness:install:progress'

type ProgressListener = (agentId: string, progress: InstallProgress) => void

type AgentsBackend = {
    list(): Promise<Agent[]>
    install(agentId: string): Promise<void>
    uninstall(agentId: string): Promise<void>
    startLogin(agentId: string): Promise<string>
    logout(agentId: string): Promise<void>
    submitAuthCode(agentId: string, code: string): Promise<void>
}

const localListeners = new Set<ProgressListener>()

const wailsAgents: AgentsBackend = {
    list: async () => {
        const infos = await bridge(AgentStatuses)

        return infos.map(({id, status}) => ({
            id,
            name: status?.name || formatAgentName(id),
            version: status?.version ?? '',
            installed: status?.installed ?? false,
            loggedIn: status?.logged_in ?? false,
            instanceCount: status?.instance_count ?? 0,
        }))
    },
    install: async (agentId) => {
        await bridge(() => InstallAgent(agentId))
    },
    uninstall: async (agentId) => {
        await bridge(() => UninstallAgent(agentId))
    },
    startLogin: async (agentId) => bridge(() => AuthAgent(agentId)),
    logout: async (agentId) => {
        await bridge(() => LogoutAgent(agentId))
    },
    submitAuthCode: async (agentId, code) => {
        await bridge(() => SubmitAuthCode(agentId, code))
    },
}

const mockAgents: AgentsBackend = {
    list: mockListAgents,
    install: async (agentId) => {
        await mockInstallAgent(agentId, (progress) => {
            for (const listener of localListeners) listener(agentId, progress)
        })
    },
    uninstall: mockUninstallAgent,
    startLogin: mockStartLogin,
    logout: mockLogoutAgent,
    submitAuthCode: mockSubmitAuthCode,
}

const backend: AgentsBackend = hasWailsRuntime() ? wailsAgents : mockAgents

export async function listAgents(): Promise<Agent[]> {
    return backend.list()
}

export async function installAgent(agentId: string): Promise<void> {
    return backend.install(agentId)
}

export async function uninstallAgent(agentId: string): Promise<void> {
    return backend.uninstall(agentId)
}

export async function startAgentLogin(agentId: string): Promise<string> {
    return backend.startLogin(agentId)
}

export async function logoutAgent(agentId: string): Promise<void> {
    return backend.logout(agentId)
}

export async function submitAgentAuthCode(agentId: string, code: string): Promise<void> {
    return backend.submitAuthCode(agentId, code)
}

export function onInstallProgress(listener: ProgressListener): () => void {
    try {
        return EventsOn(INSTALL_PROGRESS_EVENT, listener)
    } catch {
        localListeners.add(listener)
        return () => localListeners.delete(listener)
    }
}

export function openExternalURL(url: string) {
    try {
        BrowserOpenURL(url)
    } catch {
        window.open(url, '_blank', 'noopener')
    }
}
