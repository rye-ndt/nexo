import {t} from '@/shared/lib/i18n'

export function formatTokens(tokens: number) {
    if (tokens < 1000) return String(tokens)
    if (tokens < 1_000_000) return compact(tokens / 1000, 'k')

    return compact(tokens / 1_000_000, 'M')
}

function compact(value: number, unit: string) {
    return `${value < 100 ? value.toFixed(1) : Math.round(value)}${unit}`
}

/**
 * A run that has barely started costs fractions of a cent, and rounding those to
 * $0.00 reads as free. Anything under a cent keeps enough places to stay a number.
 */
export function formatUSD(dollars: number) {
    const places = dollars > 0 && dollars < 0.01 ? 4 : 2

    return `$${dollars.toLocaleString([], {
        minimumFractionDigits: places,
        maximumFractionDigits: places,
    })}`
}

export function formatPercent(used: number, total: number) {
    if (total <= 0) return 0
    return Math.round(clampRatio(used / total) * 100)
}

export function clampRatio(ratio: number) {
    return Math.min(1, Math.max(0, ratio))
}

export function formatDuration(ms: number) {
    const seconds = Math.floor(Math.max(0, ms) / 1000)
    if (seconds < 60) return `${seconds}s`
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`
    return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`
}

export function formatMoment(iso?: string) {
    if (!iso) return '—'

    const date = new Date(iso)
    if (Number.isNaN(date.getTime())) return iso

    return date.toLocaleString([], {
        month: 'short',
        day: 'numeric',
        hour: '2-digit',
        minute: '2-digit',
    })
}

export function formatRelative(iso: string) {
    const then = Date.parse(iso)
    if (Number.isNaN(then)) return ''

    const seconds = Math.max(0, Math.round((Date.now() - then) / 1000))
    if (seconds < 60) return t('format.justNow')

    const minutes = Math.floor(seconds / 60)
    if (minutes < 60) return t('format.minutesAgo', {count: minutes})

    const hours = Math.floor(minutes / 60)
    if (hours < 24) return t('format.hoursAgo', {count: hours})

    const days = Math.floor(hours / 24)
    if (days < 365) return t('format.daysAgo', {count: days})

    return t('format.yearsAgo', {count: Math.floor(days / 365)})
}

export function formatAgentName(name: string) {
    return name
        .split('_')
        .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
        .join(' ')
}
