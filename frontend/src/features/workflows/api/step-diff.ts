/**
 * What a step changed is not carried on its report — it is read from the
 * workflow's snapshot when the step is opened, and reverting to a step restores
 * the working directory to the snapshot that step left behind. Reverting stops
 * the run, so the graph is rewound here too and the operator starts it again
 * when they are ready — the same backend workflow, resumed. Under the plain vite
 * dev server both are answered from features/workflows/mock-step-diff.ts.
 */

import {bridge, hasWailsRuntime} from '@/shared/api/bridge'
import {FileChangeType} from '@/shared/lib/enums'
import {rewindTo} from '@/features/workflows/graph'
import {mockStepDiff} from '@/features/workflows/mock-step-diff'
import type {FileChange} from '@/features/workflows/types'
import {RemoteChangeType} from '@/features/workflows/api/remote-enums'
import {forgetWorkflowActivity} from '@/features/workflows/api/activity'
import {remoteRefs, runs} from '@/features/workflows/api/remote-run'
import {stopRun} from '@/features/workflows/api/simulated-run'
import {findWorkflow, replaceWorkflow} from '@/features/workflows/api/store'
import type {output_itf} from '@wailsjs/go/models'
import {RevertWorkflowTo, StepDiff} from '@wailsjs/go/wails_api/API'

const ROUNDTRIP_MS = 400

async function roundtrip() {
    await new Promise((resolve) => setTimeout(resolve, ROUNDTRIP_MS))
}

export async function fetchStepDiff(workflowId: string, stepId: string): Promise<FileChange[]> {
    if (!hasWailsRuntime()) {
        await roundtrip()
        return mockStepDiff(stepId)
    }

    const refs = remoteRefs(workflowId, stepId)
    const infos = await bridge(() => StepDiff(refs.workflowId, refs.stepId))

    return infos.map(toFileChange)
}

export async function revertWorkflowTo(workflowId: string, stepId: string): Promise<void> {
    if (hasWailsRuntime()) {
        const refs = remoteRefs(workflowId, stepId)
        await bridge(() => RevertWorkflowTo(refs.workflowId, refs.stepId))
        runs.delete(workflowId)
    } else {
        await roundtrip()
    }

    stopRun(workflowId)

    const workflow = findWorkflow(workflowId)
    forgetWorkflowActivity(workflow)
    replaceWorkflow(rewindTo(workflow, stepId))
}

const REMOTE_CHANGE_TYPES: Record<RemoteChangeType, FileChangeType> = {
    [RemoteChangeType.Added]: FileChangeType.Created,
    [RemoteChangeType.Modified]: FileChangeType.Modified,
    [RemoteChangeType.Deleted]: FileChangeType.Deleted,
    [RemoteChangeType.Renamed]: FileChangeType.Renamed,
}

function toFileChange(info: output_itf.FileChangeInfo): FileChange {
    return {
        path: info.path ?? '',
        oldPath: info.old_path ?? '',
        changeType:
            REMOTE_CHANGE_TYPES[info.change_type as RemoteChangeType] ?? FileChangeType.Modified,
        additions: info.additions ?? 0,
        deletions: info.deletions ?? 0,
        unifiedDiff: info.unified_diff ?? '',
    }
}
