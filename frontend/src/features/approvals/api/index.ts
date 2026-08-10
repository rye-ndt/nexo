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

async function roundtrip() {
    await new Promise((resolve) => setTimeout(resolve, ROUNDTRIP_MS))
}

export async function listApprovals(): Promise<Approval[]> {
    if (!hasWailsRuntime()) return listMockApprovals()

    const infos = await bridge(() => PendingApprovals())

    return infos.map((info) => ({
        id: info.id,
        agentId: info.agent_id,
        kind:
            info.kind === ApprovalKind.Permission ? ApprovalKind.Permission : ApprovalKind.Decision,
        question: info.question,
        detail: info.detail,
        options: (info.options ?? []).map((option) => ({
            id: option.id,
            label: option.label,
            description: option.description || '',
        })),
        multiSelect: info.multi_select,
        requestedAt: info.requested_at,
    }))
}

export async function answerApproval(id: string, answer: ApprovalAnswer): Promise<void> {
    if (hasWailsRuntime()) {
        await bridge(() => AnswerApproval(id, answer.approved, answer.optionIds, answer.guidance))
        return
    }

    await roundtrip()

    answerMockApproval(id, answer)
}
