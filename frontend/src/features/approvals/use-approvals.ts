import {useMutation, useQuery, useQueryClient} from '@tanstack/react-query'

import * as api from '@/features/approvals/api'
import {t} from '@/shared/lib/i18n'
import type {ApprovalAnswer} from '@/features/approvals/types'

const APPROVALS_KEY = ['approvals']

const POLL_MS = 900

/** An agent is stopped for as long as its request is pending, so this polls for the whole run. */
export function useApprovals() {
    const queryClient = useQueryClient()

    const {data} = useQuery({
        queryKey: APPROVALS_KEY,
        queryFn: api.listApprovals,
        meta: {action: t('approval.error.load')},
        refetchInterval: POLL_MS,
        refetchIntervalInBackground: true,
    })

    const answer = useMutation({
        meta: {action: t('approval.error.answer')},
        mutationFn: (args: {id: string; answer: ApprovalAnswer}) =>
            api.answerApproval(args.id, args.answer),
        onSettled: () => {
            queryClient.invalidateQueries({queryKey: APPROVALS_KEY})
        },
    })

    return {
        approvals: data ?? [],
        answering: answer.isPending,
        answer: (id: string, value: ApprovalAnswer) => answer.mutate({id, answer: value}),
    }
}
