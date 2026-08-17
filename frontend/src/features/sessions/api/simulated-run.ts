/**
 * Outside the webview (the plain vite dev server) there is no runtime, so the
 * graph starts from the mock sessions and finalizing falls back to a simulated
 * run that walks the graph in order. A node whose template asks for manual
 * acceptance halts the walk until `resolveAcceptance` answers it, and a node
 * whose title names one of the forks below halts partway through its own run,
 * blocked on the approval it raised, until the operator answers that.
 */

import {ApprovalKind, TaskState} from '@/shared/lib/enums'
import {mockOutcome} from '@/features/sessions/mock-sessions'
import {hasRunningTask, isRunnable, label, withTaskPatch} from '@/features/sessions/graph'
import {specOf} from '@/features/sessions/task-spec'
import {cachedAutopilot} from '@/features/settings/api/preferences'
import {cachedTemplates} from '@/features/templates/api'
import type {ApprovalOption} from '@/features/approvals/types'
import {
    isMockApprovalPending,
    mockApprovalAnswer,
    raiseMockApproval,
} from '@/features/approvals/mock-approvals'
import type {Session, Task, TaskReport} from '@/features/sessions/types'
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

    const forked = share >= FORK_SHARE ? atFork(task, now, outcome.durationMs) : undefined
    if (forked) return {...forked, run: {...forked.run, context}}

    if (share < 1) return {...task, run: {...task.run, context}}

    const state =
        outcome.state === TaskState.Done && gated(task, templates)
            ? TaskState.AwaitingAccept
            : outcome.state

    return {
        ...task,
        state,
        run: {...task.run, finishedAt: new Date(now).toISOString(), context},
        report: {...outcome.report, status: state},
    }
}

/**
 * There is no backend clock or counter here, so both session readouts come off the
 * nodes: no mock node retries, so what every node holds now is what the run spent.
 */
function withRunTotals(session: Session): Session {
    const started = session.tasks.flatMap((task) => task.run?.startedAt ?? [])
    const finished = session.tasks.flatMap((task) => task.run?.finishedAt ?? [])

    return {
        ...session,
        tokensUsed: session.tasks.reduce(
            (total, task) => total + (task.run?.context?.used ?? 0),
            0,
        ),
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
