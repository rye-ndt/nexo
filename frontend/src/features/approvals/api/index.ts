/**
 * The only place the generated Wails bindings for approvals are touched. An
 * agent is blocked inside request_approval while a request is pending, so the
 * list is polled and the answer is a single write. Under the plain vite dev
 * server the same contract is answered from features/approvals/mock-approvals.ts.
 */

import {bridge, hasWailsRuntime} from '@/shared/api/bridge'
import {ApprovalKind} from '@/shared/lib/enums'
import {answerMockApproval, listMockApprovals} from '@/features/approvals/mock-approvals'
import type {Approval, ApprovalAnswer} from '@/features/approvals/types'
import {AnswerApproval, PendingApprovals} from '@wailsjs/go/wails_api/API'

const ROUNDTRIP_MS = 400

type ApprovalsBackend = {
    list(): Promise<Approval[]>
    answer(id: string, answer: ApprovalAnswer): Promise<void>
}

async function roundtrip() {
    await new Promise((resolve) => setTimeout(resolve, ROUNDTRIP_MS))
}

const wailsApprovals: ApprovalsBackend = {
    list: async () => {
        const infos = await bridge(() => PendingApprovals())

        return infos.map((info) => ({
            id: info.id,
            agentId: info.agent_id,
            kind:
                info.kind === ApprovalKind.Permission
                    ? ApprovalKind.Permission
                    : ApprovalKind.Decision,
            question: info.question,
            detail: info.detail,
            options: (info.options ?? []).map((option) => ({
                id: option.id,
                label: option.label,
                description: option.description || '',
                recommended: option.recommended ?? false,
            })),
            multiSelect: info.multi_select,
            requestedAt: info.requested_at,
        }))
    },
    answer: async (id, answer) => {
        await bridge(() => AnswerApproval(id, answer.approved, answer.optionIds, answer.guidance))
    },
}

const mockApprovals: ApprovalsBackend = {
    list: async () => listMockApprovals(),
    answer: async (id, answer) => {
        await roundtrip()
        answerMockApproval(id, answer)
    },
}

const backend: ApprovalsBackend = hasWailsRuntime() ? wailsApprovals : mockApprovals

export async function listApprovals(): Promise<Approval[]> {
    return backend.list()
}

export async function answerApproval(id: string, answer: ApprovalAnswer): Promise<void> {
    return backend.answer(id, answer)
}
