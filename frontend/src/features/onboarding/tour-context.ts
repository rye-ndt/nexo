import {createContext, use} from 'react'

import {TOUR_STOPS, type TourStopId, type TourStop} from '@/features/onboarding/tour'

/**
 * The tour drives the app rather than describing it: hosts read the active stop
 * and open what it explains. Kept in a context because the surfaces it opens —
 * the store, the rail, the canvas, the header — sit four levels apart, and
 * threading one overlay through all of them would touch every component between.
 */
export type Tour = {
    stop: TourStop | null
    index: number
    total: number
    /** True while the stop being explained lives in the store, which the tour opens itself. */
    openStore: boolean
    next: () => void
    /** Advances past every stop the store hosts, for when the user adds from any of them. */
    leaveStore: () => void
    skip: () => void
}

const IDLE_TOUR: Tour = {
    stop: null,
    index: 0,
    total: TOUR_STOPS.length,
    openStore: false,
    next: () => {},
    leaveStore: () => {},
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
