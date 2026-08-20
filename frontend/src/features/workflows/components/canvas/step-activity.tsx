import {ActivityBubble} from '@/features/workflows/components/canvas/activity-bubble'
import {useStepActivity} from '@/features/workflows/use-step-activity'

/** Subscribes one running step to its own activity, so a tick repaints the bubble and not the graph. */
export function StepActivity({stepId}: {stepId: string}) {
    const lines = useStepActivity(stepId, true)

    return <ActivityBubble lines={lines} />
}
