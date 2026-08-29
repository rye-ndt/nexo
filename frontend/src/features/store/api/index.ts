/**
 * The store has no backend of its own and needs none: the catalog is compiled in,
 * and adding from it writes through the seams that already exist. Roles go out
 * over the roles api here; the workflow is only built, and the rail's own import
 * mutation lands it, the same way a workflow read from a file does.
 */

import {WorkflowLifecycle, StepState} from '@/shared/lib/enums'
import {STORE_ROLES, STORE_TEMPLATES} from '@/features/store/catalog'
import {copyWorkflow} from '@/features/workflows/graph'
import {listRoles, upsertRole} from '@/features/roles/api'
import type {Role} from '@/features/roles/types'
import type {StoreTemplate} from '@/features/store/types'
import type {Workflow} from '@/features/workflows/types'

export function storeRoles(): Role[] {
    return structuredClone(STORE_ROLES)
}

export function storeTemplates(): StoreTemplate[] {
    return structuredClone(STORE_TEMPLATES)
}

/** The roles a template's steps run on, in the order the steps first ask for them. */
export function templateRoles(template: StoreTemplate): Role[] {
    const seen = new Set<string>()
    const roles: Role[] = []

    for (const step of template.steps) {
        if (seen.has(step.roleId)) continue
        seen.add(step.roleId)

        const role = STORE_ROLES.find((candidate) => candidate.id === step.roleId)
        if (role) roles.push(role)
    }

    return structuredClone(roles)
}

function seedWorkflow(template: StoreTemplate): Workflow {
    return {
        id: template.id,
        name: template.name,
        createdAt: new Date().toISOString(),
        lifecycle: WorkflowLifecycle.Draft,
        projectDir: '',
        steps: template.steps.map((step) => ({
            ...step,
            state: StepState.Idle,
            values: {},
        })),
    }
}

/**
 * Installs the roles a template's steps run on, then answers the graph itself for
 * the rail to import. The roles go first because a graph pointing at roles the
 * library does not hold would lock but never run; the ones already there are left
 * exactly as the user has them.
 */
export async function prepareStoreTemplate(template: StoreTemplate): Promise<Workflow> {
    const held = new Set((await listRoles()).map((role) => role.id))

    for (const role of templateRoles(template)) {
        if (!held.has(role.id)) await upsertRole(role)
    }

    return copyWorkflow(seedWorkflow(template), template.name)
}
