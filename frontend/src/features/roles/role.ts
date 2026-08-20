import {InputType, Effort} from '@/shared/lib/enums'
import {structureIssues} from '@/features/roles/output-structure'
import type {FieldValue, InputValue, Role, RoleDraft, RoleInput} from '@/features/roles/types'

export const NO_INSTRUCTIONS_ISSUE = 'A role needs at least one instruction.'

export function emptyInput(): RoleInput {
    return {key: '', label: '', type: InputType.Text, required: false}
}

export function emptyRole(): RoleDraft {
    return {
        name: '',
        description: '',
        effort: Effort.Standard,
        retryable: true,
        pauseForReview: false,
        inputs: [emptyInput()],
        instructions: [{key: 'base', value: ''}],
        outputStructure: '',
    }
}

export function chosenOptions(value: FieldValue | undefined): string[] {
    return String(value ?? '')
        .split(',')
        .map((option) => option.trim())
        .filter(Boolean)
}

export function joinOptions(options: string[]): string {
    return options.join(', ')
}

function defaultOf(input: RoleInput): FieldValue {
    if (input.type === InputType.Boolean) return input.default === 'true'
    if (input.type === InputType.Select && !input.options?.includes(input.default ?? '')) return ''
    if (input.type === InputType.MultiSelect)
        return joinOptions(
            chosenOptions(input.default).filter((option) => input.options?.includes(option)),
        )
    return input.default ?? ''
}

export function defaultFieldValues(role: Role): Record<string, FieldValue> {
    return Object.fromEntries(role.inputs.map((input) => [input.key, defaultOf(input)]))
}

export function toFieldValues(
    role: Role,
    values: Record<string, InputValue> = {},
): Record<string, FieldValue> {
    return Object.fromEntries(
        role.inputs.map((input) => {
            const value = values[input.key]
            if (input.type === InputType.Boolean)
                return [input.key, value === undefined ? input.default === 'true' : value === true]

            return [input.key, value === undefined ? (input.default ?? '') : String(value)]
        }),
    )
}

export function toInputValues(
    role: Role,
    values: Record<string, FieldValue>,
): Record<string, InputValue> {
    return Object.fromEntries(
        role.inputs.map((input) => {
            const value = values[input.key]
            if (input.type === InputType.Boolean) return [input.key, value === true]
            if (input.type === InputType.Number)
                return [input.key, value === '' ? 0 : Number(value)]
            return [input.key, String(value ?? '')]
        }),
    )
}

export function missingRequired(role: Role, values: Record<string, FieldValue>) {
    return role.inputs.filter((input) => {
        const answerable = input.required && input.type !== InputType.Boolean
        return answerable && String(values[input.key] ?? '') === ''
    })
}

export function roleIssues(draft: RoleDraft) {
    const issues: string[] = []
    const keys = draft.inputs.map((input) => input.key.trim()).filter(Boolean)

    if (!draft.name.trim()) issues.push('Give the role a name.')
    if (draft.inputs.some((input) => !input.key.trim()))
        issues.push('Every input needs a key the agent can read.')
    if (new Set(keys).size !== keys.length) issues.push('Two inputs share the same key.')
    if (
        draft.inputs.some(
            (input) =>
                (input.type === InputType.Select || input.type === InputType.MultiSelect) &&
                !input.options?.length,
        )
    )
        issues.push('A choice input needs at least one option.')
    if (draft.instructions.length === 0) issues.push(NO_INSTRUCTIONS_ISSUE)
    if (draft.instructions.some((prompt) => !prompt.key.trim()))
        issues.push('Every instruction needs a key.')
    if (draft.outputStructure.trim()) issues.push(...structureIssues(draft.outputStructure))

    return issues
}

export function promptFromRole(role: Role) {
    return role.instructions.map((prompt) => prompt.value).join('\n\n')
}
