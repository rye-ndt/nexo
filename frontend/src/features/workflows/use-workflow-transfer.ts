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
import {t} from '@/shared/lib/i18n'
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
        meta: {action: t('workflow.error.readFile')},
        mutationFn: (path: string) => api.readWorkflowFile(path),
    })

    const write = useMutation({
        meta: {action: t('workflow.error.export')},
        mutationFn: ({workflowId, path}: {workflowId: string; path: string}) =>
            api.exportWorkflow(workflowId, path),
    })

    const pickPath = async (pick: () => Promise<string>) => {
        try {
            return await pick()
        } catch (cause) {
            reportError(cause, t('workflow.error.filePicker'))
            return ''
        }
    }

    const beginImport = async () => {
        const path = await pickPath(() =>
            chooseFile(t('workflow.transfer.importPicker'), JSON_FILES),
        )
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
            title: t('workflow.transfer.imported.title', {name: pending.name}),
            description: t('workflow.transfer.imported.description'),
            detail: locations.projectDir,
        })
    }

    const exportWorkflow = async (workflowId: string) => {
        const path = await pickPath(() =>
            chooseSaveFile(t('workflow.transfer.exportPicker'), exportFileName(), JSON_FILES),
        )
        if (!path) return

        const name = await write.mutateAsync({workflowId, path}).catch(() => null)
        if (name === null) return

        setNotice({
            title: t('workflow.transfer.exported.title', {name}),
            description: t('workflow.transfer.exported.description'),
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
