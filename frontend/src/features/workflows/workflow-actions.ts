import {
    CirclePlay,
    CircleStop,
    Copy,
    FolderOpen,
    Lock,
    LockOpen,
    Pause,
    Play,
    Plus,
} from 'lucide-react'
import type {LucideIcon} from 'lucide-react'

import {
    hasActiveStep,
    hasStarted,
    isCancellable,
    isLocked,
    isPausable,
    isResumable,
} from '@/features/workflows/graph'
import {t, type MessageKey} from '@/shared/lib/i18n'
import type {GlossaryTerm} from '@/shared/lib/glossary'
import type {Workflow} from '@/features/workflows/types'

export const WorkflowActionId = {
    NewStep: 'new_step',
    Duplicate: 'duplicate',
    ChooseFolder: 'choose_folder',
    Lock: 'lock',
    Unlock: 'unlock',
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
}

export type WorkflowActionHandlers = Record<WorkflowActionId, () => void>

type ActionShape = {label: MessageKey; icon: LucideIcon; term?: GlossaryTerm}

const ACTIONS: Record<WorkflowActionId, ActionShape> = {
    [WorkflowActionId.NewStep]: {label: 'workflow.action.newStep', icon: Plus},
    [WorkflowActionId.Duplicate]: {label: 'workflow.action.duplicate', icon: Copy},
    [WorkflowActionId.ChooseFolder]: {
        label: 'workflow.action.chooseFolder',
        icon: FolderOpen,
        term: 'projectFolder',
    },
    [WorkflowActionId.Lock]: {label: 'workflow.action.lock', icon: Lock, term: 'lock'},
    [WorkflowActionId.Unlock]: {label: 'workflow.action.unlock', icon: LockOpen, term: 'lock'},
    [WorkflowActionId.Run]: {label: 'workflow.action.run', icon: Play},
    [WorkflowActionId.Pause]: {label: 'workflow.action.pause', icon: Pause},
    [WorkflowActionId.Resume]: {label: 'workflow.action.resume', icon: CirclePlay},
    [WorkflowActionId.Cancel]: {label: 'workflow.action.cancel', icon: CircleStop},
}

const {Primary, Outline, Ghost} = ActionEmphasis

type Offer = [WorkflowActionId, ActionEmphasis]

/** The folder comes before the steps: a duplicate starts without one and cannot lock until it has one. */
function draftChain(workflow: Workflow): Offer[] {
    if (!workflow.projectDir.trim())
        return [
            [WorkflowActionId.NewStep, Outline],
            [WorkflowActionId.Duplicate, Ghost],
            [WorkflowActionId.ChooseFolder, Primary],
        ]

    if (workflow.steps.length === 0)
        return [
            [WorkflowActionId.Duplicate, Ghost],
            [WorkflowActionId.NewStep, Primary],
        ]

    return [
        [WorkflowActionId.NewStep, Outline],
        [WorkflowActionId.Duplicate, Ghost],
        [WorkflowActionId.Lock, Primary],
    ]
}

function chainOf(workflow: Workflow): Offer[] {
    if (!isLocked(workflow)) return draftChain(workflow)

    const cancel: Offer[] = isCancellable(workflow) ? [[WorkflowActionId.Cancel, Outline]] : []

    if (isPausable(workflow))
        return [[WorkflowActionId.Pause, Outline], ...cancel, [WorkflowActionId.Duplicate, Ghost]]

    if (isResumable(workflow))
        return [...cancel, [WorkflowActionId.Duplicate, Ghost], [WorkflowActionId.Resume, Primary]]

    const unlock: Offer[] = hasActiveStep(workflow) ? [] : [[WorkflowActionId.Unlock, Ghost]]

    if (!hasStarted(workflow))
        return [...unlock, [WorkflowActionId.Duplicate, Ghost], [WorkflowActionId.Run, Primary]]

    return [...unlock, [WorkflowActionId.Duplicate, Ghost]]
}

/** Rendered by both the header bar and the mobile menu, so the rules live here only. */
export function workflowActions(workflow: Workflow | null): WorkflowAction[] {
    if (!workflow) return []

    return chainOf(workflow).map(([id, emphasis]) => ({
        id,
        ...ACTIONS[id],
        emphasis,
        label: t(ACTIONS[id].label),
    }))
}
