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

export async function exportWorkflow(workflowId: string, path: string): Promise<string> {
    await hydrate()

    const workflow = findWorkflow(workflowId)
    const body = JSON.stringify(toExportedWorkflow(workflow, await listRoles()))

    if (hasWailsRuntime()) await bridge(() => ExportWorkflow(path, body))
    else mockWriteFile(path, mockWorkflowArchive(body))

    return workflow.name
}

export async function readWorkflowFile(path: string): Promise<Workflow> {
    const body = hasWailsRuntime()
        ? await bridge(() => ImportWorkflow(path))
        : mockWorkflowBody(mockReadFile(path))

    return fromExportedWorkflow(body)
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
