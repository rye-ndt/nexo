import {useMutation, useQuery, useQueryClient} from '@tanstack/react-query'

import * as api from '@/features/settings/api/preferences'
import {t} from '@/shared/lib/i18n'

const AUTOPILOT_KEY = ['autopilot']

export function useAutopilot() {
    const queryClient = useQueryClient()

    const autopilot = useQuery({
        queryKey: AUTOPILOT_KEY,
        queryFn: api.autopilot,
        meta: {action: t('settings.error.loadAutopilot')},
    })

    const save = useMutation({
        meta: {action: t('settings.error.saveAutopilot')},
        mutationFn: (on: boolean) => api.setAutopilot(on),
        onSuccess: () => queryClient.invalidateQueries({queryKey: AUTOPILOT_KEY}),
    })

    return {
        autopilot: autopilot.data ?? false,
        loading: autopilot.isPending,
        saving: save.isPending,
        setAutopilot: save.mutate,
    }
}
