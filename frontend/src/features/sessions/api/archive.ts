/**
 * The session-transfer seam. Inside the Wails webview Go owns the file and its
 * version envelope; outside it the same envelope is written into the stand-in
 * filesystem, so a round trip is testable on port 8888.
 *
 * Reading a file writes nothing: it answers a candidate session, and the
 * directories it will run in are asked for before it lands in the rail.
 */

import {bridge, hasWailsRuntime} from '@/shared/api/bridge'
import {fromExportedSession, toExportedSession} from '@/features/sessions/session-file'
import {listTemplates} from '@/features/templates/api'
import {mockSessionArchive, mockSessionBody} from '@/features/sessions/mock-sessions'
import {mockReadFile, mockWriteFile} from '@/shared/api/mock-fs'
import type {Session, SessionLocations} from '@/features/sessions/types'
import {findSession, hydrate, saveDraft, sessions, setSessions} from '@/features/sessions/api/store'
import {ExportSession, ImportSession} from '@wailsjs/go/wails_api/API'

export async function exportSession(sessionId: string, path: string): Promise<string> {
    await hydrate()

    const session = findSession(sessionId)
    const body = JSON.stringify(toExportedSession(session, await listTemplates()))

    if (hasWailsRuntime()) await bridge(() => ExportSession(path, body))
    else mockWriteFile(path, mockSessionArchive(body))

    return session.name
}

export async function readSessionFile(path: string): Promise<Session> {
    const body = hasWailsRuntime()
        ? await bridge(() => ImportSession(path))
        : mockSessionBody(mockReadFile(path))

    return fromExportedSession(body)
}

export async function importSession(
    session: Session,
    locations: SessionLocations,
): Promise<Session> {
    if (!locations.workingDir.trim()) throw new Error('A session needs a working directory.')
    if (!locations.contextDir.trim()) throw new Error('A session needs a context directory.')

    await hydrate()

    const next = {
        ...session,
        workingDir: locations.workingDir.trim(),
        contextDir: locations.contextDir.trim(),
    }

    setSessions([next, ...sessions])
    await saveDraft(next)

    return structuredClone(next)
}
