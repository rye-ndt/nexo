/**
 * What a step changed is not carried on its report — it is read from the
 * session's snapshot when the step is opened, and reverting to a step restores
 * the working directory to the snapshot that step left behind. Reverting stops
 * the run, so the graph is rewound here too and the operator starts it again
 * when they are ready — the same backend session, resumed. Under the plain vite
 * dev server both are answered from features/sessions/mock-task-diff.ts.
 */

import {bridge, hasWailsRuntime} from '@/shared/api/bridge'
import {FileChangeType} from '@/shared/lib/enums'
import {rewindTo} from '@/features/sessions/graph'
import {mockTaskDiff} from '@/features/sessions/mock-task-diff'
import type {FileChange} from '@/features/sessions/types'
import {RemoteChangeType} from '@/features/sessions/api/remote-enums'
import {forgetSessionActivity} from '@/features/sessions/api/activity'
import {remoteRefs, runs} from '@/features/sessions/api/remote-run'
import {stopRun} from '@/features/sessions/api/simulated-run'
import {findSession, replaceSession} from '@/features/sessions/api/store'
import type {output_itf} from '@wailsjs/go/models'
import {RevertSessionTo, TaskDiff} from '@wailsjs/go/wails_api/API'

const ROUNDTRIP_MS = 400

async function roundtrip() {
    await new Promise((resolve) => setTimeout(resolve, ROUNDTRIP_MS))
}

export async function fetchTaskDiff(sessionId: string, taskId: string): Promise<FileChange[]> {
    if (!hasWailsRuntime()) {
        await roundtrip()
        return mockTaskDiff(taskId)
    }

    const refs = remoteRefs(sessionId, taskId)
    const infos = await bridge(() => TaskDiff(refs.sessionId, refs.taskId))

    return infos.map(toFileChange)
}

export async function revertSessionTo(sessionId: string, taskId: string): Promise<void> {
    if (hasWailsRuntime()) {
        const refs = remoteRefs(sessionId, taskId)
        await bridge(() => RevertSessionTo(refs.sessionId, refs.taskId))
        runs.delete(sessionId)
    } else {
        await roundtrip()
    }

    stopRun(sessionId)

    const session = findSession(sessionId)
    forgetSessionActivity(session)
    replaceSession(rewindTo(session, taskId))
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
