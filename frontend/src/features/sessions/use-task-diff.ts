import {useQuery} from '@tanstack/react-query'

import * as api from '@/features/sessions/api'

/** Read when the step is opened, not carried on its report, so one node is one call. */
export function useTaskDiff(sessionId: string, taskId: string) {
    const {data, isPending, isError, refetch} = useQuery({
        queryKey: ['task-diff', sessionId, taskId],
        queryFn: () => api.fetchTaskDiff(sessionId, taskId),
        meta: {action: 'Could not read what this step changed'},
    })

    return {
        changes: data ?? [],
        loading: isPending,
        failed: isError,
        retry: () => void refetch(),
    }
}
