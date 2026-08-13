import {useMutation, useQuery, useQueryClient} from '@tanstack/react-query'

import * as api from '@/features/templates/api'

const TEMPLATES_KEY = ['templates']

export function useTemplates() {
    const queryClient = useQueryClient()
    const {data, isPending} = useQuery({
        queryKey: TEMPLATES_KEY,
        queryFn: api.listTemplates,
        meta: {action: 'Could not load your templates'},
    })

    const invalidate = () => queryClient.invalidateQueries({queryKey: TEMPLATES_KEY})

    const save = useMutation({
        meta: {action: 'Could not save the template'},
        mutationFn: api.upsertTemplate,
        onSuccess: invalidate,
    })
    const remove = useMutation({
        meta: {action: 'Could not delete the template'},
        mutationFn: api.removeTemplate,
        onSuccess: invalidate,
    })

    return {
        templates: data ?? [],
        loading: isPending,
        saveTemplate: save.mutate,
        saving: save.isPending,
        removeTemplate: remove.mutate,
    }
}

export function useTemplateTransfer() {
    const queryClient = useQueryClient()

    const send = useMutation({
        meta: {action: 'Could not export your templates'},
        mutationFn: ({templateIds, path}: {templateIds: string[]; path: string}) =>
            api.exportTemplates(templateIds, path),
    })

    const receive = useMutation({
        meta: {action: 'Could not import that file'},
        mutationFn: (path: string) => api.importTemplates(path),
        onSuccess: () => queryClient.invalidateQueries({queryKey: TEMPLATES_KEY}),
    })

    return {
        exportTemplates: send.mutateAsync,
        importTemplates: receive.mutateAsync,
        sending: send.isPending,
        receiving: receive.isPending,
    }
}

export function useTemplate(templateId: string | undefined) {
    const {templates} = useTemplates()
    return templates.find((template) => template.id === templateId)
}
