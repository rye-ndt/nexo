import type {ParamType, TaskLevel} from '@/shared/lib/enums'

export type TemplateParam = {
    key: string
    label: string
    type: ParamType
    required: boolean
    default?: string
    options?: string[]
}

export type SystemPrompt = {
    key: string
    value: string
}

export type Template = {
    id: string
    name: string
    role: string
    taskLevel: TaskLevel
    retryable: boolean
    manualAcceptRequired: boolean
    params: TemplateParam[]
    systemPrompts: SystemPrompt[]
    /** The fields every node must return, or empty for a free handover. */
    outputStructure: string
}

export type TemplateDraft = Omit<Template, 'id'> & {id?: string}

export type TemplateRecord = {
    id: string
    name: string
    role: string
    task_level: string
    retryable: boolean
    manual_accept_required: boolean
    params: Record<string, TemplateParamRecord>
    system_prompts: Record<string, string>
    output_structure: string
}

export type TemplateParamRecord = {
    description: string
    required: boolean
    type: string
    default: string
    options: string[]
}

export type TemplateArchive = {
    version: number
    exported_at: string
    templates: TemplateRecord[]
}

export type FieldValue = string | boolean

export type ParamValue = string | number | boolean
