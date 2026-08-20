import {useQuery} from '@tanstack/react-query'

import * as api from '@/features/workflows/api'
import {t} from '@/shared/lib/i18n'

/** Read when the step is opened, not carried on its report, so one step is one call. */
export function useStepDiff(workflowId: string, stepId: string) {
    const {data, isPending, isError, refetch} = useQuery({
        queryKey: ['step-diff', workflowId, stepId],
        queryFn: () => api.fetchStepDiff(workflowId, stepId),
        meta: {action: t('workflow.error.diff')},
    })

    return {
        changes: data ?? [],
        loading: isPending,
        failed: isError,
        retry: () => void refetch(),
    }
}
