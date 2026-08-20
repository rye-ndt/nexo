import {useEffect, useLayoutEffect, useRef, useState} from 'react'
import {createPortal} from 'react-dom'

import {TourCard} from '@/features/onboarding/components/tour-card'
import {haloOf, placeCard, type Side, type Size} from '@/features/onboarding/spotlight'
import {useAnchorRect} from '@/features/onboarding/use-anchor'
import {useTour} from '@/features/onboarding/tour-context'
import {cn} from '@/shared/lib/utils'
import type {Rect} from '@/features/onboarding/types'
import type {TourStop} from '@/features/onboarding/tour'

const BEAK: Record<Side, string> = {
    right: 'top-1/2 -left-1 -translate-y-1/2',
    left: 'top-1/2 -right-1 -translate-y-1/2',
    bottom: '-top-1 left-1/2 -translate-x-1/2',
    top: '-bottom-1 left-1/2 -translate-x-1/2',
}

const ADRIFT = 'top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2'

function useViewport(): Size {
    const [size, setSize] = useState({width: window.innerWidth, height: window.innerHeight})

    useEffect(() => {
        const resize = () => setSize({width: window.innerWidth, height: window.innerHeight})

        window.addEventListener('resize', resize)
        return () => window.removeEventListener('resize', resize)
    }, [])

    return size
}

export function TourLayer() {
    const tour = useTour()
    const rect = useAnchorRect(tour.stop?.anchor ?? null)

    if (!tour.stop?.anchor) return null

    return createPortal(
        <TourSpotlight
            stop={tour.stop}
            rect={rect}
            index={tour.index}
            total={tour.total}
            onNext={tour.next}
            onSkip={tour.skip}
        />,
        document.body,
    )
}

/**
 * With no rect the element this stop explains is not on screen — the rail is
 * collapsed, the workflow is empty, the user deleted it. The card comes to the
 * middle rather than vanishing, because it carries the only way out of the tour.
 */
function TourSpotlight({
    stop,
    rect,
    index,
    total,
    onNext,
    onSkip,
}: {
    stop: TourStop
    rect: Rect | null
    index: number
    total: number
    onNext: () => void
    onSkip: () => void
}) {
    const card = useRef<HTMLDivElement>(null)
    const [size, setSize] = useState<Size | null>(null)
    const viewport = useViewport()

    useLayoutEffect(() => {
        const element = card.current
        if (!element) return

        const measure = () => {
            const {width, height} = element.getBoundingClientRect()
            setSize((current) =>
                current?.width === width && current.height === height ? current : {width, height},
            )
        }

        measure()

        const observer = new ResizeObserver(measure)
        observer.observe(element)
        return () => observer.disconnect()
    }, [])

    const halo = rect ? haloOf(rect) : null
    const placement = rect && size ? placeCard(rect, stop.side, size, viewport) : null

    return (
        <>
            {halo && (
                <div
                    aria-hidden="true"
                    className="veil-around"
                    style={{top: halo.top, left: halo.left, width: halo.width, height: halo.height}}
                />
            )}

            <div
                ref={card}
                style={placement ? {top: placement.top, left: placement.left} : undefined}
                className={cn(
                    'pointer-events-auto fixed z-50 w-[320px] transition-opacity duration-200',
                    !rect && ADRIFT,
                    rect && !placement && 'opacity-0',
                )}
            >
                {placement && (
                    <span
                        aria-hidden="true"
                        className={cn(
                            'absolute -z-10 size-2 rotate-45 rounded-[2px] bg-card ring-1 ring-border-strong',
                            BEAK[placement.side],
                        )}
                    />
                )}

                <TourCard stop={stop} index={index} total={total} onNext={onNext} onSkip={onSkip} />
            </div>
        </>
    )
}
