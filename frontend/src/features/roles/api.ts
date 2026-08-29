/**
 * The only place the generated Wails bindings for roles are touched. The
 * wire shape keeps inputs and instructions as maps keyed by name, so this
 * file maps them to the ordered lists the frontend Role type uses. Under
 * the plain vite dev server there is no Go side and roles live in this
 * module's memory on top of src/lib/mock-roles.ts.
 */

import {bridge, hasWailsRuntime} from '@/shared/api/bridge'
import {INPUT_TYPES, InputType, EFFORTS, Effort} from '@/shared/lib/enums'
import {listAgents} from '@/features/agents/api'
import {
    mockArchive,
    mockHelperBlocked,
    mockImported,
    mockRefined,
    MOCK_ROLES,
} from '@/features/roles/mock-roles'
import {mockReadFile, mockWriteFile} from '@/shared/api/mock-fs'
import {t} from '@/shared/lib/i18n'
import type {DraftContext, Role, RoleDraft} from '@/features/roles/types'
import {core_itf, output_itf} from '@wailsjs/go/models'
import {
    ExportRoles,
    ImportRoles,
    RefineRole,
    RemoveRole,
    Role as FetchRole,
    RoleHelperBlocked,
    Roles as FetchRoles,
    UpsertRole,
} from '@wailsjs/go/wails_api/API'

const ROUNDTRIP_MS = 400

/** The real call spawns an agent and waits for it, so the mock has to feel like it. */
const REFINE_MS = 2600

let roles: Role[] = structuredClone(MOCK_ROLES)

async function roundtrip() {
    await new Promise((resolve) => setTimeout(resolve, ROUNDTRIP_MS))
}

function effortOf(value: string): Effort {
    return EFFORTS.find((level) => level === value) ?? Effort.Standard
}

function byKey([left]: [string, unknown], [right]: [string, unknown]) {
    return left.localeCompare(right)
}

function toRole(info: output_itf.RoleInfo): Role {
    return {
        id: info.id,
        name: info.name,
        description: info.description,
        effort: effortOf(info.effort),
        retryable: info.retryable,
        pauseForReview: info.pause_for_review,
        inputs: Object.entries(info.inputs ?? {})
            .sort(byKey)
            .map(([key, input]) => ({
                key,
                label: input?.description ?? '',
                type: INPUT_TYPES.find((type) => type === input?.type) ?? InputType.Text,
                required: input?.required ?? false,
                default: input?.default || undefined,
                options: input?.options?.length ? input.options : undefined,
            })),
        instructions: Object.entries(info.instructions ?? {})
            .sort(byKey)
            .map(([key, value]) => ({key, value})),
        outputStructure: info.output_structure ?? '',
    }
}

function toInfo(draft: RoleDraft): output_itf.RoleInfo {
    return new output_itf.RoleInfo({
        id: draft.id ?? '',
        name: draft.name,
        description: draft.description,
        effort: draft.effort,
        retryable: draft.retryable,
        pause_for_review: draft.pauseForReview,
        inputs: Object.fromEntries(
            draft.inputs.map((input) => [
                input.key,
                {
                    description: input.label,
                    required: input.required,
                    type: input.type,
                    default: input.default ?? '',
                    options: input.options ?? [],
                },
            ]),
        ),
        instructions: Object.fromEntries(
            draft.instructions.map((prompt) => [prompt.key, prompt.value]),
        ),
        output_structure: draft.outputStructure,
    })
}

type RolesBackend = {
    list(): Promise<Role[]>
    upsert(draft: RoleDraft): Promise<Role>
    export(roleIds: string[], path: string): Promise<number>
    import(path: string): Promise<number>
    helperBlocked(): Promise<string>
    refine(draft: RoleDraft, context?: DraftContext): Promise<RoleDraft>
    remove(roleId: string): Promise<void>
}

const wailsRoles: RolesBackend = {
    list: async () => {
        roles = (await bridge(FetchRoles)).map(toRole)
        return roles
    },
    upsert: async (draft) => {
        const id = await bridge(() => UpsertRole(toInfo(draft)))
        return toRole(await bridge(() => FetchRole(id)))
    },
    export: async (roleIds, path) => bridge(() => ExportRoles(roleIds, path)),
    import: async (path) => bridge(() => ImportRoles(path)),
    helperBlocked: async () => bridge(RoleHelperBlocked),
    refine: async (draft, context) => {
        const request = new core_itf.DraftRequest({
            name: draft.name.trim(),
            description: draft.description.trim(),
            ...context,
        })
        const filled = toRole(await bridge(() => RefineRole(request)))

        return {...filled, id: draft.id}
    },
    remove: async (roleId) => {
        await bridge(() => RemoveRole(roleId))
    },
}

const mockRolesBackend: RolesBackend = {
    list: async () => roles,
    upsert: async (draft) => {
        await roundtrip()

        const next: Role = {...draft, id: draft.id ?? crypto.randomUUID()}
        roles = roles.some((role) => role.id === next.id)
            ? roles.map((role) => (role.id === next.id ? next : role))
            : [...roles, next]

        return next
    },
    export: async (roleIds, path) => {
        await roundtrip()

        const picked = roles.filter((role) => roleIds.includes(role.id))
        if (picked.length !== roleIds.length) throw new Error(t('role.error.roleGone'))

        mockWriteFile(path, mockArchive(picked))

        return picked.length
    },
    import: async (path) => {
        await roundtrip()

        const imported = mockImported(roles, mockReadFile(path), path)
        roles = [...roles, ...imported]

        return imported.length
    },
    helperBlocked: async () => mockHelperBlocked(await listAgents()),
    refine: async (draft, context) => {
        const blocked = mockHelperBlocked(await listAgents())
        if (blocked) throw new Error(blocked)

        await new Promise((resolve) => setTimeout(resolve, REFINE_MS))

        return {
            ...mockRefined(draft.name.trim(), draft.description.trim(), context),
            id: draft.id,
        }
    },
    remove: async (roleId) => {
        await roundtrip()
        roles = roles.filter((role) => role.id !== roleId)
    },
}

const backend: RolesBackend = hasWailsRuntime() ? wailsRoles : mockRolesBackend

export async function listRoles(): Promise<Role[]> {
    return structuredClone(await backend.list())
}

/** The last fetched roles, for the run loop, which cannot await. */
export function cachedRoles(): Role[] {
    return roles
}

export async function upsertRole(draft: RoleDraft): Promise<Role> {
    if (!draft.name.trim()) throw new Error(t('role.error.nameRequired'))

    return structuredClone(await backend.upsert(draft))
}

export async function exportRoles(roleIds: string[], path: string): Promise<number> {
    if (roleIds.length === 0) throw new Error(t('role.error.pickToExport'))

    return backend.export(roleIds, path)
}

export async function importRoles(path: string): Promise<number> {
    return backend.import(path)
}

/**
 * Why a role cannot be filled in right now, empty when it can. Under vite the
 * answer comes from the mock agent roster, so logging the harness in and out in
 * Agents actually opens and closes this.
 */
export async function roleHelperBlocked(): Promise<string> {
    return backend.helperBlocked()
}

/**
 * Hands the name, the role and the graph the role is being written for to an
 * agent, which reads the project before it writes, and waits for the whole role
 * back. The Go side refuses anything the agent leaves half-written, so what lands
 * here is either a complete role or an error.
 */
export async function refineRole(draft: RoleDraft, context?: DraftContext): Promise<RoleDraft> {
    return backend.refine(draft, context)
}

export async function removeRole(roleId: string): Promise<void> {
    return backend.remove(roleId)
}
