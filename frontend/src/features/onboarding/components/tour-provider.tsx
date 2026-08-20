import {useCallback, useEffect, useMemo, useState, type ReactNode} from 'react'

import {TOUR_STOPS, TourStopId} from '@/features/onboarding/tour'
import {TourContext} from '@/features/onboarding/tour-context'

const REVEAL_MS = 700

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
    const [revealed, setRevealed] = useState(false)

    const stop = active ? (TOUR_STOPS[index] ?? null) : null
    const onRoleStop = stop?.id === TourStopId.Role

    useEffect(() => {
        if (!onRoleStop) return

        const timer = setTimeout(() => setRevealed(true), REVEAL_MS)
        return () => clearTimeout(timer)
    }, [onRoleStop])

    const next = useCallback(() => {
        if (index + 1 >= TOUR_STOPS.length) return onDone()
        setIndex(index + 1)
    }, [index, onDone])

    const value = useMemo(
        () => ({
            stop,
            index,
            total: TOUR_STOPS.length,
            openRole: onRoleStop && revealed,
            next,
            skip: onDone,
        }),
        [stop, index, onRoleStop, revealed, next, onDone],
    )

    return <TourContext value={value}>{children}</TourContext>
}
