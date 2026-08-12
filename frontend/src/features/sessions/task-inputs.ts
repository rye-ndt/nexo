/**
 * A node asks for inputs: the required ones, plus any input its prompt embeds
 * as `{{key}}`. Inputs stay editable on the node until the run starts, and
 * pressing Run points out the empty ones, but none of them block it — a
 * reference left unfilled reaches the agent as the literal `{{key}}`.
 */

import {ParamType} from '@/shared/lib/enums'
import {fillPrompt, paramRefs} from '@/features/templates/param-refs'
import {toFieldValues} from '@/features/templates/template'
import type {Session, Task} from '@/features/sessions/types'
import type {Template, TemplateParam} from '@/features/templates/types'

export function templateOf(task: Task, templates: Template[]) {
    return templates.find((template) => template.id === task.templateId)
}

export function pendingParams(task: Task, template: Template | undefined): TemplateParam[] {
    if (!template) return []

    const values = toFieldValues(template, task.values)
    const embedded = new Set(paramRefs(task.prompt))

    return template.params.filter((param) => {
        if (param.type === ParamType.Boolean) return false
        if (!param.required && !embedded.has(param.key)) return false
        return String(values[param.key] ?? '') === ''
    })
}

export type MissingNodeInputs = {task: Task; params: TemplateParam[]}

export function missingInputs(session: Session, templates: Template[]): MissingNodeInputs[] {
    return session.tasks
        .map((task) => ({task, params: pendingParams(task, templateOf(task, templates))}))
        .filter((entry) => entry.params.length > 0)
}

/** Embedded inputs land inline; the rest follow as a list the agent can read. */
export function resolvedPrompt(task: Task): string {
    const values = task.values ?? {}
    const embedded = new Set(paramRefs(task.prompt))
    const rest = Object.entries(values).filter(
        ([key, value]) => !embedded.has(key) && String(value) !== '',
    )

    const filled = fillPrompt(task.prompt, values)
    if (rest.length === 0) return filled

    const lines = rest.map(([key, value]) => `- ${key}: ${String(value)}`)
    return `${filled}\n\nInputs:\n${lines.join('\n')}`
}
