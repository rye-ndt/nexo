/**
 * Outside the webview (the plain vite dev server) there is no runtime, so the
 * graph starts from the mock workflows and locking falls back to a simulated
 * run that walks the graph in order. A step whose role asks for manual
 * acceptance halts the walk until `resolveAcceptance` answers it, and a step
 * whose title names one of the forks below halts partway through its own run,
 * blocked on the approval it raised, until the operator answers that. What a run
 * costs is priced off the model prices the settings panel edits, so blanking a
 * model's input or output price in Settings → Preferences is what puts a run back into
 * the unpriced state.
 */

import {ApprovalKind, StepState, type Effort} from '@/shared/lib/enums'
import {mockOutcome} from '@/features/workflows/mock-workflows'
import {hasRunningStep, isRunnable, label, withStepPatch} from '@/features/workflows/graph'
import {specOf} from '@/features/workflows/step-spec'
import {
    cachedAgentDefaults,
    cachedAutopilot,
    cachedMaxRunningAgents,
    cachedModelPrices,
} from '@/features/settings/api/preferences'
import {cachedRoles} from '@/features/roles/api'
import type {ApprovalOption} from '@/features/approvals/types'
import {
    isMockApprovalPending,
    mockApprovalAnswer,
    raiseMockApproval,
} from '@/features/approvals/mock-approvals'
import type {Workflow, Spend, Step, StepResult} from '@/features/workflows/types'
import type {Role} from '@/features/roles/types'
import {replaceWorkflow, workflows, setWorkflows} from '@/features/workflows/api/store'
import {mergeActivity} from '@/features/workflows/api/activity'
import {mockActivity} from '@/features/workflows/api/mock-activity'
import {TICK_MS, timers} from '@/features/workflows/api/timers'

const FORK_SHARE = 0.45

type Fork = {
    match: string
    question: string
    detail: string
    options: ApprovalOption[]
}

const FORKS: Fork[] = [
    {
        match: 'graph ports',
        question: 'How much of internal/interface/core should the port map cover?',
        detail: 'Twenty-one ports are declared there and the coordinator plausibly calls seven of them. Listing all twenty-one is the honest answer, but it buries the seven the next step has to implement against, and that step inherits this doc as the head of its prompt.',
        options: [
            {
                id: 'called-only',
                label: 'Only the ports the coordinator calls',
                description:
                    'Seven ports, each with the call that reaches it. The next step gets a list it can work straight through.',
            },
            {
                id: 'everything',
                label: 'Every port in interface/core',
                description:
                    'All twenty-one, marked called or not. Complete, but the next step has to do this filtering again.',
            },
            {
                id: 'split',
                label: 'The seven first, the rest as an appendix',
                description:
                    'Lead with what the coordinator needs and keep the remainder below a divider for later.',
            },
        ],
    },
    {
        match: 'coordinator interface',
        question: 'Should the coordinator port be one Advance(), or a call per lifecycle step?',
        detail: 'One method keeps the port minimal and leaves the state machine inside the implementation. Three methods name the lifecycle at the wire.go seam, which reads better, but they freeze that lifecycle into the interface and undoing it later touches every caller.',
        options: [
            {
                id: 'single-advance',
                label: 'One Advance(workflowID)',
                description:
                    'The port stays a single method. The implementation owns when a step is finished, when the next context is built, and when it is assigned.',
            },
            {
                id: 'per-step',
                label: 'Finish, Build, Assign',
                description:
                    'Three calls that say what the coordinator does. Clearer to read, harder to change once wire.go depends on the shape.',
            },
        ],
    },
]

const asked = new Map<string, string>()

function forkOf(step: Step) {
    const title = step.title.toLowerCase()
    return FORKS.find((fork) => title.includes(fork.match))
}

function held(step: Step, now: number, durationMs: number): Step {
    return {
        ...step,
        run: {...step.run, startedAt: new Date(now - durationMs * FORK_SHARE).toISOString()},
    }
}

function rejection(step: Step, fork: Fork, guidance: string): StepResult {
    const said = guidance.trim()

    return {
        status: StepState.Failed,
        handoffs: [
            {
                step: step.title,
                tldr: said
                    ? `I stopped at a fork I could not settle on my own, and you turned down every option I offered. You said: ${said}`
                    : 'I stopped at a fork I could not settle on my own, and you turned down every option I offered without saying which way to go.',
                outcome:
                    'Stopped at the fork. Nothing was written, so the tree is as the upstream step left it.',
                blockers: {fork: fork.question},
                approvedDecisions: {},
                rejectedDecisions: {
                    every_option: fork.options.map((option) => option.label).join(' / '),
                },
                currentBehaviors: {},
                changedBehaviors: {},
                mustAvoid: {},
                nuances: {},
                knownGaps: said
                    ? {guidance: said}
                    : {direction: 'Duplicate the workflow and say in the prompt which way to go.'},
            },
        ],
    }
}

function atFork(step: Step, now: number, durationMs: number): Step | undefined {
    const fork = forkOf(step)
    if (!fork) return undefined

    const approvalId = asked.get(step.id)

    if (approvalId === undefined) {
        asked.set(
            step.id,
            raiseMockApproval({
                agentId: step.agentId ?? step.id,
                kind: ApprovalKind.Decision,
                question: fork.question,
                detail: fork.detail,
                options: fork.options,
                multiSelect: false,
            }),
        )

        return held(step, now, durationMs)
    }

    if (isMockApprovalPending(approvalId)) return held(step, now, durationMs)

    const answer = mockApprovalAnswer(approvalId)
    if (!answer || answer.approved) return undefined

    return {
        ...step,
        state: StepState.Failed,
        run: {...step.run, finishedAt: new Date(now).toISOString()},
        report: rejection(step, fork, answer.guidance),
    }
}

function start(step: Step, now: number): Step {
    const {contextTotal, contextPeak} = mockOutcome(step)

    return {
        ...step,
        state: StepState.Running,
        agentId: crypto.randomUUID(),
        run: {
            ...step.run,
            startedAt: new Date(now).toISOString(),
            finishedAt: undefined,
            context: {used: Math.round(contextTotal * contextPeak * 0.2), total: contextTotal},
            spent: step.run?.spent ?? {input: 0, cached: 0, output: 0},
        },
    }
}

function gated(step: Step, roles: Role[]) {
    if (cachedAutopilot()) return false
    return specOf(step, roles).pauseForReview
}

function progress(step: Step, now: number, roles: Role[]): Step {
    const outcome = mockOutcome(step)
    const startedAt = step.run?.startedAt ? Date.parse(step.run.startedAt) : now
    const share = Math.min(1, Math.max(0, (now - startedAt) / outcome.durationMs))
    const context = {
        used: Math.round(outcome.contextTotal * outcome.contextPeak * (0.2 + 0.8 * share)),
        total: outcome.contextTotal,
    }
    const spent = {
        input: Math.max(step.run?.spent?.input ?? 0, Math.round(outcome.inputTotal * share)),
        cached: Math.max(step.run?.spent?.cached ?? 0, Math.round(outcome.cachedTotal * share)),
        output: Math.max(step.run?.spent?.output ?? 0, Math.round(outcome.outputTotal * share)),
    }

    const forked = share >= FORK_SHARE ? atFork(step, now, outcome.durationMs) : undefined
    if (forked) return {...forked, run: {...forked.run, context, spent}}

    if (share < 1) return {...step, run: {...step.run, context, spent}}

    const state =
        outcome.state === StepState.Done && gated(step, roles)
            ? StepState.AwaitingReview
            : outcome.state

    return {
        ...step,
        state,
        run: {...step.run, finishedAt: new Date(now).toISOString(), context, spent},
        report: {...outcome.report, status: state},
    }
}

type Rates = {input: number; cached: number; output: number}

function ratesOf(effort: Effort): Rates | undefined {
    const model = cachedAgentDefaults().find((current) => current.effort === effort)?.model
    const prices = cachedModelPrices().find((current) => current.model === model)?.prices
    if (!prices || prices.input === '' || prices.output === '') return undefined

    return {
        input: Number(prices.input),
        cached: Number(prices.cachedInput === '' ? prices.input : prices.cachedInput),
        output: Number(prices.output),
    }
}

function spentAnything(spent?: Spend) {
    return !!spent && (spent.input > 0 || spent.cached > 0 || spent.output > 0)
}

function costOf(spent: Spend, rates: Rates) {
    return (
        (spent.input * rates.input + spent.cached * rates.cached + spent.output * rates.output) /
        1_000_000
    )
}

function withCost(step: Step, roles: Role[]): Step {
    const spent = step.run?.spent
    if (!spent) return step

    const rates = ratesOf(specOf(step, roles).effort)
    if (!rates) return {...step, run: {...step.run, costUsd: undefined, priced: false}}

    return {...step, run: {...step.run, costUsd: costOf(spent, rates), priced: true}}
}

/**
 * There is no backend clock, counter or price table here, so every run readout comes
 * off the steps. A step keeps the tokens all of its attempts have spent, which only
 * ever grow, and is priced at what the model behind its step level costs: that model
 * counts as priced only when it carries both an input and an output price, a blank
 * cached price falls back to the input one, and the cost is
 * (input * inputPrice + cached * cachedPrice + output * outputPrice) / 1_000_000.
 * The workflow sums what its steps spent and what the priced ones cost, and goes
 * unpriced as soon as one step that spent something has no prices behind it.
 */
function withRunTotals(workflow: Workflow): Workflow {
    const roles = cachedRoles()
    const steps = workflow.steps.map((step) => withCost(step, roles))

    const spends = steps.flatMap((step) => step.run?.spent ?? [])
    const started = steps.flatMap((step) => step.run?.startedAt ?? [])
    const finished = steps.flatMap((step) => step.run?.finishedAt ?? [])

    return {
        ...workflow,
        steps,
        spent: spends.reduce(
            (total, spend) => ({
                input: total.input + spend.input,
                cached: total.cached + spend.cached,
                output: total.output + spend.output,
            }),
            {input: 0, cached: 0, output: 0},
        ),
        costUsd: steps.reduce((total, step) => total + (step.run?.costUsd ?? 0), 0),
        priced: !steps.some((step) => spentAnything(step.run?.spent) && !step.run?.priced),
        startedAt: started.sort()[0],
        finishedAt: finished.sort().at(-1),
    }
}

function advance(workflow: Workflow): Workflow {
    const now = Date.now()
    const roles = cachedRoles()

    const progressed = label({
        ...workflow,
        steps: workflow.steps.map((step) =>
            step.state === StepState.Running ? progress(step, now, roles) : step,
        ),
    })

    let room =
        cachedMaxRunningAgents() -
        progressed.steps.filter((step) => step.state === StepState.Running).length

    const started = label({
        ...progressed,
        steps: progressed.steps.map((step) => {
            if (room < 1 || !isRunnable(progressed, step)) return step
            room -= 1
            return start(step, now)
        }),
    })

    return withRunTotals(started)
}

function narrate(workflow: Workflow) {
    const now = Date.now()

    for (const step of workflow.steps)
        if (step.state === StepState.Running) mergeActivity(step.id, mockActivity(step, now))
}

export function tick(workflowId: string) {
    const workflow = workflows.find((workflow) => workflow.id === workflowId)
    if (!workflow?.started || workflow.cancelled || workflow.paused) return

    const next = replaceWorkflow(advance(workflow))
    narrate(next)
    if (hasRunningStep(next)) schedule(workflowId)
}

export function resolveAcceptance(workflowId: string, stepId: string, accepted: boolean) {
    const workflow = workflows.find((workflow) => workflow.id === workflowId)
    const step = workflow?.steps.find((step) => step.id === stepId)
    if (!workflow || !step) return

    const state = accepted ? StepState.Done : StepState.Failed

    replaceWorkflow(
        withStepPatch(workflow, stepId, {
            state,
            report: step.report ? {...step.report, status: state} : undefined,
        }),
    )

    tick(workflowId)
}

function schedule(workflowId: string) {
    if (timers.has(workflowId)) return

    timers.set(
        workflowId,
        setTimeout(() => {
            timers.delete(workflowId)
            tick(workflowId)
        }, TICK_MS),
    )
}

export function forgetWorkflowForks(workflow: Workflow) {
    for (const step of workflow.steps) asked.delete(step.id)
}

export function stopRun(workflowId: string) {
    const timer = timers.get(workflowId)
    if (timer === undefined) return

    clearTimeout(timer)
    timers.delete(workflowId)
}

// A workflow that already settled never ticks, so its totals are stamped once here.
setWorkflows(workflows.map(withRunTotals))

for (const workflow of workflows)
    if (!workflow.cancelled && hasRunningStep(workflow)) schedule(workflow.id)
