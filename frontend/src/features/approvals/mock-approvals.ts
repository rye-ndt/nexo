/**
 * The queue the api seam answers from when there is no Wails runtime, standing
 * in for the Go broker. A simulated run raises a request to block a node and
 * polls mockApprovalAnswer to resume it — the same shape as an agent blocking
 * inside request_approval until the operator answers.
 */

import type {
    Approval,
    ApprovalAnswer,
    ApprovalKind,
    ApprovalOption,
} from '@/features/approvals/types'

let pending: Approval[] = []

const answers = new Map<string, ApprovalAnswer>()

export function raiseMockApproval(input: {
    agentId: string
    question: string
    detail: string
    options: ApprovalOption[]
    multiSelect: boolean
    kind: ApprovalKind
}): string {
    const id = crypto.randomUUID()

    pending = [
        ...pending,
        {
            id,
            agentId: input.agentId,
            kind: input.kind,
            question: input.question,
            detail: input.detail,
            options: structuredClone(input.options),
            multiSelect: input.multiSelect,
            requestedAt: new Date().toISOString(),
        },
    ]

    return id
}

export function listMockApprovals(): Approval[] {
    return structuredClone(pending).sort((a, b) => a.requestedAt.localeCompare(b.requestedAt))
}

export function answerMockApproval(id: string, answer: ApprovalAnswer): void {
    if (!pending.some((approval) => approval.id === id))
        throw new Error('That request is no longer waiting for an answer.')

    pending = pending.filter((approval) => approval.id !== id)
    answers.set(id, structuredClone(answer))
}

export function isMockApprovalPending(id: string): boolean {
    return pending.some((approval) => approval.id === id)
}

export function mockApprovalAnswer(id: string): ApprovalAnswer | undefined {
    const answer = answers.get(id)

    return answer ? structuredClone(answer) : undefined
}

export function forgetMockApprovals(agentIds: Iterable<string>): void {
    const dropped = new Set(agentIds)

    for (const approval of pending) {
        if (dropped.has(approval.agentId)) answers.delete(approval.id)
    }

    pending = pending.filter((approval) => !dropped.has(approval.agentId))
}

export function clearMockApprovals(): void {
    pending = []
    answers.clear()
}
