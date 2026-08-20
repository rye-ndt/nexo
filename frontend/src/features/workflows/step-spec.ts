/**
 * What a step tells the run. A step with a role reads it from the role;
 * an imported step carries the same answers inline, because a workflow travels
 * without the roles its steps were built from.
 */

import {Effort} from '@/shared/lib/enums'
import {roleOf} from '@/features/workflows/step-inputs'
import type {Step, StepSpec} from '@/features/workflows/types'
import type {Role} from '@/features/roles/types'

const DEFAULT_SPEC: StepSpec = {
    effort: Effort.Standard,
    instructions: [],
    outputStructure: '',
    pauseForReview: false,
}

function fromRole(role: Role): StepSpec {
    return {
        effort: role.effort,
        instructions: role.instructions.map((prompt) => prompt.value),
        outputStructure: role.outputStructure,
        pauseForReview: role.pauseForReview,
    }
}

export function specOf(step: Step, roles: Role[]): StepSpec {
    const role = roleOf(step, roles)
    if (role) return fromRole(role)

    return step.spec ?? DEFAULT_SPEC
}

/** Null when neither a role nor the step itself says which level to run at. */
export function effortOf(step: Step, roles: Role[]): Effort | null {
    return roleOf(step, roles)?.effort ?? step.spec?.effort ?? null
}
