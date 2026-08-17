/**
 * What a node tells the run. A node with a template reads it from the template;
 * an imported node carries the same answers inline, because a session travels
 * without the templates its nodes were built from.
 */

import {TaskLevel} from '@/shared/lib/enums'
import {templateOf} from '@/features/sessions/task-inputs'
import type {Task, TaskSpec} from '@/features/sessions/types'
import type {Template} from '@/features/templates/types'

const DEFAULT_SPEC: TaskSpec = {
    taskLevel: TaskLevel.Daily,
    systemPrompts: [],
    outputStructure: '',
    manualAcceptRequired: false,
}

function fromTemplate(template: Template): TaskSpec {
    return {
        taskLevel: template.taskLevel,
        systemPrompts: template.systemPrompts.map((prompt) => prompt.value),
        outputStructure: template.outputStructure,
        manualAcceptRequired: template.manualAcceptRequired,
    }
}

export function specOf(task: Task, templates: Template[]): TaskSpec {
    const template = templateOf(task, templates)
    if (template) return fromTemplate(template)

    return task.spec ?? DEFAULT_SPEC
}

/** Null when neither a template nor the node itself says which level to run at. */
export function taskLevelOf(task: Task, templates: Template[]): TaskLevel | null {
    return templateOf(task, templates)?.taskLevel ?? task.spec?.taskLevel ?? null
}
