import {StepState} from '@/shared/lib/enums'
import {stepLayers} from '@/features/workflows/graph'
import {cn} from '@/shared/lib/utils'
import type {Workflow, Step} from '@/features/workflows/types'

const MAX_LAYERS = 6

const MAX_DOTS_PER_LAYER = 3

const SPINE_DOT_CLASSES: Record<StepState, string> = {
    [StepState.Idle]: 'bg-state-idle/40',
    [StepState.Blocked]: 'bg-state-idle/40',
    [StepState.Queued]: 'bg-state-idle/40',
    [StepState.Running]: 'bg-state-running',
    [StepState.AwaitingApproval]: 'bg-state-approval',
    [StepState.AwaitingReview]: 'bg-state-approval',
    [StepState.Done]: 'bg-state-done',
    [StepState.Failed]: 'bg-state-failed',
    [StepState.Cancelled]: 'bg-state-idle',
}

function groupByLayer(workflow: Workflow) {
    const layers = stepLayers(workflow)
    const grouped = new Map<number, Step[]>()

    for (const step of workflow.steps) {
        const index = layers.get(step.id) ?? 0
        const bucket = grouped.get(index)
        if (bucket) bucket.push(step)
        else grouped.set(index, [step])
    }

    return [...grouped.entries()].sort(([a], [b]) => a - b).map(([, steps]) => steps)
}

function SpineDot({state}: {state: StepState}) {
    return (
        <span className="relative flex size-[5px]">
            {state === StepState.Running && (
                <span className="absolute -inset-[3px] animate-ping rounded-full bg-state-running/40 motion-reduce:hidden" />
            )}
            <span className={cn('size-[5px] rounded-full', SPINE_DOT_CLASSES[state])} />
        </span>
    )
}

function Overflow({count}: {count: number}) {
    return <span className="font-mono text-xs leading-none text-muted-foreground">+{count}</span>
}

export function WorkflowSpine({workflow}: {workflow: Workflow}) {
    if (workflow.steps.length === 0) {
        return (
            <div className="flex h-3.5 items-center" aria-hidden>
                <span className="h-px w-4 bg-border" />
            </div>
        )
    }

    const layers = groupByLayer(workflow)
    const overflowing = layers.length > MAX_LAYERS
    const shown = overflowing ? layers.slice(0, MAX_LAYERS - 1) : layers

    return (
        <div className="pointer-events-none flex h-3.5 items-center" aria-hidden>
            {shown.map((steps, index) => (
                <div key={index} className="flex shrink-0 items-center">
                    {index > 0 && <span className="h-px w-2 shrink-0 bg-border" />}
                    <div className="flex shrink-0 flex-col items-center gap-[3px]">
                        {steps.slice(0, MAX_DOTS_PER_LAYER).map((step) => (
                            <SpineDot key={step.id} state={step.state} />
                        ))}
                        {steps.length > MAX_DOTS_PER_LAYER && (
                            <Overflow count={steps.length - MAX_DOTS_PER_LAYER} />
                        )}
                    </div>
                </div>
            ))}
            {overflowing && (
                <>
                    <span className="h-px w-2 shrink-0 bg-border" />
                    <Overflow count={layers.length - shown.length} />
                </>
            )}
        </div>
    )
}
