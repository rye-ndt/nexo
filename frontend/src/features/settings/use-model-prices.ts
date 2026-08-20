import {useMutation, useQuery, useQueryClient} from '@tanstack/react-query'

import * as api from '@/features/settings/api/preferences'
import {t} from '@/shared/lib/i18n'
import type {TokenPrices} from '@/features/settings/types'

const MODEL_PRICES_KEY = ['model-prices']

type ModelPricesEdit = {
    model: string
    prices: TokenPrices
}

export function useModelPrices() {
    const queryClient = useQueryClient()

    const prices = useQuery({
        queryKey: MODEL_PRICES_KEY,
        queryFn: api.listModelPrices,
        meta: {action: t('settings.error.loadPrices')},
    })

    const save = useMutation({
        meta: {action: t('settings.error.savePrice')},
        mutationFn: ({model, prices}: ModelPricesEdit) => api.setModelPrices(model, prices),
        onSuccess: () => queryClient.invalidateQueries({queryKey: MODEL_PRICES_KEY}),
    })

    return {
        modelPrices: prices.data ?? [],
        loading: prices.isPending,
        setModelPrices: save.mutate,
    }
}
