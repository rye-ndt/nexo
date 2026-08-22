import {useMutation, useQuery, useQueryClient} from '@tanstack/react-query'

import * as api from '@/features/settings/api/preferences'
import {t} from '@/shared/lib/i18n'

const MAX_RUNNING_AGENTS_KEY = ['max-running-agents']

export function useMaxRunningAgents() {
    const queryClient = useQueryClient()

    const maxRunningAgents = useQuery({
        queryKey: MAX_RUNNING_AGENTS_KEY,
        queryFn: api.maxRunningAgents,
        meta: {action: t('settings.error.loadMaxRunningAgents')},
    })

    const save = useMutation({
        meta: {action: t('settings.error.saveMaxRunningAgents')},
        mutationFn: (limit: number) => api.setMaxRunningAgents(limit),
        onSuccess: () => queryClient.invalidateQueries({queryKey: MAX_RUNNING_AGENTS_KEY}),
    })

    return {
        maxRunningAgents: maxRunningAgents.data ?? 1,
        loading: maxRunningAgents.data === undefined,
        saving: save.isPending,
        setMaxRunningAgents: save.mutate,
    }
}
