import {useSyncExternalStore} from 'react'

import {currentLanguage, subscribeLanguage} from '@/shared/lib/i18n'
import type {Language} from '@/shared/lib/enums'

/**
 * Re-renders the tree under whoever calls it when the language changes. `t` is a
 * plain function so nothing else in the app has to hook into anything; this is
 * the one subscription that turns a language change into a repaint.
 */
export function useLanguage(): Language {
    return useSyncExternalStore(subscribeLanguage, currentLanguage)
}
