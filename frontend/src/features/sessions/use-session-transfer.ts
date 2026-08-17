/**
 * Import and export for the sessions rail: picking the path, holding the
 * session a file answered while its directories are asked for, and the notice
 * that says what happened. Mirrors useTemplateTransfer.
 */

import {useState} from 'react'
import {useMutation} from '@tanstack/react-query'

import * as api from '@/features/sessions/api'
import {chooseFile, chooseSaveFile} from '@/shared/api/dialogs'
import {reportError} from '@/shared/lib/error-bus'
import type {Session, SessionLocations} from '@/features/sessions/types'

export type TransferNotice = {title: string; description: string; detail: string}

const JSON_FILES = '*.json'

function exportFileName() {
    return `nexo-session-${new Date().toISOString().slice(0, 10)}.json`
}

export function useSessionTransfer(
    onImport: (session: Session, locations: SessionLocations) => void,
) {
    const [pending, setPending] = useState<Session | null>(null)
    const [notice, setNotice] = useState<TransferNotice | null>(null)

    const read = useMutation({
        meta: {action: 'Could not read that file'},
        mutationFn: (path: string) => api.readSessionFile(path),
    })

    const write = useMutation({
        meta: {action: 'Could not export the session'},
        mutationFn: ({sessionId, path}: {sessionId: string; path: string}) =>
            api.exportSession(sessionId, path),
    })

    const pickPath = async (pick: () => Promise<string>) => {
        try {
            return await pick()
        } catch (cause) {
            reportError(cause, 'Could not open the file picker')
            return ''
        }
    }

    const beginImport = async () => {
        const path = await pickPath(() => chooseFile('Import session', JSON_FILES))
        if (!path) return

        const session = await read.mutateAsync(path).catch(() => null)
        if (session) setPending(session)
    }

    const cancelImport = () => setPending(null)

    const confirmImport = (locations: SessionLocations) => {
        if (!pending) return

        onImport(pending, locations)
        setPending(null)
        setNotice({
            title: `Imported ${pending.name}`,
            description: 'It is a draft: nothing runs until you finalize and start it.',
            detail: locations.workingDir,
        })
    }

    const exportSession = async (sessionId: string) => {
        const path = await pickPath(() =>
            chooseSaveFile('Export session', exportFileName(), JSON_FILES),
        )
        if (!path) return

        const name = await write.mutateAsync({sessionId, path}).catch(() => null)
        if (name === null) return

        setNotice({
            title: `Exported ${name}`,
            description:
                'The file carries the graph and what each node runs as. Templates stay here.',
            detail: path,
        })
    }

    const dismissNotice = () => setNotice(null)

    return {
        pending,
        beginImport,
        cancelImport,
        confirmImport,
        exportSession,
        reading: read.isPending,
        writing: write.isPending,
        notice,
        dismissNotice,
    }
}
