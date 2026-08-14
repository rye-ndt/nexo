import {useMutation, useQuery} from '@tanstack/react-query'

import * as api from '@/features/templates/api'
import type {DraftContext, TemplateDraft} from '@/features/templates/types'

const HELPER_KEY = ['template-helper']

/** Cheap, and it changes when the user logs a harness in from another screen. */
const STALE_MS = 15_000

export function useTemplateHelper(
    onFilled: (filled: TemplateDraft, sent: TemplateDraft) => void,
    context?: DraftContext,
) {
    const {data} = useQuery({
        queryKey: HELPER_KEY,
        queryFn: api.templateHelperBlocked,
        staleTime: STALE_MS,
    })

    const fill = useMutation({
        meta: {action: 'Could not fill this template in'},
        mutationFn: (draft: TemplateDraft) => api.refineTemplate(draft, context),
        onSuccess: onFilled,
    })

    return {
        blocked: data ?? '',
        fillIn: fill.mutate,
        filling: fill.isPending,
    }
}
