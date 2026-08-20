import {CirclePlay, CircleStop, Copy, Lock, Pause, Play, Plus} from 'lucide-react'
import type {LucideIcon} from 'lucide-react'

import {isCancellable, isPausable, isResumable} from '@/features/workflows/graph'
import {t, type MessageKey} from '@/shared/lib/i18n'
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

type ActionShape = Omit<WorkflowAction, 'id' | 'label' | 'disabledReason'> & {label: MessageKey}

const ACTIONS: Record<WorkflowActionId, ActionShape> = {
    [WorkflowActionId.NewStep]: {
        label: 'workflow.action.newStep',
        icon: Plus,
        emphasis: ActionEmphasis.Outline,
    },
    [WorkflowActionId.Duplicate]: {
        label: 'workflow.action.duplicate',
        icon: Copy,
        emphasis: ActionEmphasis.Ghost,
    },
    [WorkflowActionId.Lock]: {
        label: 'workflow.action.lock',
        icon: Lock,
        emphasis: ActionEmphasis.Outline,
        term: 'lock',
    },
    [WorkflowActionId.Run]: {
        label: 'workflow.action.run',
        icon: Play,
        emphasis: ActionEmphasis.Primary,
    },
    [WorkflowActionId.Pause]: {
        label: 'workflow.action.pause',
        icon: Pause,
        emphasis: ActionEmphasis.Outline,
    },
    [WorkflowActionId.Resume]: {
        label: 'workflow.action.resume',
        icon: CirclePlay,
        emphasis: ActionEmphasis.Primary,
    },
    [WorkflowActionId.Cancel]: {
        label: 'workflow.action.cancel',
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

    return ids.map((id) => ({
        id,
        ...ACTIONS[id],
        label: t(ACTIONS[id].label),
        disabledReason: reasonToWait(workflow, id),
    }))
}

function reasonToWait(workflow: Workflow, id: WorkflowActionId) {
    if (id === WorkflowActionId.Lock && !workflow.projectDir.trim())
        return t('workflow.api.folderBeforeLock')

    return undefined
}
