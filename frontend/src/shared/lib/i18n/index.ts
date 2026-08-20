import {Language} from '@/shared/lib/enums'
import {MESSAGES, type MessageKey} from '@/shared/lib/i18n/messages'

export type {MessageKey}

type Vars = Record<string, string | number>

const PLACEHOLDER = /\{(\w+)\}/g

let current: Language = Language.En

const listeners = new Set<() => void>()

export function currentLanguage(): Language {
    return current
}

export function applyLanguage(language: Language) {
    if (language === current) return

    current = language
    document.documentElement.lang = language

    for (const listener of listeners) listener()
}

export function subscribeLanguage(listener: () => void) {
    listeners.add(listener)
    return () => {
        listeners.delete(listener)
    }
}

function fill(text: string, vars?: Vars) {
    if (!vars) return text

    return text.replace(PLACEHOLDER, (whole, name: string) => {
        const value = vars[name]
        return value === undefined ? whole : String(value)
    })
}

export function t(key: MessageKey, vars?: Vars): string {
    const message = MESSAGES[key]
    return fill(message[current] || message[Language.En], vars)
}

export function tn(one: MessageKey, other: MessageKey, count: number, vars?: Vars): string {
    return t(count === 1 ? one : other, {count, ...vars})
}
