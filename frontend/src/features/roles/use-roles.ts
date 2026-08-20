import {useMutation, useQuery, useQueryClient} from '@tanstack/react-query'

import * as api from '@/features/roles/api'
import {t} from '@/shared/lib/i18n'

const ROLES_KEY = ['roles']

export function useRoles() {
    const queryClient = useQueryClient()
    const {data, isPending} = useQuery({
        queryKey: ROLES_KEY,
        queryFn: api.listRoles,
        meta: {action: t('role.error.load')},
    })

    const invalidate = () => queryClient.invalidateQueries({queryKey: ROLES_KEY})

    const save = useMutation({
        meta: {action: t('role.error.save')},
        mutationFn: api.upsertRole,
        onSuccess: invalidate,
    })
    const remove = useMutation({
        meta: {action: t('role.error.remove')},
        mutationFn: api.removeRole,
        onSuccess: invalidate,
    })

    return {
        roles: data ?? [],
        loading: isPending,
        saveRole: save.mutate,
        saving: save.isPending,
        removeRole: remove.mutate,
    }
}

export function useRoleTransfer() {
    const queryClient = useQueryClient()

    const send = useMutation({
        meta: {action: t('role.error.export')},
        mutationFn: ({roleIds, path}: {roleIds: string[]; path: string}) =>
            api.exportRoles(roleIds, path),
    })

    const receive = useMutation({
        meta: {action: t('role.error.import')},
        mutationFn: (path: string) => api.importRoles(path),
        onSuccess: () => queryClient.invalidateQueries({queryKey: ROLES_KEY}),
    })

    return {
        exportRoles: send.mutateAsync,
        importRoles: receive.mutateAsync,
        sending: send.isPending,
        receiving: receive.isPending,
    }
}

export function useRole(roleId: string | undefined) {
    const {roles} = useRoles()
    return roles.find((role) => role.id === roleId)
}
