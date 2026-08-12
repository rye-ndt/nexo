import {useEffect, useState} from 'react'
import {useMutation, useQuery, useQueryClient} from '@tanstack/react-query'

import * as api from '@/features/agents/api'
import {AGENT_ACTION_FAILURES, AgentAction, InstallStage} from '@/shared/lib/enums'
import {formatAgentName} from '@/shared/lib/format'
import type {InstallProgress} from '@/features/agents/types'

const AGENTS_KEY = ['agents']

const AUTH_POLL_MS = 2000

type PendingAction = {
    agentId: string
    kind: AgentAction
    /** Read by the query client's error handler to title the dialog. */
    action: string
    run: () => Promise<void>
}

function withoutKey(record: Record<string, string>, key: string) {
    const next = {...record}
    delete next[key]
    return next
}

function progressLabel(progress: InstallProgress) {
    if (progress.stage !== InstallStage.Download || progress.total <= 0) return progress.stage

    return `Downloading ${Math.round((progress.downloaded / progress.total) * 100)}%`
}

export function useAgents() {
    const queryClient = useQueryClient()
    const [authUrls, setAuthUrls] = useState<Record<string, string>>({})
    const [progress, setProgress] = useState<Record<string, string>>({})

    const agentsQuery = useQuery({
        queryKey: AGENTS_KEY,
        queryFn: api.listAgents,
        meta: {action: 'Could not load the agents'},
        refetchInterval: (query) => {
            const awaitingLogin = query.state.data?.some(
                (agent) => Boolean(authUrls[agent.id]) && !agent.loggedIn,
            )
            return awaitingLogin ? AUTH_POLL_MS : false
        },
        refetchIntervalInBackground: true,
    })

    const mutation = useMutation({
        mutationFn: (input: PendingAction) => input.run(),
        onSettled: (_data, _error, input) => {
            setProgress((current) => withoutKey(current, input.agentId))
            queryClient.invalidateQueries({queryKey: AGENTS_KEY})
        },
    })

    useEffect(
        () =>
            api.onInstallProgress((agentId, update) => {
                setProgress((current) => ({...current, [agentId]: progressLabel(update)}))
            }),
        [],
    )

    const pending = mutation.isPending ? mutation.variables : null

    const start = (agentId: string, kind: AgentAction, run: () => Promise<void>) =>
        mutation.mutate({
            agentId,
            kind,
            action: `${AGENT_ACTION_FAILURES[kind]} ${formatAgentName(agentId)}`,
            run,
        })

    const install = (agentId: string) =>
        start(agentId, AgentAction.Install, () => api.installAgent(agentId))

    const uninstall = (agentId: string) =>
        start(agentId, AgentAction.Uninstall, () => api.uninstallAgent(agentId))

    const logIn = (agentId: string) =>
        start(agentId, AgentAction.LogIn, async () => {
            const url = await api.startAgentLogin(agentId)
            if (url) setAuthUrls((current) => ({...current, [agentId]: url}))
        })

    const submitAuthCode = (agentId: string, code: string) =>
        start(agentId, AgentAction.Verify, async () => {
            await api.submitAgentAuthCode(agentId, code)
            setAuthUrls((current) => withoutKey(current, agentId))
        })

    return {
        agents: agentsQuery.data ?? [],
        loading: agentsQuery.isPending,
        busy: pending !== null,
        actionOf: (agentId: string) => (pending?.agentId === agentId ? pending.kind : null),
        progressOf: (agentId: string) =>
            pending?.agentId === agentId ? (progress[agentId] ?? null) : null,
        authUrlOf: (agentId: string) => authUrls[agentId] ?? null,
        install,
        uninstall,
        logIn,
        submitAuthCode,
        openAuthUrl: api.openExternalURL,
    }
}
