/**
 * What a running step is doing, right now. Activity is chatter rather than
 * state: it lives in this module's memory and never lands on a Step or a
 * Workflow, because every change to a Workflow is written back to disk as one
 * JSON doc and this feed grows for the whole life of a run.
 *
 * The array held for a step keeps its identity until something newer merges,
 * so a poll that brings nothing new does not re-render the step.
 */

import type {ActivityLine, Workflow} from '@/features/workflows/types'

const MAX_LINES = 200

const NO_LINES: ActivityLine[] = []

const linesByStep = new Map<string, ActivityLine[]>()

function highestSeq(lines: ActivityLine[]) {
    return lines.length === 0 ? 0 : lines[lines.length - 1].seq
}

export function mergeActivity(stepId: string, incoming: ActivityLine[]) {
    if (incoming.length === 0) return

    const held = linesByStep.get(stepId) ?? NO_LINES
    const highest = highestSeq(held)

    // A retry hands the step to a fresh agent, which counts from one again.
    if (highestSeq(incoming) < highest) {
        linesByStep.set(stepId, incoming.slice(-MAX_LINES))
        return
    }

    const fresh = incoming.filter((line) => line.seq > highest)
    if (fresh.length === 0) return

    linesByStep.set(stepId, [...held, ...fresh].slice(-MAX_LINES))
}

export async function stepActivity(stepId: string): Promise<ActivityLine[]> {
    return linesByStep.get(stepId) ?? NO_LINES
}

function forgetActivity(stepId: string) {
    linesByStep.delete(stepId)
}

export function forgetWorkflowActivity(workflow: Workflow) {
    for (const step of workflow.steps) forgetActivity(step.id)
}
