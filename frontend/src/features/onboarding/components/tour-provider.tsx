import {useCallback, useMemo, useState, type ReactNode} from 'react'

import {STORE_STOPS, TOUR_STOPS} from '@/features/onboarding/tour'
import {TourContext} from '@/features/onboarding/tour-context'

export function TourProvider({
    active,
    onDone,
    children,
}: {
    active: boolean
    onDone: () => void
    children: ReactNode
}) {
    const [index, setIndex] = useState(0)

    const stop = active ? (TOUR_STOPS[index] ?? null) : null

    const next = useCallback(() => {
        if (index + 1 >= TOUR_STOPS.length) return onDone()
        setIndex(index + 1)
    }, [index, onDone])

    /**
     * The store has said everything it has to say — the user added something, or
     * navigated out. Skips whichever of its stops are still ahead, so leaving from
     * the first one does not strand the tour on a surface the app has moved off.
     */
    const leaveStore = useCallback(() => {
        if (!active) return

        let ahead = index
        while (ahead < TOUR_STOPS.length && STORE_STOPS.has(TOUR_STOPS[ahead].id)) ahead += 1

        if (ahead >= TOUR_STOPS.length) return onDone()
        setIndex(ahead)
    }, [active, index, onDone])

    const value = useMemo(
        () => ({
            stop,
            index,
            total: TOUR_STOPS.length,
            openStore: Boolean(stop && STORE_STOPS.has(stop.id)),
            next,
            leaveStore,
            skip: onDone,
        }),
        [stop, index, next, leaveStore, onDone],
    )

    return <TourContext value={value}>{children}</TourContext>
}
