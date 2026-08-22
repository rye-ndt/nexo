import type {Rect} from '@/features/onboarding/types'

/** Breathing room between the ring and what it rings. */
const HALO = 6

const GAP = 14
const MARGIN = 16

export type Side = 'right' | 'bottom' | 'left' | 'top'

export type Size = {width: number; height: number}

export type Placement = {top: number; left: number; side: Side}

const FALLBACKS: Record<Side, Side[]> = {
    right: ['right', 'left', 'bottom', 'top'],
    bottom: ['bottom', 'top', 'right', 'left'],
    left: ['left', 'right', 'bottom', 'top'],
    top: ['top', 'bottom', 'right', 'left'],
}

export function haloOf(rect: Rect): Rect {
    return {
        top: rect.top - HALO,
        left: rect.left - HALO,
        width: rect.width + HALO * 2,
        height: rect.height + HALO * 2,
    }
}

const clamp = (value: number, low: number, high: number) =>
    Math.max(low, Math.min(value, Math.max(low, high)))

function fits(side: Side, halo: Rect, card: Size, viewport: Size) {
    if (side === 'right')
        return halo.left + halo.width + GAP + card.width + MARGIN <= viewport.width
    if (side === 'left') return halo.left - GAP - card.width - MARGIN >= 0
    if (side === 'bottom')
        return halo.top + halo.height + GAP + card.height + MARGIN <= viewport.height
    return halo.top - GAP - card.height - MARGIN >= 0
}

/**
 * Where the card sits next to the element it explains: the asked-for side when
 * it fits, otherwise the first of the remaining three that does. The card never
 * covers what it points at, so a side that fits nowhere still falls back to a
 * clamped position rather than centring over the target.
 */
export function placeCard(rect: Rect, side: Side, card: Size, viewport: Size): Placement {
    const halo = haloOf(rect)
    const chosen = FALLBACKS[side].find((option) => fits(option, halo, card, viewport)) ?? side

    const alongX = clamp(
        halo.left + halo.width / 2 - card.width / 2,
        MARGIN,
        viewport.width - card.width - MARGIN,
    )
    const alongY = clamp(
        halo.top + halo.height / 2 - card.height / 2,
        MARGIN,
        viewport.height - card.height - MARGIN,
    )

    const acrossX = (left: number) => clamp(left, MARGIN, viewport.width - card.width - MARGIN)
    const acrossY = (top: number) => clamp(top, MARGIN, viewport.height - card.height - MARGIN)

    if (chosen === 'right')
        return {top: alongY, left: acrossX(halo.left + halo.width + GAP), side: chosen}
    if (chosen === 'left')
        return {top: alongY, left: acrossX(halo.left - GAP - card.width), side: chosen}
    if (chosen === 'bottom')
        return {top: acrossY(halo.top + halo.height + GAP), left: alongX, side: chosen}

    return {top: acrossY(halo.top - GAP - card.height), left: alongX, side: chosen}
}
