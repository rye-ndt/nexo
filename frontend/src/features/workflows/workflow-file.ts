/**
 * The two pure halves of workflow transfer. A workflow travels without its step
 * roles, so what each role would have told the run is inlined onto the
 * step on the way out and read back off it on the way in.
 */

import {copyWorkflow} from '@/features/workflows/graph'
import {specOf} from '@/features/workflows/step-spec'
import {t} from '@/shared/lib/i18n'
import type {Workflow} from '@/features/workflows/types'
import type {Role} from '@/features/roles/types'

function notAWorkflow() {
    return new Error(t('workflow.api.fileInvalid'))
}

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
        throw notAWorkflow()
    }

    if (typeof parsed?.name !== 'string' || !Array.isArray(parsed.steps)) throw notAWorkflow()

    const ids = new Set(parsed.steps.map((step) => step?.id))

    for (const step of parsed.steps) {
        if (typeof step?.id !== 'string') throw notAWorkflow()
        if (typeof step.title !== 'string' || typeof step.prompt !== 'string') throw notAWorkflow()
        if (!isPoint(step.position)) throw notAWorkflow()
        if (!Array.isArray(step.dependsOn)) throw notAWorkflow()
        if (step.dependsOn.some((id) => !ids.has(id))) throw notAWorkflow()
    }

    return parsed
}

export function fromExportedWorkflow(raw: string): Workflow {
    const parsed = readWorkflow(raw)
    return copyWorkflow(parsed, parsed.name)
}
