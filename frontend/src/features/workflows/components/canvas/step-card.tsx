import {Handle, Position, type Node, type NodeProps} from '@xyflow/react'

import {StateBadge, StateIcon} from '@/shared/components/step-state'
import {StatusChip} from '@/shared/components/status-chip'
import {EffortTag} from '@/shared/components/effort-tag'
import {useElapsed} from '@/shared/hooks/use-elapsed'
import {StepState, type Effort} from '@/shared/lib/enums'
import {ContextStamina} from '@/features/workflows/components/canvas/context-stamina'
import {FailureBubble} from '@/features/workflows/components/canvas/failure-bubble'
import {StepActivity} from '@/features/workflows/components/canvas/step-activity'
import {upstreamOf} from '@/features/workflows/graph'
import {cn} from '@/shared/lib/utils'
import type {Workflow, Step} from '@/features/workflows/types'

type StepNodeData = {
    step: Step
    workflow: Workflow
    unlinkable: boolean
    needsInput: boolean
    effort: Effort | null
}

export type StepNodeType = Node<StepNodeData, 'step'>

export const UNTITLED = 'Untitled step'

const TITLE_CLASSES: Record<StepState, string> = {
    [StepState.Idle]: 'text-state-approval',
    [StepState.Blocked]: 'text-state-approval',
    [StepState.Queued]: 'text-state-approval',
    [StepState.AwaitingApproval]: 'text-state-approval',
    [StepState.AwaitingReview]: 'text-state-approval',
    [StepState.Running]: 'text-info',
    [StepState.Done]: 'text-state-done',
    [StepState.Failed]: 'text-state-failed',
    [StepState.Cancelled]: 'text-muted-foreground',
}

function accentClass(state: StepState) {
    if (state === StepState.Running) return 'bg-live'
    if (state === StepState.Cancelled) return 'bg-state-idle'
    if (state === StepState.AwaitingApproval) return 'bg-state-approval'
    return null
}

function failureTldr(step: Step) {
    if (step.state !== StepState.Failed) return ''
    return step.report?.handoffs.at(-1)?.tldr ?? ''
}

function upstreamLine(upstream: Step[]) {
    if (upstream.length === 1) return `Runs after ${upstream[0].title || UNTITLED}`
    return `Runs after all ${upstream.length} upstream steps`
}

export function StepCard({data, selected}: NodeProps<StepNodeType>) {
    const {step, workflow, unlinkable, needsInput, effort} = data
    const elapsed = useElapsed(step.run?.startedAt, step.run?.finishedAt)

    const upstream = upstreamOf(workflow, step)
    const context = step.run?.context

    const blocked = step.state === StepState.Blocked
    const running = step.state === StepState.Running
    const cancelled = step.state === StepState.Cancelled
    const connectable = !workflow.locked && !unlinkable
    const accent = accentClass(step.state)
    const tldr = failureTldr(step)

    return (
        <div
            className={cn(
                'relative w-[260px] rounded-xl bg-card p-3 shadow-[0_2px_16px_rgba(27,28,30,0.04)] transition-all duration-120',
                blocked
                    ? 'border-[1.75px] border-dashed border-border-strong opacity-60 ring-0'
                    : 'ring-[1.75px] ring-border-strong',
                running && 'bg-live-tint ring-2 ring-live',
                cancelled && 'bg-state-idle-tint ring-[1.75px] ring-state-idle/30',
                selected && 'ring-2 ring-live',
                unlinkable && 'opacity-30',
            )}
        >
            {tldr && <FailureBubble tldr={tldr} />}

            {running && <StepActivity stepId={step.id} />}

            {accent && (
                <span className="pointer-events-none absolute inset-0 overflow-hidden rounded-xl">
                    <span className={cn('absolute inset-y-0 left-0 w-1', accent)} />
                </span>
            )}

            <Handle type="target" position={Position.Left} isConnectable={connectable} />

            <div className="relative flex items-center gap-2">
                <StateIcon state={step.state} />
                <span
                    className={cn(
                        'min-w-0 flex-1 truncate text-lg font-medium',
                        TITLE_CLASSES[step.state],
                        !step.title && 'text-muted-foreground',
                    )}
                >
                    {step.title || UNTITLED}
                </span>
                {elapsed && (
                    <span className="shrink-0 font-mono text-xs text-muted-foreground">
                        {elapsed}
                    </span>
                )}
            </div>

            <div className="relative mt-2 flex items-center justify-between gap-2 text-sm">
                <span className="flex min-w-0 items-center gap-2">
                    {effort && <EffortTag effort={effort} />}
                    {needsInput && <StatusChip tone="attention">Inputs</StatusChip>}
                </span>
                <span className="flex shrink-0 items-center gap-2">
                    {context && <ContextStamina used={context.used} total={context.total} />}
                    <StateBadge state={step.state} />
                </span>
            </div>

            {upstream.length > 0 && (
                <p className="relative mt-1 truncate text-sm text-muted-foreground">
                    {upstreamLine(upstream)}
                </p>
            )}

            <Handle type="source" position={Position.Right} isConnectable={connectable} />
        </div>
    )
}
