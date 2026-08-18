import {useState} from 'react'

import {useSessions} from '@/features/sessions/use-sessions'
import * as graph from '@/features/sessions/graph'
import type {ParamValue} from '@/features/templates/types'
import type {
    Point,
    Session,
    SessionDraft,
    SessionLocations,
    Task,
    TaskDraft,
} from '@/features/sessions/types'

/**
 * Single source of truth for the session graph. Sessions are server state and
 * live in the query cache; the active session and the selected task are view
 * state and stay local.
 *
 * Rail-level writes name the session they touch. Everything under `active` acts
 * on the open session and is inert when there is none, so no caller repeats
 * that guard.
 */
export function useSessionStore() {
    const store = useSessions()
    const {sessions} = store

    const [selectedSessionId, setSelectedSessionId] = useState<string | null>(null)
    const [selectedTaskId, setSelectedTaskId] = useState<string | null>(null)

    const activeSession = graph.findSession(sessions, selectedSessionId) ?? sessions[0] ?? null
    const activeSessionId = activeSession?.id ?? null

    const selectSession = (sessionId: string | null) => {
        setSelectedSessionId(sessionId)
        setSelectedTaskId(null)
    }

    /** Lifts a write that needs a session id into one the open session supplies. */
    const onActive =
        <TArgs extends unknown[]>(write: (sessionId: string, ...args: TArgs) => void) =>
        (...args: TArgs) => {
            if (activeSessionId) write(activeSessionId, ...args)
        }

    /** A draft is the only thing structural edits may touch; a finalized graph is frozen. */
    const openSession = (sessionId: string) => {
        const session = graph.findSession(sessions, sessionId)
        return session && !session.finalized ? session : null
    }

    const addSession = (draft: SessionDraft) => {
        const sessionId = crypto.randomUUID()
        store.createSession({sessionId, draft})
        selectSession(sessionId)
    }

    const cloneSession = (sourceId: string) => {
        const sessionId = crypto.randomUUID()
        store.cloneSession({sourceId, sessionId})
        selectSession(sessionId)
    }

    const importSession = (session: Session, locations: SessionLocations) => {
        store.importSession({session, locations})
        selectSession(session.id)
    }

    const reorderSession = (sessionId: string, toIndex: number) => {
        store.reorderSession({sessionId, toIndex})
    }

    const deleteSession = (sessionId: string) => {
        store.deleteSession({sessionId})
        if (selectedSessionId === sessionId) selectSession(null)
    }

    const addTask = (sessionId: string, draft: TaskDraft, position: Point) => {
        const session = openSession(sessionId)
        if (!session) return

        const taskId = crypto.randomUUID()
        store.createTask({
            sessionId,
            taskId,
            draft,
            position: graph.freePosition(session, position),
        })
        setSelectedTaskId(taskId)
    }

    const removeTask = (sessionId: string, taskId: string) => {
        store.deleteTask({sessionId, taskId})
        if (selectedTaskId === taskId) setSelectedTaskId(null)
    }

    const connectTasks = (sessionId: string, sourceId: string, targetId: string) => {
        const session = openSession(sessionId)
        if (!session || graph.createsCycle(session.tasks, sourceId, targetId)) return

        store.connectTasks({sessionId, sourceId, targetId})
    }

    const active = {
        session: activeSession,

        rename: onActive((sessionId, name: string) => store.renameSession({sessionId, name})),
        setLocations: onActive((sessionId, locations: SessionLocations) =>
            store.setSessionLocations({sessionId, ...locations}),
        ),
        finalize: onActive((sessionId) => store.finalizeSession({sessionId})),
        start: onActive((sessionId) => store.startSession({sessionId})),
        pause: onActive((sessionId, onSettled?: () => void) =>
            store.pauseSession({sessionId}, {onSettled}),
        ),
        resume: onActive((sessionId) => store.resumeSession({sessionId})),
        cancel: onActive((sessionId, onSettled?: () => void) =>
            store.cancelSession({sessionId}, {onSettled}),
        ),
        clone: onActive((sessionId) => cloneSession(sessionId)),

        addTask: onActive((sessionId, draft: TaskDraft, position: Point) =>
            addTask(sessionId, draft, position),
        ),
        updateTask: onActive((sessionId, taskId: string, patch: Partial<Task>) =>
            store.updateTask({sessionId, taskId, patch}),
        ),
        moveTask: onActive((sessionId, taskId: string, position: Point) =>
            store.moveTask({sessionId, taskId, position}),
        ),
        removeTask: onActive((sessionId, taskId: string) => removeTask(sessionId, taskId)),
        saveTaskInputs: onActive((sessionId, taskId: string, values: Record<string, ParamValue>) =>
            store.saveTaskInputs({sessionId, taskId, values}),
        ),
        answerTaskAcceptance: onActive((sessionId, taskId: string, accepted: boolean) =>
            store.answerTaskAcceptance({sessionId, taskId, accepted}),
        ),
        revertToTask: onActive((sessionId, taskId: string, onSettled?: () => void) =>
            store.revertToTask({sessionId, taskId}, {onSettled}),
        ),

        connectTasks: onActive((sessionId, sourceId: string, targetId: string) =>
            connectTasks(sessionId, sourceId, targetId),
        ),
        disconnectTasks: onActive((sessionId, sourceId: string, targetId: string) =>
            store.disconnectTasks({sessionId, sourceId, targetId}),
        ),
    }

    return {
        sessions,
        activeSessionId,
        active,

        selectedTaskId,
        selectSession,
        selectTask: setSelectedTaskId,

        addSession,
        cloneSession,
        importSession,
        reorderSession,
        deleteSession,

        pausing: store.pausingSession,
        resuming: store.resumingSession,
        cancelling: store.cancellingSession,
        savingTaskInputs: store.savingTaskInputs,
        answeringTaskAcceptance: store.answeringTaskAcceptance,
        revertingToTask: store.revertingToTask,
    }
}

export type SessionStore = ReturnType<typeof useSessionStore>
