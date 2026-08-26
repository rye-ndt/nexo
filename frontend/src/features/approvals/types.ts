import type {ApprovalKind} from '@/shared/lib/enums'

export type {ApprovalKind}

export type ApprovalOption = {
    id: string
    label: string
    description: string
    recommended: boolean
}

export type Approval = {
    id: string
    agentId: string
    kind: ApprovalKind
    question: string
    detail: string
    options: ApprovalOption[]
    multiSelect: boolean
    requestedAt: string
}

export type ApprovalAnswer = {approved: boolean; optionIds: string[]; guidance: string}
