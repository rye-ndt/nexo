import {WorkflowLifecycle, WorkflowStatus, StepState} from '@/shared/lib/enums'
import {t} from '@/shared/lib/i18n'
import type {
    PastRun,
    Point,
    Workflow,
    WorkflowDraft,
    Step,
    StepDraft,
} from '@/features/workflows/types'
import type {DraftContext} from '@/features/roles/types'

const SETTLED_STATES = new Set<StepState>([
    StepState.Running,
    StepState.AwaitingReview,
    StepState.Done,
    StepState.Failed,
    StepState.Cancelled,
])

const FINISHED_STATES = new Set<StepState>([StepState.Done, StepState.Failed, StepState.Cancelled])

const KEPT_ON_CANCEL = new Set<StepState>([StepState.Done, StepState.Failed])

const WAITING_STATES = new Set<StepState>([StepState.Idle, StepState.Queued, StepState.Blocked])

const STARTED_LIFECYCLES = new Set<WorkflowLifecycle>([
    WorkflowLifecycle.Running,
    WorkflowLifecycle.Paused,
    WorkflowLifecycle.Cancelled,
])

const STEP_STAGGER = 40

const MAX_PAST_RUNS = 3

function isSettled(state: StepState) {
    return SETTLED_STATES.has(state)
}

export function isFinished(state: StepState) {
    return FINISHED_STATES.has(state)
}

export function isLocked(workflow: Workflow) {
    return workflow.lifecycle !== WorkflowLifecycle.Draft
}

export function hasStarted(workflow: Workflow) {
    return STARTED_LIFECYCLES.has(workflow.lifecycle)
}

export function createStep(draft: StepDraft, position: Point): Step {
    return {
        id: crypto.randomUUID(),
        title: draft.title,
        prompt: draft.prompt,
        state: StepState.Idle,
        position,
        dependsOn: [],
        roleId: draft.roleId,
        values: draft.values,
    }
}

/**
 * A copied step is the step and nothing around it: its role, prompt and the values
 * typed into it all come across, and it arrives unconnected at both ends. You draw
 * the edges it should have yourself.
 */
export function copyStep(step: Step, id: string, position: Point): Step {
    return {
        id,
        title: step.title ? `${step.title} copy` : '',
        prompt: step.prompt,
        state: StepState.Idle,
        position,
        dependsOn: [],
        roleId: step.roleId,
        spec: step.spec,
        values: step.values ? {...step.values} : undefined,
    }
}

export function createWorkflow(draft: WorkflowDraft): Workflow {
    return {
        id: crypto.randomUUID(),
        name: draft.name.trim(),
        createdAt: new Date().toISOString(),
        lifecycle: WorkflowLifecycle.Draft,
        projectDir: draft.projectDir.trim(),
        steps: [],
    }
}

/** Deep copy: new workflow id, new step ids, remapped edges, run history cleared. */
export function copyWorkflow(
    workflow: Workflow,
    name = t('workflow.duplicate.name', {name: workflow.name}),
): Workflow {
    const idMap = new Map(workflow.steps.map((step) => [step.id, crypto.randomUUID()]))

    return {
        id: crypto.randomUUID(),
        name,
        createdAt: new Date().toISOString(),
        lifecycle: WorkflowLifecycle.Draft,
        projectDir: workflow.projectDir,
        steps: workflow.steps.map((step) => ({
            id: idMap.get(step.id)!,
            title: step.title,
            prompt: step.prompt,
            state: StepState.Idle,
            position: {...step.position},
            dependsOn: step.dependsOn.map((id) => idMap.get(id)!).filter(Boolean),
            roleId: step.roleId,
            spec: step.spec,
            values: step.values ? {...step.values} : undefined,
        })),
    }
}

/**
 * A duplicate takes the graph and nothing the graph was pointed at: the steps,
 * their roles and their edges come across, while the project folder starts over
 * empty. The values typed into each step come across only if the operator asked
 * for them when they duplicated.
 */
export function duplicateWorkflow(workflow: Workflow, copyInputs: boolean): Workflow {
    const copy = copyWorkflow(workflow)

    return {
        ...copy,
        projectDir: '',
        steps: copyInputs ? copy.steps : copy.steps.map((step) => ({...step, values: undefined})),
    }
}

export function haltSteps(workflow: Workflow): Workflow {
    return {
        ...workflow,
        steps: workflow.steps.map((step) =>
            step.state === StepState.AwaitingReview || isFinished(step.state)
                ? step
                : notStarted(step),
        ),
    }
}

export function pauseRun(workflow: Workflow): Workflow {
    return {...haltSteps(workflow), lifecycle: WorkflowLifecycle.Paused}
}

export function resumeRun(workflow: Workflow): Workflow {
    return label({...workflow, lifecycle: WorkflowLifecycle.Running})
}

export function isPausable(workflow: Workflow): boolean {
    return workflow.lifecycle === WorkflowLifecycle.Running && isCancellable(workflow)
}

export function isResumable(workflow: Workflow): boolean {
    return workflow.lifecycle === WorkflowLifecycle.Paused
}

function notStarted(step: Step): Step {
    return {...step, state: StepState.Idle, agentId: undefined, run: undefined, report: undefined}
}

function pastRun(workflow: Workflow): PastRun {
    return {
        runId: crypto.randomUUID(),
        remote: workflow.remote,
        startedAt: workflow.startedAt,
        finishedAt: workflow.finishedAt,
        spent: workflow.spent,
        costUsd: workflow.costUsd,
        priced: workflow.priced,
        steps: workflow.steps.map((step) => ({
            id: step.id,
            title: step.title,
            state: step.state,
            agentId: step.agentId,
            run: step.run,
            report: step.report,
        })),
    }
}

export function archiveRun(workflow: Workflow): Workflow {
    const ran = hasStarted(workflow) || workflow.remote

    return {
        ...workflow,
        lifecycle: WorkflowLifecycle.Draft,
        remote: undefined,
        spent: undefined,
        costUsd: undefined,
        priced: undefined,
        startedAt: undefined,
        finishedAt: undefined,
        history: ran
            ? [...(workflow.history ?? []), pastRun(workflow)].slice(-MAX_PAST_RUNS)
            : workflow.history,
        steps: workflow.steps.map(notStarted),
    }
}

/** Cancel is terminal: finished steps keep their reports, everything else loses its work. */
export function cancelRun(workflow: Workflow): Workflow {
    const now = new Date().toISOString()

    return {
        ...workflow,
        lifecycle: WorkflowLifecycle.Cancelled,
        steps: workflow.steps.map((step) => {
            if (KEPT_ON_CANCEL.has(step.state)) return step

            return {
                ...step,
                state: StepState.Cancelled,
                run: step.run ? {...step.run, finishedAt: now, context: undefined} : undefined,
                report: undefined,
            }
        }),
    }
}

/** Reverting to a step keeps it and everything upstream; the run stops and every step after it is back to not started. */
export function rewindTo(workflow: Workflow, stepId: string): Workflow {
    const undone = descendantsOf(workflow.steps, stepId)

    return {
        ...workflow,
        lifecycle:
            workflow.lifecycle === WorkflowLifecycle.Cancelled
                ? WorkflowLifecycle.Cancelled
                : WorkflowLifecycle.Ready,
        steps: workflow.steps.map((step) => {
            const kept =
                step.id === stepId || (!undone.has(step.id) && KEPT_ON_CANCEL.has(step.state))

            return kept ? step : notStarted(step)
        }),
    }
}

export function isCancellable(workflow: Workflow): boolean {
    const underway =
        workflow.lifecycle === WorkflowLifecycle.Running ||
        workflow.lifecycle === WorkflowLifecycle.Paused

    return underway && workflow.steps.some((step) => !isFinished(step.state))
}

/** What an agent writing a new role is told about the workflow it is being written for. */
export function draftContext(workflow: Workflow): DraftContext {
    return {
        workflow_name: workflow.name,
        project_dir: workflow.projectDir,
        steps: workflow.steps.map((step) => ({
            id: step.id,
            title: step.title,
            role_id: step.roleId ?? '',
            depends_on: step.dependsOn,
        })),
    }
}

export function findWorkflow(workflows: Workflow[], workflowId: string | null) {
    if (!workflowId) return undefined
    return workflows.find((workflow) => workflow.id === workflowId)
}

export function moveWorkflow(
    workflows: Workflow[],
    workflowId: string,
    toIndex: number,
): Workflow[] {
    const from = workflows.findIndex((workflow) => workflow.id === workflowId)
    const to = Math.min(Math.max(toIndex, 0), workflows.length - 1)
    if (from < 0 || from === to) return workflows

    const next = [...workflows]
    const [moved] = next.splice(from, 1)
    next.splice(to, 0, moved)

    return next.map((workflow, railRank) =>
        workflow.railRank === railRank ? workflow : {...workflow, railRank},
    )
}

export function byRailRank(a: Workflow, b: Workflow) {
    return (a.railRank ?? 0) - (b.railRank ?? 0)
}

export function findStep(workflow: Workflow, stepId: string | null) {
    if (!stepId) return undefined
    return workflow.steps.find((step) => step.id === stepId)
}

export function withStep(workflow: Workflow, step: Step): Workflow {
    return {...workflow, steps: [...workflow.steps, step]}
}

export function withStepPatch(workflow: Workflow, stepId: string, patch: Partial<Step>): Workflow {
    return {
        ...workflow,
        steps: workflow.steps.map((step) =>
            step.id === stepId ? {...step, ...patch, id: stepId} : step,
        ),
    }
}

export function withoutStep(workflow: Workflow, stepId: string): Workflow {
    return {
        ...workflow,
        steps: workflow.steps
            .filter((step) => step.id !== stepId)
            .map((step) => ({...step, dependsOn: step.dependsOn.filter((id) => id !== stepId)})),
    }
}

export function withDependency(workflow: Workflow, sourceId: string, targetId: string): Workflow {
    if (createsCycle(workflow.steps, sourceId, targetId)) return workflow

    return {
        ...workflow,
        steps: workflow.steps.map((step) => {
            const needsEdge = step.id === targetId && !step.dependsOn.includes(sourceId)
            return needsEdge ? {...step, dependsOn: [...step.dependsOn, sourceId]} : step
        }),
    }
}

export function withoutDependency(
    workflow: Workflow,
    sourceId: string,
    targetId: string,
): Workflow {
    return {
        ...workflow,
        steps: workflow.steps.map((step) =>
            step.id === targetId
                ? {...step, dependsOn: step.dependsOn.filter((id) => id !== sourceId)}
                : step,
        ),
    }
}

export function freePosition(workflow: Workflow, wanted: Point): Point {
    const taken = (point: Point) =>
        workflow.steps.some((step) => step.position.x === point.x && step.position.y === point.y)

    let position = wanted
    while (taken(position)) position = {x: position.x + STEP_STAGGER, y: position.y + STEP_STAGGER}

    return position
}

const HALTED_STATUSES: Partial<Record<WorkflowLifecycle, WorkflowStatus>> = {
    [WorkflowLifecycle.Draft]: WorkflowStatus.Draft,
    [WorkflowLifecycle.Ready]: WorkflowStatus.Ready,
    [WorkflowLifecycle.Paused]: WorkflowStatus.Paused,
    [WorkflowLifecycle.Cancelled]: WorkflowStatus.Cancelled,
}

export function workflowStatus(workflow: Workflow): WorkflowStatus {
    if (workflow.steps.length === 0) return WorkflowStatus.Empty

    const halted = HALTED_STATUSES[workflow.lifecycle]
    if (halted) return halted

    if (workflow.steps.some((step) => step.state === StepState.Failed)) return WorkflowStatus.Failed
    if (workflow.steps.every((step) => step.state === StepState.Done)) return WorkflowStatus.Done

    return WorkflowStatus.Running
}

export function hasRunningStep(workflow: Workflow) {
    return workflow.steps.some((step) => step.state === StepState.Running)
}

export function hasActiveStep(workflow: Workflow) {
    return workflow.steps.some(
        (step) =>
            step.state === StepState.Running ||
            step.state === StepState.Queued ||
            step.state === StepState.AwaitingReview ||
            step.state === StepState.AwaitingApproval,
    )
}

export function workflowProgress(workflow: Workflow) {
    const done = workflow.steps.filter((step) => step.state === StepState.Done).length
    return {done, total: workflow.steps.length}
}

/**
 * A run that ends on a failure is kept open for a retry, so it never gets a completion
 * time — fall back to when its last step stopped, or the clock would run forever.
 */
export function workflowRunWindow(workflow: Workflow) {
    if (hasActiveStep(workflow)) return {startedAt: workflow.startedAt}

    const stops = workflow.steps.flatMap((step) => step.run?.finishedAt ?? [])

    return {startedAt: workflow.startedAt, finishedAt: workflow.finishedAt ?? stops.sort().at(-1)}
}

/** Running the workflow releases the roots; every other step also waits on its upstream — the fan-in join. */
export function isRunnable(workflow: Workflow, step: Step) {
    const upstreamDone = step.dependsOn.every(
        (id) => findStep(workflow, id)?.state === StepState.Done,
    )

    return (
        workflow.lifecycle === WorkflowLifecycle.Running &&
        WAITING_STATES.has(step.state) &&
        upstreamDone
    )
}

export function label(workflow: Workflow): Workflow {
    return {
        ...workflow,
        steps: workflow.steps.map((step) => {
            if (isSettled(step.state)) return step
            return {
                ...step,
                state: isRunnable(workflow, step) ? StepState.Queued : StepState.Blocked,
            }
        }),
    }
}

export function upstreamOf(workflow: Workflow, step: Step) {
    return step.dependsOn
        .map((id) => findStep(workflow, id))
        .filter((candidate): candidate is Step => Boolean(candidate))
}

function reachable(adjacency: Map<string, string[]>, from: string) {
    const seen = new Set<string>()
    const stack = [...(adjacency.get(from) ?? [])]

    while (stack.length > 0) {
        const current = stack.pop()!
        if (seen.has(current)) continue
        seen.add(current)
        stack.push(...(adjacency.get(current) ?? []))
    }

    return seen
}

function ancestorsOf(steps: Step[], stepId: string) {
    return reachable(new Map(steps.map((step) => [step.id, step.dependsOn])), stepId)
}

export function descendantsOf(steps: Step[], stepId: string) {
    const adjacency = new Map<string, string[]>(steps.map((step) => [step.id, []]))
    for (const step of steps)
        for (const dependency of step.dependsOn) adjacency.get(dependency)?.push(step.id)

    return reachable(adjacency, stepId)
}

export function createsCycle(steps: Step[], sourceId: string, targetId: string) {
    return sourceId === targetId || ancestorsOf(steps, sourceId).has(targetId)
}

export function unlinkableFrom(steps: Step[], sourceId: string) {
    const blocked = ancestorsOf(steps, sourceId)
    blocked.add(sourceId)
    return blocked
}

export function unlinkableInto(steps: Step[], targetId: string) {
    const blocked = descendantsOf(steps, targetId)
    blocked.add(targetId)
    return blocked
}
