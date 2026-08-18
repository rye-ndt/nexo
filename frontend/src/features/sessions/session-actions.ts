import {CirclePlay, CircleStop, Copy, Lock, Pause, Play, Plus} from 'lucide-react'
import type {LucideIcon} from 'lucide-react'

import {isCancellable, isPausable, isResumable} from '@/features/sessions/graph'
import type {Session} from '@/features/sessions/types'

export const SessionActionId = {
    NewNode: 'new_node',
    Clone: 'clone',
    Finalize: 'finalize',
    Run: 'run',
    Pause: 'pause',
    Resume: 'resume',
    Cancel: 'cancel',
} as const

export type SessionActionId = (typeof SessionActionId)[keyof typeof SessionActionId]

export const ActionEmphasis = {
    Primary: 'primary',
    Outline: 'outline',
    Ghost: 'ghost',
} as const

export type ActionEmphasis = (typeof ActionEmphasis)[keyof typeof ActionEmphasis]

export type SessionAction = {
    id: SessionActionId
    label: string
    icon: LucideIcon
    emphasis: ActionEmphasis
}

export type SessionActionHandlers = Record<SessionActionId, () => void>

const ACTIONS: Record<SessionActionId, Omit<SessionAction, 'id'>> = {
    [SessionActionId.NewNode]: {label: 'New node', icon: Plus, emphasis: ActionEmphasis.Outline},
    [SessionActionId.Clone]: {label: 'Duplicate', icon: Copy, emphasis: ActionEmphasis.Ghost},
    [SessionActionId.Finalize]: {label: 'Finalize', icon: Lock, emphasis: ActionEmphasis.Outline},
    [SessionActionId.Run]: {label: 'Run', icon: Play, emphasis: ActionEmphasis.Primary},
    [SessionActionId.Pause]: {
        label: 'Pause run',
        icon: Pause,
        emphasis: ActionEmphasis.Outline,
    },
    [SessionActionId.Resume]: {
        label: 'Resume run',
        icon: CirclePlay,
        emphasis: ActionEmphasis.Primary,
    },
    [SessionActionId.Cancel]: {
        label: 'Cancel run',
        icon: CircleStop,
        emphasis: ActionEmphasis.Outline,
    },
}

/** Rendered by both the header bar and the mobile menu, so the rules live here only. */
export function sessionActions(session: Session | null): SessionAction[] {
    if (!session) return []

    const draft = !session.finalized

    const ids = [
        draft && SessionActionId.NewNode,
        SessionActionId.Clone,
        draft && SessionActionId.Finalize,
        session.finalized && !session.started && SessionActionId.Run,
        isPausable(session) && SessionActionId.Pause,
        isResumable(session) && SessionActionId.Resume,
        isCancellable(session) && SessionActionId.Cancel,
    ].filter((id): id is SessionActionId => Boolean(id))

    return ids.map((id) => ({id, ...ACTIONS[id]}))
}
