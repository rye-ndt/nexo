import {useAgentDefaults} from '@/features/settings/use-agent-defaults'
import {useAgents} from '@/features/agents/use-agents'
import {requiredHarnesses} from '@/features/settings/agent-default'
import type {Agent} from '@/features/agents/types'

export type RequiredAgent = {
    harness: string
    modelLabels: string[]
    agent: Agent | null
}

export function isAgentReady(required: RequiredAgent) {
    return Boolean(required.agent?.installed && required.agent.loggedIn)
}

/**
 * The agents the current preferences commit a workflow to. Every step level may
 * be used once the graph has steps, so all four defaults count — not just the
 * ones a workflow happens to reach.
 */
export function useRequiredAgents() {
    const {defaults, options, loading: defaultsLoading} = useAgentDefaults()
    const controls = useAgents()

    const required: RequiredAgent[] = requiredHarnesses(defaults, options?.models ?? []).map(
        (demand) => ({
            ...demand,
            agent: controls.agents.find((agent) => agent.id === demand.harness) ?? null,
        }),
    )

    const loading = defaultsLoading || controls.loading

    return {
        required,
        ready: !loading && required.length > 0 && required.every(isAgentReady),
        loading,
        controls,
    }
}
