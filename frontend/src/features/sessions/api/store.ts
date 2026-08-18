/**
 * Sessions are authored in this module's memory. Inside the Wails webview they
 * are hydrated from the stored drafts and every change is written back as one
 * JSON doc.
 */

import {bridge, hasWailsRuntime} from '@/shared/api/bridge'
import {byRailRank, isPausable, pauseRun} from '@/features/sessions/graph'
import {MOCK_SESSIONS} from '@/features/sessions/mock-sessions'
import type {Session} from '@/features/sessions/types'
import {SaveSessionDraft, SessionDrafts} from '@wailsjs/go/wails_api/API'

export let sessions: Session[] = hasWailsRuntime() ? [] : structuredClone(MOCK_SESSIONS)

let hydrated = false

export function setSessions(next: Session[]) {
    sessions = next
}

export function prependSession(session: Session): Session {
    const top = Math.min(0, ...sessions.map((existing) => existing.railRank ?? 0))
    const ranked = {...session, railRank: top - 1}

    setSessions([ranked, ...sessions])

    return ranked
}

export function findSession(sessionId: string) {
    const session = sessions.find((session) => session.id === sessionId)
    if (!session) throw new Error('That session is gone. Pick another one from the rail.')
    return session
}

export function findOpenSession(sessionId: string) {
    const session = findSession(sessionId)
    if (session.finalized)
        throw new Error('This session is finalized. Duplicate it to make changes.')
    return session
}

export function findTask(session: Session, taskId: string) {
    const task = session.tasks.find((task) => task.id === taskId)
    if (!task) throw new Error('That task is gone. Pick another one on the canvas.')
    return task
}

export function replaceSession(next: Session) {
    sessions = sessions.map((session) => (session.id === next.id ? next : session))
    void saveDraft(next).catch(() => {})
    return structuredClone(next)
}

export async function saveDraft(session: Session) {
    if (!hasWailsRuntime()) return

    await bridge(() => SaveSessionDraft(session.id, JSON.stringify(session)))
}

export async function hydrate() {
    if (hydrated || !hasWailsRuntime()) return

    const drafts = await bridge(SessionDrafts)
    sessions = drafts.map((draft) => restore(JSON.parse(draft.doc) as Session)).sort(byRailRank)
    hydrated = true
}

/**
 * No agent survives the process, so a run that was still going comes back paused —
 * the same nodes it would have lost to a pause. A cancelled session is terminal, so
 * it comes back exactly as it was stored.
 */
function restore(stored: Session): Session {
    if (stored.cancelled) return {...stored, cancelled: true}

    const halted = {...stored, cancelled: false, paused: false}

    return {...pauseRun(halted), paused: isPausable(halted)}
}
