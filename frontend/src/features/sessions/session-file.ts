/**
 * The two pure halves of session transfer. A session travels without its node
 * templates, so what each template would have told the run is inlined onto the
 * node on the way out and read back off it on the way in.
 */

import {duplicateSession} from '@/features/sessions/graph'
import {specOf} from '@/features/sessions/task-spec'
import type {Session} from '@/features/sessions/types'
import type {Template} from '@/features/templates/types'

const NOT_A_SESSION = 'That file is not a session export.'

export function toExportedSession(session: Session, templates: Template[]): Session {
    const clean = duplicateSession(session, session.name)

    return {
        ...clean,
        tasks: clean.tasks.map((task) => ({
            ...task,
            spec: specOf(task, templates),
            templateId: undefined,
        })),
    }
}

function isPoint(value: unknown): boolean {
    const point = value as {x?: unknown; y?: unknown} | null
    return typeof point?.x === 'number' && typeof point?.y === 'number'
}

function readSession(raw: string): Session {
    let parsed: Session

    try {
        parsed = JSON.parse(raw) as Session
    } catch {
        throw new Error(NOT_A_SESSION)
    }

    if (typeof parsed?.name !== 'string' || !Array.isArray(parsed.tasks))
        throw new Error(NOT_A_SESSION)

    const ids = new Set(parsed.tasks.map((task) => task?.id))

    for (const task of parsed.tasks) {
        if (typeof task?.id !== 'string') throw new Error(NOT_A_SESSION)
        if (typeof task.title !== 'string' || typeof task.prompt !== 'string')
            throw new Error(NOT_A_SESSION)
        if (!isPoint(task.position)) throw new Error(NOT_A_SESSION)
        if (!Array.isArray(task.dependsOn)) throw new Error(NOT_A_SESSION)
        if (task.dependsOn.some((id) => !ids.has(id))) throw new Error(NOT_A_SESSION)
    }

    return parsed
}

export function fromExportedSession(raw: string): Session {
    const parsed = readSession(raw)
    return duplicateSession(parsed, parsed.name)
}
