import {useMemo} from 'react'
import {useMutation, useQueryClient} from '@tanstack/react-query'

import * as api from '@/features/store/api'
import {ROLES_KEY, useRoles} from '@/features/roles/use-roles'
import {t} from '@/shared/lib/i18n'
import type {StoreTemplate} from '@/features/store/types'
import type {Workflow, WorkflowLocations} from '@/features/workflows/types'

/**
 * The catalog is compiled in, so the only thing the store has to fetch is the
 * user's own role library — which is what tells a card whether it is already
 * theirs. Adding a workflow prepares its roles and hands the graph to the rail's
 * own import, the way useWorkflowTransfer hands over one read from a file.
 */
export function useStore(onImport: (workflow: Workflow, locations: WorkflowLocations) => void) {
    const queryClient = useQueryClient()
    const {roles: library, loading, saveRole, saving} = useRoles()

    const roles = useMemo(() => api.storeRoles(), [])
    const templates = useMemo(() => api.storeTemplates(), [])
    const held = useMemo(() => new Set(library.map((role) => role.id)), [library])

    const prepare = useMutation({
        meta: {action: t('store.error.addWorkflow')},
        mutationFn: (template: StoreTemplate) => api.prepareStoreTemplate(template),
        onSuccess: () => queryClient.invalidateQueries({queryKey: ROLES_KEY}),
    })

    const addTemplate = async (template: StoreTemplate, locations: WorkflowLocations) => {
        const workflow = await prepare.mutateAsync(template).catch(() => null)
        if (!workflow) return false

        onImport(workflow, locations)
        return true
    }

    return {
        roles,
        templates,
        /** True until the library has been read, so no card may claim to know either way. */
        busy: loading || saving,
        held: (roleId: string) => held.has(roleId),
        rolesOf: api.templateRoles,

        addRole: saveRole,
        addTemplate,
        addingTemplate: prepare.isPending,
    }
}

export type Store = ReturnType<typeof useStore>
