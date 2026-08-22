/**
 * The step-graph seam. Workflows are authored in this module's memory: inside
 * the Wails webview they are hydrated from the stored drafts and running one
 * starts the real run over the bindings; outside the webview (the plain vite
 * dev server) it falls back to a simulated run. Locking only locks the
 * graph — the run waits for an explicit start, and inputs stay editable until
 * it happens.
 */

import {bridge, hasWailsRuntime} from '@/shared/api/bridge'
import {StepState} from '@/shared/lib/enums'
import {t} from '@/shared/lib/i18n'
import {
    createWorkflow as buildWorkflow,
    createStep as buildStep,
    cancelRun,
    copyStep,
    createsCycle,
    duplicateWorkflow as duplicateGraph,
    moveWorkflow,
    isCancellable,
    isPausable,
    label,
    pauseRun,
    resumeRun,
    withDependency,
    withStep,
    withStepPatch,
    withoutDependency,
    withoutStep,
} from '@/features/workflows/graph'
import {listRoles} from '@/features/roles/api'
import type {InputValue} from '@/features/roles/types'
import type {Point, Workflow, WorkflowDraft, Step, StepDraft} from '@/features/workflows/types'
import {
    findOpenWorkflow,
    findWorkflow,
    findStep,
    hydrate,
    prependWorkflow,
    replaceWorkflow,
    saveDraft,
    workflows,
    setWorkflows,
} from '@/features/workflows/api/store'
import {forgetWorkflowActivity} from '@/features/workflows/api/activity'
import {forgetMockApprovals} from '@/features/approvals/mock-approvals'
import {resolveAcceptance, stopRun, tick} from '@/features/workflows/api/simulated-run'
import {
    answerRemoteAcceptance,
    cancelRemoteRun,
    pauseRemoteRun,
    runs,
    startRemoteRun,
} from '@/features/workflows/api/remote-run'
import {DeleteWorkflowDraft} from '@wailsjs/go/wails_api/API'

export {stepActivity} from '@/features/workflows/api/activity'
export {fetchStepDiff, revertWorkflowTo} from '@/features/workflows/api/step-diff'
export {exportWorkflow, importWorkflow, readWorkflowFile} from '@/features/workflows/api/archive'

export async function listWorkflows(): Promise<Workflow[]> {
    await hydrate()
    return structuredClone(workflows)
}

export async function reorderWorkflow(workflowId: string, toIndex: number): Promise<Workflow[]> {
    await hydrate()

    const previous = workflows
    const next = moveWorkflow(previous, workflowId, toIndex)
    const ranks = new Map(previous.map((workflow) => [workflow.id, workflow.railRank]))

    setWorkflows(next)

    for (const workflow of next.filter((workflow) => ranks.get(workflow.id) !== workflow.railRank))
        await saveDraft(workflow)

    return structuredClone(next)
}

export async function createWorkflow(workflowId: string, draft: WorkflowDraft): Promise<Workflow> {
    if (!draft.projectDir.trim()) throw new Error(t('workflow.api.needsProjectDir'))

    await hydrate()

    const workflow = prependWorkflow({...buildWorkflow(draft), id: workflowId})
    await saveDraft(workflow)

    return structuredClone(workflow)
}

export async function duplicateWorkflow(
    sourceId: string,
    workflowId: string,
    copyInputs: boolean,
): Promise<Workflow> {
    await hydrate()

    const copy = prependWorkflow({
        ...duplicateGraph(findWorkflow(sourceId), copyInputs),
        id: workflowId,
    })
    await saveDraft(copy)

    return structuredClone(copy)
}

export async function updateWorkflow(
    workflowId: string,
    patch: Partial<Pick<Workflow, 'name' | 'locked' | 'started' | 'projectDir'>>,
): Promise<Workflow> {
    if (patch.locked === false) throw new Error(t('workflow.api.noUnlock'))

    if (patch.projectDir !== undefined && !patch.projectDir.trim())
        throw new Error(t('workflow.api.needsProjectDir'))

    await hydrate()

    const locking = patch.locked || patch.started
    const workflow = locking ? findWorkflow(workflowId) : findOpenWorkflow(workflowId)

    if (patch.locked && !workflow.projectDir.trim())
        throw new Error(t('workflow.api.folderBeforeLock'))

    if (patch.started && !workflow.locked) throw new Error(t('workflow.api.lockBeforeRun'))

    if (patch.started && workflow.cancelled) throw new Error(t('workflow.api.cancelled'))

    if (!patch.started) {
        replaceWorkflow({...workflow, ...patch})
        return structuredClone(findWorkflow(workflowId))
    }

    await listRoles()
    await startRun(label({...workflow, ...patch}))

    return structuredClone(findWorkflow(workflowId))
}

async function startRun(workflow: Workflow) {
    if (!hasWailsRuntime()) {
        replaceWorkflow(workflow)
        tick(workflow.id)
        return
    }

    await startRemoteRun(replaceWorkflow(workflow))
}

/** Inputs are editable on the step until the run starts, locked or not. */
export async function setStepInputs(
    workflowId: string,
    stepId: string,
    values: Record<string, InputValue>,
): Promise<Step> {
    await hydrate()

    const workflow = findWorkflow(workflowId)
    if (workflow.started) throw new Error(t('workflow.api.inputsLocked'))

    const step = findStep(workflow, stepId)
    replaceWorkflow(withStepPatch(workflow, stepId, {values: {...step.values, ...values}}))

    return structuredClone(findStep(findWorkflow(workflowId), stepId))
}

/** Answers the gate a finished step is holding: confirm releases downstream, reject fails the step. */
export async function answerStepAcceptance(
    workflowId: string,
    stepId: string,
    accepted: boolean,
): Promise<Workflow> {
    await hydrate()

    const workflow = findWorkflow(workflowId)
    const step = findStep(workflow, stepId)

    if (step.state !== StepState.AwaitingReview)
        throw new Error(t('workflow.api.notAwaitingReview'))

    if (hasWailsRuntime()) await answerRemoteAcceptance(workflowId, stepId, accepted)
    else resolveAcceptance(workflowId, stepId, accepted)

    return structuredClone(findWorkflow(workflowId))
}

/** Pause is not terminal — the ids of the backend run are kept so resuming reattaches to it. */
export async function pauseWorkflow(workflowId: string): Promise<Workflow> {
    return haltWorkflow(workflowId, isPausable, pauseRemoteRun, pauseRun)
}

export async function resumeWorkflow(workflowId: string): Promise<Workflow> {
    await hydrate()

    const workflow = findWorkflow(workflowId)
    if (!workflow.paused) throw new Error(t('workflow.api.notPaused'))

    await listRoles()
    await startRun(resumeRun(workflow))

    return structuredClone(findWorkflow(workflowId))
}

/** A held step loses its question when its workflow goes away, or the queue outlives the run. */
function forgetWorkflowApprovals(workflow: Workflow) {
    forgetMockApprovals(
        workflow.steps.map((step) => step.agentId).filter((id): id is string => Boolean(id)),
    )
}

/** Cancel is terminal — the step that was running loses its work and the workflow can only be duplicated. */
export async function cancelWorkflow(workflowId: string): Promise<Workflow> {
    return haltWorkflow(workflowId, isCancellable, cancelRemoteRun, cancelRun)
}

async function haltWorkflow(
    workflowId: string,
    halts: (workflow: Workflow) => boolean,
    stopRemoteRun: (workflowId: string) => Promise<void>,
    halt: (workflow: Workflow) => Workflow,
): Promise<Workflow> {
    await hydrate()

    const workflow = findWorkflow(workflowId)
    if (!halts(workflow)) return structuredClone(workflow)

    stopRun(workflowId)
    await stopRemoteRun(workflowId)
    forgetWorkflowActivity(workflow)
    forgetWorkflowApprovals(workflow)

    return replaceWorkflow(halt(workflow))
}

export async function deleteWorkflow(workflowId: string): Promise<void> {
    await hydrate()

    const doomed = workflows.find((workflow) => workflow.id === workflowId)

    stopRun(workflowId)
    await cancelRemoteRun(workflowId).catch(() => {})
    runs.delete(workflowId)

    if (doomed) {
        forgetWorkflowActivity(doomed)
        forgetWorkflowApprovals(doomed)
    }

    setWorkflows(workflows.filter((workflow) => workflow.id !== workflowId))

    if (hasWailsRuntime()) await bridge(() => DeleteWorkflowDraft(workflowId))
}

export async function createStep(
    workflowId: string,
    stepId: string,
    draft: StepDraft,
    position: Point,
): Promise<Step> {
    await hydrate()

    const workflow = findOpenWorkflow(workflowId)
    const step = {...buildStep(draft, position), id: stepId}
    replaceWorkflow(withStep(workflow, step))
    return structuredClone(step)
}

export async function duplicateStep(
    workflowId: string,
    stepId: string,
    copyId: string,
    position: Point,
): Promise<Step> {
    await hydrate()

    const workflow = findOpenWorkflow(workflowId)
    const copy = copyStep(findStep(workflow, stepId), copyId, position)
    replaceWorkflow(withStep(workflow, copy))

    return structuredClone(copy)
}

export async function updateStep(
    workflowId: string,
    stepId: string,
    patch: Partial<Step>,
): Promise<Step> {
    await hydrate()

    const workflow = findOpenWorkflow(workflowId)
    const next = {...findStep(workflow, stepId), ...patch, id: stepId}
    replaceWorkflow(withStepPatch(workflow, stepId, patch))
    return structuredClone(next)
}

/**
 * Where a step sits says nothing about what the run will do, so a locked graph
 * can still be rearranged. Every other step edit goes through findOpenWorkflow.
 */
export async function moveStep(workflowId: string, stepId: string, position: Point): Promise<Step> {
    await hydrate()

    const workflow = findWorkflow(workflowId)
    const next = {...findStep(workflow, stepId), position}
    replaceWorkflow(withStepPatch(workflow, stepId, {position}))
    return structuredClone(next)
}

export async function deleteStep(workflowId: string, stepId: string): Promise<void> {
    await hydrate()
    replaceWorkflow(withoutStep(findOpenWorkflow(workflowId), stepId))
}

export async function addDependency(
    workflowId: string,
    sourceId: string,
    targetId: string,
): Promise<void> {
    await hydrate()

    const workflow = findOpenWorkflow(workflowId)
    if (createsCycle(workflow.steps, sourceId, targetId)) throw new Error(t('workflow.api.cycle'))

    replaceWorkflow(withDependency(workflow, sourceId, targetId))
}

export async function removeDependency(
    workflowId: string,
    sourceId: string,
    targetId: string,
): Promise<void> {
    await hydrate()
    replaceWorkflow(withoutDependency(findOpenWorkflow(workflowId), sourceId, targetId))
}
