import {useMutation, useQuery} from '@tanstack/react-query'

import * as api from '@/features/roles/api'
import {t} from '@/shared/lib/i18n'
import type {DraftContext, RoleDraft} from '@/features/roles/types'

const HELPER_KEY = ['role-helper']

/** Cheap, and it changes when the user logs a harness in from another screen. */
const STALE_MS = 15_000

export function useRoleHelper(
    onFilled: (filled: RoleDraft, sent: RoleDraft) => void,
    context?: DraftContext,
) {
    const {data} = useQuery({
        queryKey: HELPER_KEY,
        queryFn: api.roleHelperBlocked,
        staleTime: STALE_MS,
    })

    const fill = useMutation({
        meta: {action: t('role.error.fill')},
        mutationFn: (draft: RoleDraft) => api.refineRole(draft, context),
        onSuccess: onFilled,
    })

    return {
        blocked: data ?? '',
        fillIn: fill.mutate,
        filling: fill.isPending,
    }
}
