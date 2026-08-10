/**
 * A run interrupts you in two ways: a finished node holding an accept gate, and
 * a running agent asking a question. Both queue up, both are answered one at a
 * time, and both can be waved off for now. This resolves which one you are
 * being shown and keeps the two queues from talking over each other — an
 * approval always wins, because an agent is stopped dead until it is answered.
 */

import {useCallback, useMemo, useState} from 'react'

import {useApprovals} from '@/features/approvals/use-approvals'
import {TaskState} from '@/shared/lib/enums'
import type {Approval, ApprovalAnswer} from '@/features/approvals/types'
import type {Session, Task} from '@/features/sessions/types'
import type {SessionStore} from '@/features/sessions/use-session-store'

export type Interrupt<TSubject> = {
    subject: TSubject
    waiting: number
    busy: boolean
    dismiss: () => void
}

export type Interrupts = {
    /** The session as drawn: nodes whose agent is blocked read as awaiting approval. */
    session: Session | null
    gate: (Interrupt<Task> & {answer: (accepted: boolean) => void}) | null
    approval: (Interrupt<Approval> & {answer: (reply: ApprovalAnswer) => void}) | null
}

function useDismissals() {
    const [dismissed, setDismissed] = useState<ReadonlySet<string>>(() => new Set())

    const dismiss = useCallback(
        (id: string) => setDismissed((current) => new Set(current).add(id)),
        [],
    )

    return {dismissed, dismiss}
}

/** The selected one if it is waiting, otherwise the first you have not waved off. */
function nextInLine<T>(
    queue: T[],
    idOf: (item: T) => string,
    selected: T | undefined,
    dismissed: ReadonlySet<string>,
) {
    return selected ?? queue.find((item) => !dismissed.has(idOf(item)))
}

/** A node whose agent is blocked on a question is not really running. */
function withApprovalState(session: Session, blockedAgents: Set<string>): Session {
    if (blockedAgents.size === 0) return session

    return {
        ...session,
        tasks: session.tasks.map((task) =>
            task.state === TaskState.Running && task.agentId && blockedAgents.has(task.agentId)
                ? {...task, state: TaskState.AwaitingApproval}
                : task,
        ),
    }
}

export function useInterrupts(store: SessionStore): Interrupts {
    const {approvals, answering, answer} = useApprovals()
    const gates = useDismissals()
    const questions = useDismissals()

    const stored = store.active.session
    const {selectedTaskId} = store

    const blockedAgents = useMemo(
        () => new Set(approvals.map((approval) => approval.agentId)),
        [approvals],
    )

    const session = useMemo(
        () => (stored ? withApprovalState(stored, blockedAgents) : null),
        [stored, blockedAgents],
    )

    // Only this session's agents; the queue is global and may outlive a run.
    const sessionApprovals = useMemo(() => {
        const agents = new Set(session?.tasks.map((task) => task.agentId).filter(Boolean))
        return approvals.filter((approval) => agents.has(approval.agentId))
    }, [session, approvals])

    const selectedTask = session?.tasks.find((task) => task.id === selectedTaskId)

    const asking = nextInLine(
        sessionApprovals,
        (approval) => approval.id,
        sessionApprovals.find((approval) => approval.agentId === selectedTask?.agentId),
        questions.dismissed,
    )

    const gated = session?.tasks.filter((task) => task.state === TaskState.AwaitingAccept) ?? []

    const gating = nextInLine(
        gated,
        (task) => task.id,
        gated.find((task) => task.id === selectedTaskId),
        gates.dismissed,
    )

    const deselect = (taskId: string | undefined) => {
        if (taskId && selectedTaskId === taskId) store.selectTask(null)
    }

    const approval = asking
        ? {
              subject: asking,
              waiting: sessionApprovals.length,
              busy: answering,
              answer: (reply: ApprovalAnswer) => answer(asking.id, reply),
              dismiss: () => {
                  questions.dismiss(asking.id)
                  deselect(session?.tasks.find((task) => task.agentId === asking.agentId)?.id)
              },
          }
        : null

    const gate =
        gating && !asking
            ? {
                  subject: gating,
                  waiting: gated.length,
                  busy: store.answeringTaskAcceptance,
                  answer: (accepted: boolean) =>
                      store.active.answerTaskAcceptance(gating.id, accepted),
                  dismiss: () => {
                      gates.dismiss(gating.id)
                      deselect(gating.id)
                  },
              }
            : null

    return {session, gate, approval}
}
