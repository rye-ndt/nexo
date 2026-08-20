import {useMutation, useQuery, useQueryClient} from '@tanstack/react-query'

import {
    authorizeMCPServer,
    listMCPServers,
    revokeMCPServer,
    setMCPCredential,
} from '@/features/settings/api/mcp'
import {t} from '@/shared/lib/i18n'

const MCP_SERVERS_KEY = ['mcp-servers']

export function useMCPServers() {
    const queryClient = useQueryClient()

    const {data, isPending} = useQuery({
        queryKey: MCP_SERVERS_KEY,
        queryFn: listMCPServers,
        meta: {action: t('settings.error.loadServers')},
    })

    const authorize = useMutation({
        meta: {action: t('settings.error.authorizeServer')},
        mutationFn: authorizeMCPServer,
        onSuccess: () => queryClient.invalidateQueries({queryKey: MCP_SERVERS_KEY}),
    })

    const setToken = useMutation({
        meta: {action: t('settings.error.saveToken')},
        mutationFn: ({serverId, token}: {serverId: string; token: string}) =>
            setMCPCredential(serverId, token),
        onSuccess: () => queryClient.invalidateQueries({queryKey: MCP_SERVERS_KEY}),
    })

    const revoke = useMutation({
        meta: {action: t('settings.error.revokeServer')},
        mutationFn: revokeMCPServer,
        onSuccess: () => queryClient.invalidateQueries({queryKey: MCP_SERVERS_KEY}),
    })

    return {
        servers: data ?? [],
        loading: isPending,
        pendingId: authorize.isPending ? authorize.variables : null,
        authorize: authorize.mutate,
        setToken: setToken.mutate,
        savingId: setToken.isPending ? (setToken.variables?.serverId ?? null) : null,
        revoke: revoke.mutate,
        revokingId: revoke.isPending ? revoke.variables : null,
    }
}
