import {createContext, use} from 'react'

import {TOUR_STOPS, type TourStopId, type TourStop} from '@/features/onboarding/tour'

/**
 * The tour drives the app rather than describing it: hosts read the active stop
 * and open what it explains. Kept in a context because the surfaces it opens —
 * settings, the role editor, the rail, the canvas — sit four levels apart, and
 * threading one overlay through all of them would touch every component between.
 */
export type Tour = {
    stop: TourStop | null
    index: number
    total: number
    /** False for the first beat of the role stop, so the user sees the list before the editor. */
    openRole: boolean
    next: () => void
    skip: () => void
}

export const IDLE_TOUR: Tour = {
    stop: null,
    index: 0,
    total: TOUR_STOPS.length,
    openRole: false,
    next: () => {},
    skip: () => {},
}

export const TourContext = createContext<Tour>(IDLE_TOUR)

export function useTour() {
    return use(TourContext)
}

/** The tour as seen by the one surface a stop explains, or null when it is elsewhere. */
export function useTourStop(id: TourStopId) {
    const tour = useTour()
    return tour.stop?.id === id ? tour : null
}
