/**
 * A step asks for inputs: the required ones, plus any input its prompt embeds
 * as `{{key}}`. Inputs stay editable on the step until the run starts, and
 * pressing Run points out the empty ones, but none of them block it — a
 * reference left unfilled reaches the agent as the literal `{{key}}`.
 */

import {InputType} from '@/shared/lib/enums'
import {fillPrompt, inputRefs} from '@/features/roles/input-refs'
import {toFieldValues} from '@/features/roles/role'
import type {Workflow, Step} from '@/features/workflows/types'
import type {Role, RoleInput} from '@/features/roles/types'

export function roleOf(step: Step, roles: Role[]) {
    return roles.find((role) => role.id === step.roleId)
}

export function pendingInputs(step: Step, role: Role | undefined): RoleInput[] {
    if (!role) return []

    const values = toFieldValues(role, step.values)
    const embedded = new Set(inputRefs(step.prompt))

    return role.inputs.filter((input) => {
        if (input.type === InputType.Boolean) return false
        if (!input.required && !embedded.has(input.key)) return false
        return String(values[input.key] ?? '') === ''
    })
}

export type MissingStepInputs = {step: Step; inputs: RoleInput[]}

export function missingInputs(workflow: Workflow, roles: Role[]): MissingStepInputs[] {
    return workflow.steps
        .map((step) => ({step, inputs: pendingInputs(step, roleOf(step, roles))}))
        .filter((entry) => entry.inputs.length > 0)
}

/** Embedded inputs land inline; the rest follow as a list the agent can read. */
export function resolvedPrompt(step: Step): string {
    const values = step.values ?? {}
    const embedded = new Set(inputRefs(step.prompt))
    const rest = Object.entries(values).filter(
        ([key, value]) => !embedded.has(key) && String(value) !== '',
    )

    const filled = fillPrompt(step.prompt, values)
    if (rest.length === 0) return filled

    const lines = rest.map(([key, value]) => `- ${key}: ${String(value)}`)
    return `${filled}\n\nInputs:\n${lines.join('\n')}`
}
