/**
 * Finalizing builds a RunSessionSpec, starts the real run through the
 * generated bindings, and polls SessionStatus until the run settles. A node
 * that asks for manual acceptance halts in awaiting_accept until the answer
 * goes back over the same bindings.
 *
 * A session the Go side already knows is resumed rather than rebuilt. The ids it
 * handed out ride on the session itself, so they outlive the page: reverting
 * rewinds that session in place, and a run halted by a pause or a restart
 * reattaches instead of starting a second session over the same directory.
 */

import {bridge, hasWailsRuntime} from '@/shared/api/bridge'
import {listTemplates} from '@/features/templates/api'
import {TaskState} from '@/shared/lib/enums'
import {hasActiveTask, isFinished, label} from '@/features/sessions/graph'
import {resolvedPrompt} from '@/features/sessions/task-inputs'
import {specOf} from '@/features/sessions/task-spec'
import type {
    ActivityLine,
    ContextUsage,
    HandoverDoc,
    RemoteRunIds,
    Session,
    Spend,
    Task,
} from '@/features/sessions/types'
import {RemoteTaskStatus} from '@/features/sessions/api/remote-enums'
import {findSession, replaceSession, sessions} from '@/features/sessions/api/store'
import {mergeActivity} from '@/features/sessions/api/activity'
import {TICK_MS, timers} from '@/features/sessions/api/timers'
import {output_itf} from '@wailsjs/go/models'
import type {input_itf} from '@wailsjs/go/models'
import {
    AnswerTaskAcceptance,
    CancelSession,
    PauseSession,
    ResumeSession,
    RunSession,
    SessionStatus,
} from '@wailsjs/go/wails_api/API'

const MAX_FAILED_POLLS = 5

type RemoteRun = {
    remoteSessionId: string
    clientTaskIds: Map<string, string>
    failedPolls: number
}

export const runs = new Map<string, RemoteRun>()

export function remoteRefs(sessionId: string, taskId: string) {
    const remote = findSession(sessionId).remote
    const remoteTaskId = remote?.taskIds[taskId]

    if (!remote || !remoteTaskId)
        throw new Error('This node has not run yet, so there is nothing to inspect.')

    return {sessionId: remote.sessionId, taskId: remoteTaskId}
}

async function buildRunSpec(session: Session): Promise<output_itf.RunSessionSpec> {
    const templates = await listTemplates()

    return new output_itf.RunSessionSpec({
        working_dir_path: session.workingDir.trim(),
        context_dir_path: session.contextDir.trim(),
        tasks: session.tasks.map((task) => {
            const spec = specOf(task, templates)
            return {
                client_id: task.id,
                name: task.title,
                prompt: resolvedPrompt(task),
                task_level: spec.taskLevel,
                system_prompts: spec.systemPrompts,
                output_structure: spec.outputStructure,
                depends_on: [...task.dependsOn],
                auto_retry: false,
                manual_accept_required: spec.manualAcceptRequired,
            }
        }),
    })
}

function track(session: Session, remote: RemoteRunIds) {
    const clientTaskIds = new Map(
        Object.entries(remote.taskIds).map(([clientId, remoteId]) => [remoteId, clientId] as const),
    )

    runs.set(session.id, {remoteSessionId: remote.sessionId, clientTaskIds, failedPolls: 0})
    replaceSession({...session, remote})

    pollRemoteRun(session.id)
}

export async function startRemoteRun(session: Session): Promise<boolean> {
    if (!hasWailsRuntime()) return false
    if (runs.has(session.id)) return true
    if (!session.workingDir.trim())
        throw new Error('Set a working directory before running this session.')

    const known = session.remote

    if (known) {
        await bridge(() => ResumeSession(known.sessionId))
        track(session, known)
        return true
    }

    const spec = await buildRunSpec(session)
    const result = await bridge(() => RunSession(spec))

    if (!result.session_id)
        throw new Error('The run started without a session id. Check the app log and try again.')

    track(session, {sessionId: result.session_id, taskIds: result.task_ids ?? {}})
    return true
}

export async function answerRemoteAcceptance(
    sessionId: string,
    taskId: string,
    accepted: boolean,
): Promise<void> {
    const refs = remoteRefs(sessionId, taskId)

    await bridge(() => AnswerTaskAcceptance(refs.taskId, accepted))
    pollRemoteRun(sessionId)
}

export async function pauseRemoteRun(sessionId: string): Promise<void> {
    return haltRemoteRun(sessionId, PauseSession)
}

export async function cancelRemoteRun(sessionId: string): Promise<void> {
    return haltRemoteRun(sessionId, CancelSession)
}

async function haltRemoteRun(sessionId: string, halt: (remoteSessionId: string) => Promise<void>) {
    const run = runs.get(sessionId)
    if (!run) return

    await bridge(() => halt(run.remoteSessionId))
    runs.delete(sessionId)
}

function pollRemoteRun(sessionId: string) {
    if (timers.has(sessionId)) return

    timers.set(
        sessionId,
        setTimeout(() => {
            timers.delete(sessionId)
            void pollRemoteTick(sessionId)
        }, TICK_MS),
    )
}

async function pollRemoteTick(sessionId: string) {
    const run = runs.get(sessionId)
    const session = sessions.find((session) => session.id === sessionId)
    if (!run || !session || session.cancelled) return

    let status: output_itf.SessionStatusInfo | null

    try {
        status = await bridge(() => SessionStatus(run.remoteSessionId))
    } catch {
        status = null
    }

    const afterPoll = sessions.find((session) => session.id === sessionId)
    if (!afterPoll) return

    if (!status) {
        run.failedPolls += 1

        if (run.failedPolls >= MAX_FAILED_POLLS) {
            runs.delete(sessionId)
            replaceSession(loseRun(afterPoll))
            return
        }
    }

    if (status) {
        run.failedPolls = 0

        const next = replaceSession(applyRemoteStatus(afterPoll, run, status))
        if (!hasActiveTask(next)) {
            runs.delete(sessionId)
            return
        }
    }

    pollRemoteRun(sessionId)
}

function loseRun(session: Session): Session {
    return {
        ...session,
        tasks: session.tasks.map((task) =>
            isFinished(task.state) ? task : {...task, state: TaskState.Failed},
        ),
    }
}

const REMOTE_STATES: Record<RemoteTaskStatus, TaskState> = {
    [RemoteTaskStatus.Processing]: TaskState.Running,
    [RemoteTaskStatus.AwaitingAccept]: TaskState.AwaitingAccept,
    [RemoteTaskStatus.Completed]: TaskState.Done,
    [RemoteTaskStatus.Failed]: TaskState.Failed,
    [RemoteTaskStatus.Cancelled]: TaskState.Failed,
}

function applyRemoteStatus(
    session: Session,
    run: RemoteRun,
    status: output_itf.SessionStatusInfo,
): Session {
    const now = new Date().toISOString()
    const infoByClientId = new Map(
        (status.tasks ?? []).map((info) => [run.clientTaskIds.get(info.task_id), info]),
    )

    return label({
        ...session,
        spent: {
            input: status.tokens_input ?? 0,
            cached: status.tokens_cached ?? 0,
            output: status.tokens_billed ?? 0,
        },
        costUsd: status.cost_usd ?? 0,
        priced: status.priced ?? false,
        startedAt: status.started_at || session.startedAt,
        finishedAt: status.completed_at || undefined,
        tasks: session.tasks.map((task) => {
            const info = infoByClientId.get(task.id)
            return info ? applyRemoteTask(task, info, now) : task
        }),
    })
}

function applyRemoteTask(task: Task, info: output_itf.SessionTaskInfo, now: string): Task {
    mergeActivity(task.id, (info.activity ?? []).map(toActivityLine))

    const state = REMOTE_STATES[info.status as RemoteTaskStatus]
    if (!state) return task

    const startedAt = task.run?.startedAt ?? now
    const finishedAt = state === TaskState.Running ? undefined : (task.run?.finishedAt ?? now)
    const context = toContextUsage(info.context_usage) ?? task.run?.context
    const spent = toSpend(info.spent) ?? task.run?.spent

    return {
        ...task,
        state,
        agentId: info.agent_id || task.agentId,
        run: {
            ...task.run,
            startedAt,
            finishedAt,
            context,
            spent,
            costUsd: info.cost_usd ?? 0,
            priced: info.priced ?? false,
        },
        report:
            state === TaskState.Running
                ? task.report
                : {
                      status: state,
                      handoverDocs: (info.handover_docs ?? []).map(toHandoverDoc),
                  },
    }
}

function toContextUsage(info?: input_itf.ContextUsage): ContextUsage | undefined {
    if (!info || !info.total) return undefined
    return {used: info.used ?? 0, total: info.total}
}

function toSpend(info?: input_itf.ContextUsage): Spend | undefined {
    if (!info) return undefined
    return {input: info.input ?? 0, cached: info.cached ?? 0, output: info.billed ?? 0}
}

function toActivityLine(info: output_itf.TaskActivityInfo): ActivityLine {
    return {seq: info.seq ?? 0, at: info.at ?? '', text: info.text ?? ''}
}

function toHandoverDoc(info: output_itf.HandoverDocInfo): HandoverDoc {
    return {
        task: info.task ?? '',
        tldr: info.tldr ?? '',
        outcome: info.outcome ?? '',
        blockers: info.blockers ?? {},
        approvedDecisions: info.approved_decisions ?? {},
        rejectedDecisions: info.rejected_decisions ?? {},
        currentBehaviors: info.current_behaviors ?? {},
        changedBehaviors: info.changed_behaviors ?? {},
        mustAvoid: info.must_avoid ?? {},
        nuances: info.nuances ?? {},
        knownGaps: info.known_gaps ?? {},
    }
}
