/**
 * A run interrupts you in two ways: a finished step holding an accept gate, and
 * a running agent asking a question. Both queue up, both are answered one at a
 * time, and both can be waved off for now. This resolves which one you are
 * being shown and keeps the two queues from talking over each other — an
 * approval always wins, because an agent is stopped dead until it is answered.
 */

import {useCallback, useMemo, useState} from 'react'

import {useApprovals} from '@/features/approvals/use-approvals'
import {StepState} from '@/shared/lib/enums'
import type {Approval, ApprovalAnswer} from '@/features/approvals/types'
import type {Workflow, Step} from '@/features/workflows/types'
import type {WorkflowStore} from '@/features/workflows/use-workflow-store'

type Interrupt<TSubject> = {
    subject: TSubject
    waiting: number
    busy: boolean
    dismiss: () => void
}

type Interrupts = {
    /** The workflow as drawn: steps whose agent is blocked read as awaiting approval. */
    workflow: Workflow | null
    gate: (Interrupt<Step> & {answer: (accepted: boolean) => void}) | null
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

/** A step whose agent is blocked on a question is not really running. */
function withApprovalState(workflow: Workflow, blockedAgents: Set<string>): Workflow {
    if (blockedAgents.size === 0) return workflow

    return {
        ...workflow,
        steps: workflow.steps.map((step) =>
            step.state === StepState.Running && step.agentId && blockedAgents.has(step.agentId)
                ? {...step, state: StepState.AwaitingApproval}
                : step,
        ),
    }
}

export function useInterrupts(store: WorkflowStore): Interrupts {
    const {approvals, answering, answer} = useApprovals()
    const gates = useDismissals()
    const questions = useDismissals()

    const stored = store.active.workflow
    const {selectedStepId} = store

    const blockedAgents = useMemo(
        () => new Set(approvals.map((approval) => approval.agentId)),
        [approvals],
    )

    const workflow = useMemo(
        () => (stored ? withApprovalState(stored, blockedAgents) : null),
        [stored, blockedAgents],
    )

    // Only this workflow's agents; the queue is global and may outlive a run.
    const workflowApprovals = useMemo(() => {
        const agents = new Set(workflow?.steps.map((step) => step.agentId).filter(Boolean))
        return approvals.filter((approval) => agents.has(approval.agentId))
    }, [workflow, approvals])

    const selectedStep = workflow?.steps.find((step) => step.id === selectedStepId)

    const asking = nextInLine(
        workflowApprovals,
        (approval) => approval.id,
        workflowApprovals.find((approval) => approval.agentId === selectedStep?.agentId),
        questions.dismissed,
    )

    const gated = workflow?.steps.filter((step) => step.state === StepState.AwaitingReview) ?? []

    const gating = nextInLine(
        gated,
        (step) => step.id,
        gated.find((step) => step.id === selectedStepId),
        gates.dismissed,
    )

    const deselect = (stepId: string | undefined) => {
        if (stepId && selectedStepId === stepId) store.selectStep(null)
    }

    const approval = asking
        ? {
              subject: asking,
              waiting: workflowApprovals.length,
              busy: answering,
              answer: (reply: ApprovalAnswer) => answer(asking.id, reply),
              dismiss: () => {
                  questions.dismiss(asking.id)
                  deselect(workflow?.steps.find((step) => step.agentId === asking.agentId)?.id)
              },
          }
        : null

    const gate =
        gating && !asking
            ? {
                  subject: gating,
                  waiting: gated.length,
                  busy: store.answeringStepAcceptance,
                  answer: (accepted: boolean) =>
                      store.active.answerStepAcceptance(gating.id, accepted),
                  dismiss: () => {
                      gates.dismiss(gating.id)
                      deselect(gating.id)
                  },
              }
            : null

    return {workflow, gate, approval}
}
