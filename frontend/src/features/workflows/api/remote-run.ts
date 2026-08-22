/**
 * Locking builds a RunWorkflowSpec, starts the real run through the
 * generated bindings, and polls WorkflowStatus until the run settles. A step
 * that pauses for review halts in awaiting_review until the answer
 * goes back over the same bindings.
 *
 * A workflow the Go side already knows is resumed rather than rebuilt. The ids it
 * handed out ride on the workflow itself, so they outlive the page: reverting
 * rewinds that workflow in place, and a run halted by a pause or a restart
 * reattaches instead of starting a second workflow over the same directory.
 */

import {bridge} from '@/shared/api/bridge'
import {listRoles} from '@/features/roles/api'
import {StepState, WorkflowLifecycle} from '@/shared/lib/enums'
import {t} from '@/shared/lib/i18n'
import {hasActiveStep, isFinished, label} from '@/features/workflows/graph'
import {resolvedPrompt} from '@/features/workflows/step-inputs'
import {specOf} from '@/features/workflows/step-spec'
import type {
    ActivityLine,
    ContextUsage,
    Handoff,
    RemoteRunIds,
    Workflow,
    Spend,
    Step,
} from '@/features/workflows/types'
import {RemoteStepStatus} from '@/features/workflows/api/remote-enums'
import {findWorkflow, replaceWorkflow, workflows} from '@/features/workflows/api/store'
import {mergeActivity} from '@/features/workflows/api/activity'
import {TICK_MS, timers} from '@/features/workflows/api/timers'
import {output_itf} from '@wailsjs/go/models'
import type {core_itf} from '@wailsjs/go/models'
import type {input_itf} from '@wailsjs/go/models'
import {
    AnswerStepReview,
    CancelWorkflow,
    DiscardWorkflowRun,
    PauseWorkflow,
    ResumeWorkflow,
    RunWorkflow,
    WorkflowStatus,
} from '@wailsjs/go/wails_api/API'

const MAX_FAILED_POLLS = 5

type RemoteRun = {
    remoteWorkflowId: string
    clientStepIds: Map<string, string>
    failedPolls: number
}

export const runs = new Map<string, RemoteRun>()

export function remoteRefs(workflowId: string, stepId: string) {
    const remote = findWorkflow(workflowId).remote
    const remoteStepId = remote?.stepIds[stepId]

    if (!remote || !remoteStepId) throw new Error(t('workflow.api.stepNotRun'))

    return {workflowId: remote.workflowId, stepId: remoteStepId}
}

async function buildRunSpec(workflow: Workflow): Promise<output_itf.RunWorkflowSpec> {
    const roles = await listRoles()

    return new output_itf.RunWorkflowSpec({
        project_dir_path: workflow.projectDir.trim(),
        steps: workflow.steps.map((step) => {
            const spec = specOf(step, roles)
            return {
                client_id: step.id,
                name: step.title,
                prompt: resolvedPrompt(step),
                effort: spec.effort,
                instructions: spec.instructions,
                output_structure: spec.outputStructure,
                depends_on: [...step.dependsOn],
                auto_retry: false,
                pause_for_review: spec.pauseForReview,
            }
        }),
    })
}

function track(workflow: Workflow, remote: RemoteRunIds) {
    const clientStepIds = new Map(
        Object.entries(remote.stepIds).map(([clientId, remoteId]) => [remoteId, clientId] as const),
    )

    runs.set(workflow.id, {remoteWorkflowId: remote.workflowId, clientStepIds, failedPolls: 0})
    replaceWorkflow({...workflow, remote})

    pollRemoteRun(workflow.id)
}

export async function startRemoteRun(workflow: Workflow): Promise<void> {
    const started = replaceWorkflow(workflow)

    if (runs.has(started.id)) return
    if (!started.projectDir.trim()) throw new Error(t('workflow.api.setProjectDir'))

    const known = started.remote

    if (known) {
        await bridge(() => ResumeWorkflow(known.workflowId))
        track(started, known)
        return
    }

    const spec = await buildRunSpec(started)
    const result = await bridge(() => RunWorkflow(spec))

    if (!result.workflow_id) throw new Error(t('workflow.api.noWorkflowId'))

    track(started, {workflowId: result.workflow_id, stepIds: result.step_ids ?? {}})
}

export async function answerRemoteAcceptance(
    workflowId: string,
    stepId: string,
    accepted: boolean,
): Promise<void> {
    const refs = remoteRefs(workflowId, stepId)

    await bridge(() => AnswerStepReview(refs.stepId, accepted))
    pollRemoteRun(workflowId)
}

export async function pauseRemoteRun(workflow: Workflow): Promise<void> {
    return haltRemoteRun(workflow.id, PauseWorkflow)
}

export async function cancelRemoteRun(workflow: Workflow): Promise<void> {
    return haltRemoteRun(workflow.id, CancelWorkflow)
}

async function haltRemoteRun(
    workflowId: string,
    halt: (remoteWorkflowId: string) => Promise<void>,
) {
    const run = runs.get(workflowId)
    if (!run) return

    await bridge(() => halt(run.remoteWorkflowId))

    stopPolling(workflowId)
    runs.delete(workflowId)
}

export async function discardRemoteRun(workflow: Workflow): Promise<void> {
    stopPolling(workflow.id)
    runs.delete(workflow.id)

    const remote = workflow.remote
    if (remote) await bridge(() => DiscardWorkflowRun(remote.workflowId))
}

export function stopPolling(workflowId: string) {
    const timer = timers.get(workflowId)
    if (timer === undefined) return

    clearTimeout(timer)
    timers.delete(workflowId)
}

function pollRemoteRun(workflowId: string) {
    if (timers.has(workflowId)) return

    timers.set(
        workflowId,
        setTimeout(() => {
            timers.delete(workflowId)
            void pollRemoteTick(workflowId)
        }, TICK_MS),
    )
}

async function pollRemoteTick(workflowId: string) {
    const run = runs.get(workflowId)
    const workflow = workflows.find((workflow) => workflow.id === workflowId)
    if (!run || !workflow || workflow.lifecycle === WorkflowLifecycle.Cancelled) return

    let status: output_itf.WorkflowStatusInfo | null

    try {
        status = await bridge(() => WorkflowStatus(run.remoteWorkflowId))
    } catch {
        status = null
    }

    const afterPoll = workflows.find((workflow) => workflow.id === workflowId)
    if (!afterPoll) return

    if (!status) {
        run.failedPolls += 1

        if (run.failedPolls >= MAX_FAILED_POLLS) {
            runs.delete(workflowId)
            replaceWorkflow(loseRun(afterPoll))
            return
        }
    }

    if (status) {
        run.failedPolls = 0

        const next = replaceWorkflow(applyRemoteStatus(afterPoll, run, status))
        if (!hasActiveStep(next)) {
            runs.delete(workflowId)
            return
        }
    }

    pollRemoteRun(workflowId)
}

function loseRun(workflow: Workflow): Workflow {
    return {
        ...workflow,
        steps: workflow.steps.map((step) =>
            isFinished(step.state) ? step : {...step, state: StepState.Failed},
        ),
    }
}

const REMOTE_STATES: Record<RemoteStepStatus, StepState> = {
    [RemoteStepStatus.Processing]: StepState.Running,
    [RemoteStepStatus.AwaitingReview]: StepState.AwaitingReview,
    [RemoteStepStatus.Completed]: StepState.Done,
    [RemoteStepStatus.Failed]: StepState.Failed,
    [RemoteStepStatus.Cancelled]: StepState.Failed,
}

function applyRemoteStatus(
    workflow: Workflow,
    run: RemoteRun,
    status: output_itf.WorkflowStatusInfo,
): Workflow {
    const now = new Date().toISOString()
    const infoByClientId = new Map(
        (status.steps ?? []).map((info) => [run.clientStepIds.get(info.step_id), info]),
    )

    return label({
        ...workflow,
        spent: {
            input: status.tokens_input ?? 0,
            cached: status.tokens_cached ?? 0,
            output: status.tokens_billed ?? 0,
        },
        costUsd: status.cost_usd ?? 0,
        priced: status.priced ?? false,
        startedAt: status.started_at || workflow.startedAt,
        finishedAt: status.completed_at || undefined,
        steps: workflow.steps.map((step) => {
            const info = infoByClientId.get(step.id)
            return info ? applyRemoteStep(step, info, now) : step
        }),
    })
}

function applyRemoteStep(step: Step, info: output_itf.WorkflowStepInfo, now: string): Step {
    mergeActivity(step.id, (info.activity ?? []).map(toActivityLine))

    const state = REMOTE_STATES[info.status as RemoteStepStatus]
    if (!state) return step

    const startedAt = step.run?.startedAt ?? now
    const finishedAt = state === StepState.Running ? undefined : (step.run?.finishedAt ?? now)
    const context = toContextUsage(info.context_usage) ?? step.run?.context
    const spent = toSpend(info.spent) ?? step.run?.spent

    return {
        ...step,
        state,
        agentId: info.agent_id || step.agentId,
        run: {
            ...step.run,
            startedAt,
            finishedAt,
            context,
            spent,
            costUsd: info.cost_usd ?? 0,
            priced: info.priced ?? false,
        },
        report:
            state === StepState.Running
                ? step.report
                : {
                      status: state,
                      handoffs: (info.handoffs ?? []).map(toHandoff),
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

function toActivityLine(info: output_itf.StepActivityInfo): ActivityLine {
    return {seq: info.seq ?? 0, at: info.at ?? '', text: info.text ?? ''}
}

function toHandoff(info: core_itf.Handoff): Handoff {
    return {
        step: info.step ?? '',
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
