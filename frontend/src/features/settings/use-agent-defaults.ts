import {useMutation, useQuery, useQueryClient} from '@tanstack/react-query'

import * as api from '@/features/settings/api/preferences'
import {t} from '@/shared/lib/i18n'
import type {Effort, ThinkingLevel} from '@/shared/lib/enums'

export const AGENT_DEFAULTS_KEY = ['agent-defaults']
export const AGENT_DEFAULT_OPTIONS_KEY = ['agent-default-options']

type AgentDefaultEdit = {
    effort: Effort
    model: string
    thinkingLevel: ThinkingLevel
}

export function useAgentDefaults() {
    const queryClient = useQueryClient()

    const defaults = useQuery({
        queryKey: AGENT_DEFAULTS_KEY,
        queryFn: api.listAgentDefaults,
        meta: {action: t('settings.error.loadPreferences')},
    })
    const options = useQuery({
        queryKey: AGENT_DEFAULT_OPTIONS_KEY,
        queryFn: api.agentDefaultOptions,
        meta: {action: t('settings.error.loadModels')},
    })

    const save = useMutation({
        meta: {action: t('settings.error.saveDefault')},
        mutationFn: ({effort, model, thinkingLevel}: AgentDefaultEdit) =>
            api.setAgentDefault(effort, model, thinkingLevel),
        onSuccess: () => queryClient.invalidateQueries({queryKey: AGENT_DEFAULTS_KEY}),
    })

    return {
        defaults: defaults.data ?? [],
        options: options.data,
        loading: defaults.isPending || options.isPending,
        pendingEffort: save.isPending ? save.variables.effort : null,
        setAgentDefault: save.mutate,
    }
}
