import {clampRatio} from '@/shared/lib/format'

/**
 * Both context rings read as stamina, so they are coloured by what a step has
 * left rather than by what it has spent. Green while there is room to work,
 * orange once the window is more than half gone, red when it is nearly out.
 */
const GREEN_ABOVE = 0.5

const ORANGE_ABOVE = 0.15

const DRAINED_AT = 0.02

/** The share of the context window this step has not used yet, 0 to 1. */
export function contextLeft(used: number, total: number) {
    return total > 0 ? clampRatio(1 - used / total) : 1
}

export function contextRingClass(left: number) {
    if (left > GREEN_ABOVE) return 'stroke-state-done'
    if (left > ORANGE_ABOVE) return 'stroke-state-approval'
    return 'stroke-state-failed'
}

export function contextDrained(left: number) {
    return left <= DRAINED_AT
}
