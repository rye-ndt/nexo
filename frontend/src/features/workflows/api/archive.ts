/**
 * The workflow-transfer seam. Inside the Wails webview Go owns the file and its
 * version envelope; outside it the same envelope is written into the stand-in
 * filesystem, so a round trip is testable on port 8888.
 *
 * Reading a file writes nothing: it answers a candidate workflow, and the
 * directories it will run in are asked for before it lands in the rail.
 */

import {bridge, hasWailsRuntime} from '@/shared/api/bridge'
import {t} from '@/shared/lib/i18n'
import {fromExportedWorkflow, toExportedWorkflow} from '@/features/workflows/workflow-file'
import {listRoles} from '@/features/roles/api'
import {mockWorkflowArchive, mockWorkflowBody} from '@/features/workflows/mock-workflows'
import {mockReadFile, mockWriteFile} from '@/shared/api/mock-fs'
import type {Workflow, WorkflowLocations} from '@/features/workflows/types'
import {findWorkflow, hydrate, prependWorkflow, saveDraft} from '@/features/workflows/api/store'
import {ExportWorkflow, ImportWorkflow} from '@wailsjs/go/wails_api/API'

type WorkflowFiles = {
    write(path: string, body: string): Promise<void>
    read(path: string): Promise<string>
}

const nativeFiles: WorkflowFiles = {
    write: async (path, body) => {
        await bridge(() => ExportWorkflow(path, body))
    },
    read: async (path) => bridge(() => ImportWorkflow(path)),
}

const mockFiles: WorkflowFiles = {
    write: async (path, body) => {
        mockWriteFile(path, mockWorkflowArchive(body))
    },
    read: async (path) => mockWorkflowBody(mockReadFile(path)),
}

const files: WorkflowFiles = hasWailsRuntime() ? nativeFiles : mockFiles

export async function exportWorkflow(workflowId: string, path: string): Promise<string> {
    await hydrate()

    const workflow = findWorkflow(workflowId)
    await files.write(path, JSON.stringify(toExportedWorkflow(workflow, await listRoles())))

    return workflow.name
}

export async function readWorkflowFile(path: string): Promise<Workflow> {
    return fromExportedWorkflow(await files.read(path))
}

export async function importWorkflow(
    workflow: Workflow,
    locations: WorkflowLocations,
): Promise<Workflow> {
    if (!locations.projectDir.trim()) throw new Error(t('workflow.api.needsProjectDir'))

    await hydrate()

    const next = {
        ...workflow,
        projectDir: locations.projectDir.trim(),
    }

    const ranked = prependWorkflow(next)
    await saveDraft(ranked)

    return structuredClone(ranked)
}
