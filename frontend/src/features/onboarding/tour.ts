import type {GlossaryTerm} from '@/shared/lib/glossary'
import type {MessageKey} from '@/shared/lib/i18n'

export const TourStopId = {
    Role: 'role',
    Workflow: 'workflow',
    Step: 'step',
} as const

export type TourStopId = (typeof TourStopId)[keyof typeof TourStopId]

/**
 * A stop either docks inside the surface it explains — `anchor` null, the card
 * renders as a banner in that dialog — or floats beside an element the veil
 * cuts a hole around. The role editor fills the window, so it docks.
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
    workflow: '[data-tour="workflow"]',
    step: '.react-flow__node',
} as const

export const TOUR_STOPS: TourStop[] = [
    {
        id: TourStopId.Role,
        term: 'role',
        title: 'onboarding.tour.role.title',
        body: 'onboarding.tour.role.body',
        anchor: null,
        side: 'bottom',
    },
    {
        id: TourStopId.Workflow,
        term: 'workflow',
        title: 'onboarding.tour.workflow.title',
        body: 'onboarding.tour.workflow.body',
        anchor: TOUR_ANCHOR.workflow,
        side: 'right',
    },
    {
        id: TourStopId.Step,
        term: 'step',
        title: 'onboarding.tour.step.title',
        body: 'onboarding.tour.step.body',
        anchor: TOUR_ANCHOR.step,
        side: 'bottom',
    },
]
