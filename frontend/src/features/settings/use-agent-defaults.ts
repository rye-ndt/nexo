import {useMutation, useQuery, useQueryClient} from '@tanstack/react-query'

import * as api from '@/features/settings/api/preferences'
import type {TaskLevel, ThinkingLevel} from '@/shared/lib/enums'
import type {TokenPrices} from '@/features/settings/types'

const AGENT_DEFAULTS_KEY = ['agent-defaults']
const AGENT_DEFAULT_OPTIONS_KEY = ['agent-default-options']

type AgentDefaultEdit = {
    taskLevel: TaskLevel
    model: string
    thinkingLevel: ThinkingLevel
}

type AgentPricesEdit = {
    taskLevel: TaskLevel
    prices: TokenPrices
}

export function useAgentDefaults() {
    const queryClient = useQueryClient()

    const defaults = useQuery({
        queryKey: AGENT_DEFAULTS_KEY,
        queryFn: api.listAgentDefaults,
        meta: {action: 'Could not load your preferences'},
    })
    const options = useQuery({
        queryKey: AGENT_DEFAULT_OPTIONS_KEY,
        queryFn: api.agentDefaultOptions,
        meta: {action: 'Could not load the models this app can run'},
    })

    const save = useMutation({
        meta: {action: 'Could not save that default'},
        mutationFn: ({taskLevel, model, thinkingLevel}: AgentDefaultEdit) =>
            api.setAgentDefault(taskLevel, model, thinkingLevel),
        onSuccess: () => queryClient.invalidateQueries({queryKey: AGENT_DEFAULTS_KEY}),
    })

    const savePrices = useMutation({
        meta: {action: 'Could not save that price'},
        mutationFn: ({taskLevel, prices}: AgentPricesEdit) =>
            api.setAgentDefaultPrices(taskLevel, prices),
        onSuccess: () => queryClient.invalidateQueries({queryKey: AGENT_DEFAULTS_KEY}),
    })

    return {
        defaults: defaults.data ?? [],
        options: options.data,
        loading: defaults.isPending || options.isPending,
        pendingTaskLevel: save.isPending ? save.variables.taskLevel : null,
        setAgentDefault: save.mutate,
        setAgentDefaultPrices: savePrices.mutate,
    }
}
