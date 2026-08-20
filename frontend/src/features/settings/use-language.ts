import {useMutation} from '@tanstack/react-query'

import * as api from '@/features/settings/api/preferences'
import {useLanguage as useCurrentLanguage} from '@/shared/hooks/use-language'
import {t} from '@/shared/lib/i18n'
import type {Language} from '@/shared/lib/enums'

/**
 * The language is applied the moment it is picked and written behind that, so
 * the choice is visible before the config file has caught up. Nothing reads it
 * back — `t` already holds it.
 */
export function useLanguageChoice() {
    const language = useCurrentLanguage()

    const save = useMutation({
        meta: {action: t('settings.error.setLanguage')},
        mutationFn: (next: Language) => api.setLanguage(next),
    })

    return {
        language,
        saving: save.isPending,
        setLanguage: save.mutate,
    }
}
