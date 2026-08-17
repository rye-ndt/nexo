/**
 * Outside the webview (the plain vite dev server) there is no runtime, so the
 * graph starts from the mock sessions and finalizing falls back to a simulated
 * run that walks the graph in order. A node whose template asks for manual
 * acceptance halts the walk until `resolveAcceptance` answers it, and a node
 * whose title names one of the forks below halts partway through its own run,
 * blocked on the approval it raised, until the operator answers that. What a run
 * costs is priced off the agent defaults the settings panel edits, so blanking a
 * level's input or output price in Settings → Preferences is what puts a run back into
 * the unpriced state.
 */

import {ApprovalKind, TaskState, type TaskLevel} from '@/shared/lib/enums'
import {mockOutcome} from '@/features/sessions/mock-sessions'
import {hasRunningTask, isRunnable, label, withTaskPatch} from '@/features/sessions/graph'
import {specOf} from '@/features/sessions/task-spec'
import {cachedAgentDefaults, cachedAutopilot} from '@/features/settings/api/preferences'
import {cachedTemplates} from '@/features/templates/api'
import type {ApprovalOption} from '@/features/approvals/types'
import {
    isMockApprovalPending,
    mockApprovalAnswer,
    raiseMockApproval,
} from '@/features/approvals/mock-approvals'
import type {Session, Spend, Task, TaskReport} from '@/features/sessions/types'
import type {Template} from '@/features/templates/types'
import {replaceSession, sessions, setSessions} from '@/features/sessions/api/store'
import {mergeActivity} from '@/features/sessions/api/activity'
import {mockActivity} from '@/features/sessions/api/mock-activity'
import {TICK_MS, timers} from '@/features/sessions/api/timers'

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
        detail: 'Twenty-one ports are declared there and the coordinator plausibly calls seven of them. Listing all twenty-one is the honest answer, but it buries the seven the next node has to implement against, and that node inherits this doc as the head of its prompt.',
        options: [
            {
                id: 'called-only',
                label: 'Only the ports the coordinator calls',
                description:
                    'Seven ports, each with the call that reaches it. The next node gets a list it can work straight through.',
            },
            {
                id: 'everything',
                label: 'Every port in interface/core',
                description:
                    'All twenty-one, marked called or not. Complete, but the next node has to do this filtering again.',
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
                label: 'One Advance(sessionID)',
                description:
                    'The port stays a single method. The implementation owns when a node is finished, when the next context is built, and when it is assigned.',
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

function forkOf(task: Task) {
    const title = task.title.toLowerCase()
    return FORKS.find((fork) => title.includes(fork.match))
}

function held(task: Task, now: number, durationMs: number): Task {
    return {
        ...task,
        run: {...task.run, startedAt: new Date(now - durationMs * FORK_SHARE).toISOString()},
    }
}

function rejection(task: Task, fork: Fork, guidance: string): TaskReport {
    const said = guidance.trim()

    return {
        status: TaskState.Failed,
        handoverDocs: [
            {
                task: task.title,
                tldr: said
                    ? `I stopped at a fork I could not settle on my own, and you turned down every option I offered. You said: ${said}`
                    : 'I stopped at a fork I could not settle on my own, and you turned down every option I offered without saying which way to go.',
                outcome:
                    'Stopped at the fork. Nothing was written, so the tree is as the upstream node left it.',
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
                    : {direction: 'Duplicate the session and say in the prompt which way to go.'},
            },
        ],
    }
}

function atFork(task: Task, now: number, durationMs: number): Task | undefined {
    const fork = forkOf(task)
    if (!fork) return undefined

    const approvalId = asked.get(task.id)

    if (approvalId === undefined) {
        asked.set(
            task.id,
            raiseMockApproval({
                agentId: task.agentId ?? task.id,
                kind: ApprovalKind.Decision,
                question: fork.question,
                detail: fork.detail,
                options: fork.options,
                multiSelect: false,
            }),
        )

        return held(task, now, durationMs)
    }

    if (isMockApprovalPending(approvalId)) return held(task, now, durationMs)

    const answer = mockApprovalAnswer(approvalId)
    if (!answer || answer.approved) return undefined

    return {
        ...task,
        state: TaskState.Failed,
        run: {...task.run, finishedAt: new Date(now).toISOString()},
        report: rejection(task, fork, answer.guidance),
    }
}

function start(task: Task, now: number): Task {
    const {contextTotal, contextPeak} = mockOutcome(task)

    return {
        ...task,
        state: TaskState.Running,
        agentId: crypto.randomUUID(),
        run: {
            ...task.run,
            startedAt: new Date(now).toISOString(),
            finishedAt: undefined,
            context: {used: Math.round(contextTotal * contextPeak * 0.2), total: contextTotal},
            spent: task.run?.spent ?? {input: 0, cached: 0, output: 0},
        },
    }
}

function gated(task: Task, templates: Template[]) {
    if (cachedAutopilot()) return false
    return specOf(task, templates).manualAcceptRequired
}

function progress(task: Task, now: number, templates: Template[]): Task {
    const outcome = mockOutcome(task)
    const startedAt = task.run?.startedAt ? Date.parse(task.run.startedAt) : now
    const share = Math.min(1, Math.max(0, (now - startedAt) / outcome.durationMs))
    const context = {
        used: Math.round(outcome.contextTotal * outcome.contextPeak * (0.2 + 0.8 * share)),
        total: outcome.contextTotal,
    }
    const spent = {
        input: Math.max(task.run?.spent?.input ?? 0, Math.round(outcome.inputTotal * share)),
        cached: Math.max(task.run?.spent?.cached ?? 0, Math.round(outcome.cachedTotal * share)),
        output: Math.max(task.run?.spent?.output ?? 0, Math.round(outcome.outputTotal * share)),
    }

    const forked = share >= FORK_SHARE ? atFork(task, now, outcome.durationMs) : undefined
    if (forked) return {...forked, run: {...forked.run, context, spent}}

    if (share < 1) return {...task, run: {...task.run, context, spent}}

    const state =
        outcome.state === TaskState.Done && gated(task, templates)
            ? TaskState.AwaitingAccept
            : outcome.state

    return {
        ...task,
        state,
        run: {...task.run, finishedAt: new Date(now).toISOString(), context, spent},
        report: {...outcome.report, status: state},
    }
}

type Rates = {input: number; cached: number; output: number}

function ratesOf(taskLevel: TaskLevel): Rates | undefined {
    const prices = cachedAgentDefaults().find((current) => current.taskLevel === taskLevel)?.prices
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

function withCost(task: Task, templates: Template[]): Task {
    const spent = task.run?.spent
    if (!spent) return task

    const rates = ratesOf(specOf(task, templates).taskLevel)
    if (!rates) return {...task, run: {...task.run, costUsd: undefined, priced: false}}

    return {...task, run: {...task.run, costUsd: costOf(spent, rates), priced: true}}
}

/**
 * There is no backend clock, counter or price table here, so every run readout comes
 * off the nodes. A node keeps the tokens all of its attempts have spent, which only
 * ever grow, and is priced at the agent default for its task level: that level counts
 * as priced only when it carries both an input and an output price, a blank cached
 * price falls back to the input one, and the cost is
 * (input * inputPrice + cached * cachedPrice + output * outputPrice) / 1_000_000.
 * The session sums what its nodes spent and what the priced ones cost, and goes
 * unpriced as soon as one node that spent something has no prices behind it.
 */
function withRunTotals(session: Session): Session {
    const templates = cachedTemplates()
    const tasks = session.tasks.map((task) => withCost(task, templates))

    const spends = tasks.flatMap((task) => task.run?.spent ?? [])
    const started = tasks.flatMap((task) => task.run?.startedAt ?? [])
    const finished = tasks.flatMap((task) => task.run?.finishedAt ?? [])

    return {
        ...session,
        tasks,
        spent: spends.reduce(
            (total, spend) => ({
                input: total.input + spend.input,
                cached: total.cached + spend.cached,
                output: total.output + spend.output,
            }),
            {input: 0, cached: 0, output: 0},
        ),
        costUsd: tasks.reduce((total, task) => total + (task.run?.costUsd ?? 0), 0),
        priced: !tasks.some((task) => spentAnything(task.run?.spent) && !task.run?.priced),
        startedAt: started.sort()[0],
        finishedAt: finished.sort().at(-1),
    }
}

function advance(session: Session): Session {
    const now = Date.now()
    const templates = cachedTemplates()

    const progressed = label({
        ...session,
        tasks: session.tasks.map((task) =>
            task.state === TaskState.Running ? progress(task, now, templates) : task,
        ),
    })

    const started = label({
        ...progressed,
        tasks: progressed.tasks.map((task) =>
            isRunnable(progressed, task) ? start(task, now) : task,
        ),
    })

    return withRunTotals(started)
}

function narrate(session: Session) {
    const now = Date.now()

    for (const task of session.tasks)
        if (task.state === TaskState.Running) mergeActivity(task.id, mockActivity(task, now))
}

export function tick(sessionId: string) {
    const session = sessions.find((session) => session.id === sessionId)
    if (!session?.started || session.cancelled) return

    const next = replaceSession(advance(session))
    narrate(next)
    if (hasRunningTask(next)) schedule(sessionId)
}

export function resolveAcceptance(sessionId: string, taskId: string, accepted: boolean) {
    const session = sessions.find((session) => session.id === sessionId)
    const task = session?.tasks.find((task) => task.id === taskId)
    if (!session || !task) return

    const state = accepted ? TaskState.Done : TaskState.Failed

    replaceSession(
        withTaskPatch(session, taskId, {
            state,
            report: task.report ? {...task.report, status: state} : undefined,
        }),
    )

    tick(sessionId)
}

function schedule(sessionId: string) {
    if (timers.has(sessionId)) return

    timers.set(
        sessionId,
        setTimeout(() => {
            timers.delete(sessionId)
            tick(sessionId)
        }, TICK_MS),
    )
}

export function stopRun(sessionId: string) {
    const timer = timers.get(sessionId)
    if (timer === undefined) return

    clearTimeout(timer)
    timers.delete(sessionId)
}

// A session that already settled never ticks, so its totals are stamped once here.
setSessions(sessions.map(withRunTotals))

for (const session of sessions)
    if (!session.cancelled && hasRunningTask(session)) schedule(session.id)
