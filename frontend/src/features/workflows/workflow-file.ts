/**
 * The two pure halves of workflow transfer. A workflow travels without its step
 * roles, so what each role would have told the run is inlined onto the
 * step on the way out and read back off it on the way in.
 */

import {copyWorkflow} from '@/features/workflows/graph'
import {specOf} from '@/features/workflows/step-spec'
import type {Workflow} from '@/features/workflows/types'
import type {Role} from '@/features/roles/types'

const NOT_A_WORKFLOW = 'That file is not a workflow export.'

export function toExportedWorkflow(workflow: Workflow, roles: Role[]): Workflow {
    const clean = copyWorkflow(workflow, workflow.name)

    return {
        ...clean,
        steps: clean.steps.map((step) => ({
            ...step,
            spec: specOf(step, roles),
            roleId: undefined,
        })),
    }
}

function isPoint(value: unknown): boolean {
    const point = value as {x?: unknown; y?: unknown} | null
    return typeof point?.x === 'number' && typeof point?.y === 'number'
}

function readWorkflow(raw: string): Workflow {
    let parsed: Workflow

    try {
        parsed = JSON.parse(raw) as Workflow
    } catch {
        throw new Error(NOT_A_WORKFLOW)
    }

    if (typeof parsed?.name !== 'string' || !Array.isArray(parsed.steps))
        throw new Error(NOT_A_WORKFLOW)

    const ids = new Set(parsed.steps.map((step) => step?.id))

    for (const step of parsed.steps) {
        if (typeof step?.id !== 'string') throw new Error(NOT_A_WORKFLOW)
        if (typeof step.title !== 'string' || typeof step.prompt !== 'string')
            throw new Error(NOT_A_WORKFLOW)
        if (!isPoint(step.position)) throw new Error(NOT_A_WORKFLOW)
        if (!Array.isArray(step.dependsOn)) throw new Error(NOT_A_WORKFLOW)
        if (step.dependsOn.some((id) => !ids.has(id))) throw new Error(NOT_A_WORKFLOW)
    }

    return parsed
}

export function fromExportedWorkflow(raw: string): Workflow {
    const parsed = readWorkflow(raw)
    return copyWorkflow(parsed, parsed.name)
}
