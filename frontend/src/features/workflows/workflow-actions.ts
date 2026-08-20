import {CirclePlay, CircleStop, Copy, Lock, Pause, Play, Plus} from 'lucide-react'
import type {LucideIcon} from 'lucide-react'

import {isCancellable, isPausable, isResumable} from '@/features/workflows/graph'
import type {GlossaryTerm} from '@/shared/lib/glossary'
import type {Workflow} from '@/features/workflows/types'

export const WorkflowActionId = {
    NewStep: 'new_step',
    Duplicate: 'duplicate',
    Lock: 'lock',
    Run: 'run',
    Pause: 'pause',
    Resume: 'resume',
    Cancel: 'cancel',
} as const

export type WorkflowActionId = (typeof WorkflowActionId)[keyof typeof WorkflowActionId]

export const ActionEmphasis = {
    Primary: 'primary',
    Outline: 'outline',
    Ghost: 'ghost',
} as const

export type ActionEmphasis = (typeof ActionEmphasis)[keyof typeof ActionEmphasis]

export type WorkflowAction = {
    id: WorkflowActionId
    label: string
    icon: LucideIcon
    emphasis: ActionEmphasis
    term?: GlossaryTerm
    disabledReason?: string
}

export type WorkflowActionHandlers = Record<WorkflowActionId, () => void>

const ACTIONS: Record<WorkflowActionId, Omit<WorkflowAction, 'id'>> = {
    [WorkflowActionId.NewStep]: {label: 'New step', icon: Plus, emphasis: ActionEmphasis.Outline},
    [WorkflowActionId.Duplicate]: {label: 'Duplicate', icon: Copy, emphasis: ActionEmphasis.Ghost},
    [WorkflowActionId.Lock]: {
        label: 'Lock',
        icon: Lock,
        emphasis: ActionEmphasis.Outline,
        term: 'lock',
    },
    [WorkflowActionId.Run]: {label: 'Run', icon: Play, emphasis: ActionEmphasis.Primary},
    [WorkflowActionId.Pause]: {
        label: 'Pause run',
        icon: Pause,
        emphasis: ActionEmphasis.Outline,
    },
    [WorkflowActionId.Resume]: {
        label: 'Resume run',
        icon: CirclePlay,
        emphasis: ActionEmphasis.Primary,
    },
    [WorkflowActionId.Cancel]: {
        label: 'Cancel run',
        icon: CircleStop,
        emphasis: ActionEmphasis.Outline,
    },
}

/** Rendered by both the header bar and the mobile menu, so the rules live here only. */
export function workflowActions(workflow: Workflow | null): WorkflowAction[] {
    if (!workflow) return []

    const draft = !workflow.locked

    const ids = [
        draft && WorkflowActionId.NewStep,
        WorkflowActionId.Duplicate,
        draft && WorkflowActionId.Lock,
        workflow.locked && !workflow.started && WorkflowActionId.Run,
        isPausable(workflow) && WorkflowActionId.Pause,
        isResumable(workflow) && WorkflowActionId.Resume,
        isCancellable(workflow) && WorkflowActionId.Cancel,
    ].filter((id): id is WorkflowActionId => Boolean(id))

    return ids.map((id) => ({id, ...ACTIONS[id], disabledReason: reasonToWait(workflow, id)}))
}

function reasonToWait(workflow: Workflow, id: WorkflowActionId) {
    if (id === WorkflowActionId.Lock && !workflow.projectDir.trim())
        return 'Choose a project folder before locking this workflow.'

    return undefined
}
