import {InputType, Effort} from '@/shared/lib/enums'
import {t} from '@/shared/lib/i18n'
import {withConflictAwareness} from '@/features/roles/conflict-awareness'
import {structureIssues} from '@/features/roles/output-structure'
import type {FieldValue, InputValue, Role, RoleDraft, RoleInput} from '@/features/roles/types'

export function emptyInput(): RoleInput {
    return {key: '', label: '', type: InputType.Text, required: false}
}

export function emptyRole(): RoleDraft {
    return withConflictAwareness(
        {
            name: '',
            description: '',
            effort: Effort.Standard,
            retryable: true,
            pauseForReview: false,
            inputs: [],
            instructions: [{key: 'base', value: ''}],
            outputStructure: '',
        },
        true,
    )
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

export function hasAdvancedSettings(draft: RoleDraft) {
    return (
        draft.inputs.length > 0 ||
        Boolean(draft.outputStructure.trim()) ||
        draft.effort !== Effort.Standard ||
        !draft.retryable ||
        draft.pauseForReview
    )
}

export function advancedIssues(draft: RoleDraft) {
    const issues: string[] = []
    const keys = draft.inputs.map((input) => input.key.trim()).filter(Boolean)

    if (draft.inputs.some((input) => !input.key.trim())) issues.push(t('role.issue.inputKey'))
    if (new Set(keys).size !== keys.length) issues.push(t('role.issue.duplicateInput'))
    if (
        draft.inputs.some(
            (input) =>
                (input.type === InputType.Select || input.type === InputType.MultiSelect) &&
                !input.options?.length,
        )
    )
        issues.push(t('role.issue.options'))
    if (draft.outputStructure.trim()) issues.push(...structureIssues(draft.outputStructure))

    return issues
}

export function roleIssues(draft: RoleDraft) {
    const issues: string[] = []

    if (!draft.name.trim()) issues.push(t('role.issue.name'))
    if (draft.instructions.length === 0) issues.push(t('role.issue.noInstructions'))
    if (draft.instructions.some((prompt) => !prompt.key.trim()))
        issues.push(t('role.issue.instructionKey'))

    return [...issues, ...advancedIssues(draft)]
}

export function promptFromRole(role: Role) {
    return role.instructions.map((prompt) => prompt.value).join('\n\n')
}
