import {TourCard} from '@/features/onboarding/components/tour-card'
import {useTourStop} from '@/features/onboarding/tour-context'
import type {TourStopId} from '@/features/onboarding/tour'

export function TourBanner({stop}: {stop: TourStopId}) {
    const tour = useTourStop(stop)

    if (!tour?.stop) return null

    return (
        <TourCard
            stop={tour.stop}
            index={tour.index}
            total={tour.total}
            onNext={tour.next}
            onSkip={tour.skip}
            className="shrink-0 border-t border-border bg-live-tint/60 px-4 py-3"
        />
    )
}
