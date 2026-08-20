/**
 * Browser stand-in for the agent bindings. Outside the Wails webview there is
 * no Go side to install anything, so the roster lives here and installs walk
 * the same stages the real harness emits. Three dev-URL flags drive the paths a
 * fresh machine cannot otherwise reach: ?installed=1 starts with every agent
 * present so the welcome modal stays closed, ?loggedIn=1 starts them
 * authorized so the workflow gate opens, and ?installFail=open_code fails that
 * agent's first attempt so the retry path is reachable. Each also takes a
 * comma-separated list of agent ids, so ?loggedIn=codex authorizes that one
 * alone.
 */

import {InstallStage} from '@/shared/lib/enums'
import {t} from '@/shared/lib/i18n'
import type {Agent, InstallProgress} from '@/features/agents/types'

const RESOLVE_MS = 400
const DOWNLOAD_MS = 2600
const DOWNLOAD_TICKS = 26
const EXTRACT_MS = 600
const APPROVE_MS = 3500
const MIN_CODE_LENGTH = 8

const MOCK_AGENTS: Agent[] = [
    {
        id: 'claude_code',
        name: 'Claude Code',
        version: '1.0.42',
        installed: false,
        loggedIn: false,
        instanceCount: 0,
    },
    {
        id: 'codex',
        name: 'Codex',
        version: '0.9.7',
        installed: false,
        loggedIn: false,
        instanceCount: 0,
    },
    {
        id: 'open_code',
        name: 'Open Code',
        version: '0.4.18',
        installed: false,
        loggedIn: false,
        instanceCount: 0,
    },
]

const DOWNLOAD_BYTES: Record<string, number> = {
    claude_code: 71_303_168,
    codex: 58_720_256,
    open_code: 44_040_192,
}

function flagCovers(value: string | null, agentId: string) {
    if (!value) return false
    if (value === '1' || value === 'all') return true

    return value.split(',').includes(agentId)
}

function initialAgents() {
    const flags = new URLSearchParams(window.location.search)
    const installed = flags.get('installed')
    const loggedIn = flags.get('loggedIn')

    return MOCK_AGENTS.map((agent) => {
        const authorized = flagCovers(loggedIn, agent.id)
        return {
            ...agent,
            installed: authorized || flagCovers(installed, agent.id),
            loggedIn: authorized,
        }
    })
}

let agents = initialAgents()

const wait = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))

const failed = new Set<string>()

function shouldFail(agentId: string) {
    const forced = new URLSearchParams(window.location.search).get('installFail')
    if (!flagCovers(forced, agentId)) return false
    if (failed.has(agentId)) return false

    failed.add(agentId)
    return true
}

export async function mockListAgents(): Promise<Agent[]> {
    return structuredClone(agents)
}

export function mockLoggedInAgentIds(): string[] {
    return agents.filter((agent) => agent.loggedIn).map((agent) => agent.id)
}

function markLoggedIn(agentId: string) {
    agents = agents.map((agent) => (agent.id === agentId ? {...agent, loggedIn: true} : agent))
}

/**
 * The real AuthAgent hands back a URL, the user approves in their browser, and
 * the harness reports logged_in on the next poll. The timer stands in for the
 * part that happens outside the app.
 */
export async function mockStartLogin(agentId: string): Promise<string> {
    await wait(RESOLVE_MS)
    setTimeout(() => markLoggedIn(agentId), APPROVE_MS)

    return `https://auth.example.com/device?agent=${agentId}`
}

export async function mockSubmitAuthCode(agentId: string, code: string): Promise<void> {
    await wait(RESOLVE_MS)

    if (code.trim().length < MIN_CODE_LENGTH) throw new Error(t('agent.error.codeIncomplete'))

    markLoggedIn(agentId)
}

export async function mockLogoutAgent(agentId: string): Promise<void> {
    await wait(RESOLVE_MS)

    agents = agents.map((agent) => (agent.id === agentId ? {...agent, loggedIn: false} : agent))
}

export async function mockInstallAgent(
    agentId: string,
    emit: (progress: InstallProgress) => void,
): Promise<void> {
    const total = DOWNLOAD_BYTES[agentId] ?? 33_554_432

    emit({stage: InstallStage.Resolve, downloaded: 0, total: 0})
    await wait(RESOLVE_MS)

    for (let tick = 1; tick <= DOWNLOAD_TICKS; tick++) {
        emit({
            stage: InstallStage.Download,
            downloaded: Math.round((total * tick) / DOWNLOAD_TICKS),
            total,
        })
        await wait(DOWNLOAD_MS / DOWNLOAD_TICKS)
    }

    if (shouldFail(agentId))
        throw new Error(`${agentId} download failed: unexpected end of archive`)

    emit({stage: InstallStage.Extract, downloaded: total, total})
    await wait(EXTRACT_MS)

    agents = agents.map((agent) => (agent.id === agentId ? {...agent, installed: true} : agent))

    emit({stage: InstallStage.Done, downloaded: total, total})
}

export async function mockUninstallAgent(agentId: string): Promise<void> {
    await wait(RESOLVE_MS)

    agents = agents.map((agent) =>
        agent.id === agentId ? {...agent, installed: false, loggedIn: false} : agent,
    )
}
