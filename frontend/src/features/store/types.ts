import type {Point} from '@/features/workflows/types'

/**
 * One step of a store workflow: the graph shape and the prompt, with the role it
 * runs on named by id. Deliberately narrower than the app's Step — nothing in the
 * store has run, so it carries no state, no agent and no result.
 */
export type StoreStep = {
    id: string
    title: string
    prompt: string
    position: Point
    dependsOn: string[]
    roleId: string
}

export type StoreTemplate = {
    id: string
    name: string
    description: string
    steps: StoreStep[]
}

export const StoreSection = {
    Workflows: 'workflows',
    Roles: 'roles',
} as const

export type StoreSection = (typeof StoreSection)[keyof typeof StoreSection]
