/**
 * Import and export for the workflows rail: picking the path, holding the
 * workflow a file answered while its directories are asked for, and the notice
 * that says what happened. Mirrors useRoleTransfer.
 */

import {useState} from 'react'
import {useMutation} from '@tanstack/react-query'

import * as api from '@/features/workflows/api'
import {chooseFile, chooseSaveFile} from '@/shared/api/dialogs'
import {reportError} from '@/shared/lib/error-bus'
import type {Workflow, WorkflowLocations} from '@/features/workflows/types'

export type TransferNotice = {title: string; description: string; detail: string}

const JSON_FILES = '*.json'

function exportFileName() {
    return `nexo-workflow-${new Date().toISOString().slice(0, 10)}.json`
}

export function useWorkflowTransfer(
    onImport: (workflow: Workflow, locations: WorkflowLocations) => void,
) {
    const [pending, setPending] = useState<Workflow | null>(null)
    const [notice, setNotice] = useState<TransferNotice | null>(null)

    const read = useMutation({
        meta: {action: 'Could not read that file'},
        mutationFn: (path: string) => api.readWorkflowFile(path),
    })

    const write = useMutation({
        meta: {action: 'Could not export the workflow'},
        mutationFn: ({workflowId, path}: {workflowId: string; path: string}) =>
            api.exportWorkflow(workflowId, path),
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
        const path = await pickPath(() => chooseFile('Import workflow', JSON_FILES))
        if (!path) return

        const workflow = await read.mutateAsync(path).catch(() => null)
        if (workflow) setPending(workflow)
    }

    const cancelImport = () => setPending(null)

    const confirmImport = (locations: WorkflowLocations) => {
        if (!pending) return

        onImport(pending, locations)
        setPending(null)
        setNotice({
            title: `Imported ${pending.name}`,
            description: 'It is a draft: nothing runs until you lock and start it.',
            detail: locations.projectDir,
        })
    }

    const exportWorkflow = async (workflowId: string) => {
        const path = await pickPath(() =>
            chooseSaveFile('Export workflow', exportFileName(), JSON_FILES),
        )
        if (!path) return

        const name = await write.mutateAsync({workflowId, path}).catch(() => null)
        if (name === null) return

        setNotice({
            title: `Exported ${name}`,
            description: 'The file carries the graph and what each step runs as. Roles stay here.',
            detail: path,
        })
    }

    const dismissNotice = () => setNotice(null)

    return {
        pending,
        beginImport,
        cancelImport,
        confirmImport,
        exportWorkflow,
        reading: read.isPending,
        writing: write.isPending,
        notice,
        dismissNotice,
    }
}
