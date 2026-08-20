import {useQuery} from '@tanstack/react-query'

import * as api from '@/features/workflows/api'
import {t} from '@/shared/lib/i18n'
import type {ActivityLine} from '@/features/workflows/types'

const ACTIVITY_POLL_MS = 900

const NO_LINES: ActivityLine[] = []

/**
 * What one running step is saying about itself. The feed is already in memory
 * by the time this asks for it — the run poll puts it there — so a query per
 * running step costs nothing and can live in the leaf that renders it.
 */
export function useStepActivity(stepId: string, enabled: boolean) {
    const {data} = useQuery({
        queryKey: ['step-activity', stepId],
        queryFn: () => api.stepActivity(stepId),
        meta: {action: t('workflow.error.activity')},
        refetchInterval: enabled ? ACTIVITY_POLL_MS : false,
        refetchIntervalInBackground: true,
    })

    return data ?? NO_LINES
}
