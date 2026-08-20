import type {InputType, Effort} from '@/shared/lib/enums'

export type RoleInput = {
    key: string
    label: string
    type: InputType
    required: boolean
    default?: string
    options?: string[]
}

export type Instruction = {
    key: string
    value: string
}

export type Role = {
    id: string
    name: string
    description: string
    effort: Effort
    retryable: boolean
    pauseForReview: boolean
    inputs: RoleInput[]
    instructions: Instruction[]
    /** The fields every step must return, or empty for a free-form handoff. */
    outputStructure: string
}

export type RoleDraft = Omit<Role, 'id'> & {id?: string}

/**
 * Everything the agent filling a role in is told about where that role is
 * going: the folders its steps will run in, and the graph it is being written for.
 * Declared in the wire shape, so the api layer hands it straight to Go. Undefined
 * when the role is written from Settings, with no workflow around it.
 */
export type DraftContext = {
    workflow_name: string
    project_dir: string
    steps: {
        id: string
        title: string
        role_id: string
        depends_on: string[]
    }[]
}

export type RoleRecord = {
    id: string
    name: string
    description: string
    effort: string
    retryable: boolean
    pause_for_review: boolean
    inputs: Record<string, RoleInputRecord>
    instructions: Record<string, string>
    output_structure: string
}

export type RoleInputRecord = {
    description: string
    required: boolean
    type: string
    default: string
    options: string[]
}

export type RoleArchive = {
    version: number
    exported_at: string
    roles: RoleRecord[]
}

export type FieldValue = string | boolean

export type InputValue = string | number | boolean
