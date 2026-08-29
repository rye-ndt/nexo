import type {GlossaryTerm} from '@/shared/lib/glossary'
import type {MessageKey} from '@/shared/lib/i18n'

export const TourStopId = {
    Store: 'store',
    Template: 'template',
    Step: 'step',
    Run: 'run',
} as const

export type TourStopId = (typeof TourStopId)[keyof typeof TourStopId]

/** The stops the store hosts, which the tour opens the store for and leaves together. */
export const STORE_STOPS = new Set<TourStopId>([TourStopId.Store, TourStopId.Template])

/**
 * A stop either docks inside the surface it explains — `anchor` null, the card
 * renders as a banner along the bottom of it — or floats beside an element the
 * veil cuts a hole around. The store stop docks: the whole catalog is the
 * subject, and there is nothing smaller to ring.
 */
export type TourStop = {
    id: TourStopId
    term: GlossaryTerm
    title: MessageKey
    body: MessageKey
    anchor: string | null
    side: 'right' | 'bottom'
}

const TOUR_ANCHOR = {
    store: '[data-tour="store"]',
    step: '.react-flow__node',
    primary: '[data-tour="primary"]',
} as const

export const TOUR_STOPS: TourStop[] = [
    {
        id: TourStopId.Store,
        term: 'store',
        title: 'onboarding.tour.store.title',
        body: 'onboarding.tour.store.body',
        anchor: TOUR_ANCHOR.store,
        side: 'right',
    },
    {
        id: TourStopId.Template,
        term: 'workflow',
        title: 'onboarding.tour.template.title',
        body: 'onboarding.tour.template.body',
        anchor: null,
        side: 'bottom',
    },
    {
        id: TourStopId.Step,
        term: 'step',
        title: 'onboarding.tour.step.title',
        body: 'onboarding.tour.step.body',
        anchor: TOUR_ANCHOR.step,
        side: 'bottom',
    },
    {
        id: TourStopId.Run,
        term: 'lock',
        title: 'onboarding.tour.run.title',
        body: 'onboarding.tour.run.body',
        anchor: TOUR_ANCHOR.primary,
        side: 'bottom',
    },
]
