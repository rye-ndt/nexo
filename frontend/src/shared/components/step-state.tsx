import {Check, Pause, Square, X} from 'lucide-react'
import type {ComponentType} from 'react'

import {StatusChip, type ChipTone} from '@/shared/components/status-chip'
import {StepState, STEP_STATE_LABELS} from '@/shared/lib/enums'
import {t} from '@/shared/lib/i18n'
import {cn} from '@/shared/lib/utils'

const STATE_TONES: Record<StepState, ChipTone> = {
    [StepState.Idle]: 'muted',
    [StepState.Blocked]: 'muted',
    [StepState.Queued]: 'info',
    [StepState.Running]: 'running',
    [StepState.AwaitingApproval]: 'attention',
    [StepState.AwaitingReview]: 'attention-solid',
    [StepState.Done]: 'done',
    [StepState.Failed]: 'failed',
    [StepState.Cancelled]: 'stopped',
}

export function StateBadge({state, className}: {state: StepState; className?: string}) {
    return (
        <StatusChip tone={STATE_TONES[state]} className={className}>
            {t(STEP_STATE_LABELS[state])}
        </StatusChip>
    )
}

type IconProps = {className?: string}

function Disc({
    fill,
    glyph: Glyph,
    glyphClassName,
    className,
}: IconProps & {
    fill: string
    glyph: ComponentType<{className?: string; strokeWidth?: number}>
    glyphClassName: string
}) {
    return (
        <span
            aria-hidden="true"
            className={cn(
                'flex size-3.5 shrink-0 items-center justify-center rounded-full',
                fill,
                className,
            )}
        >
            <Glyph className={glyphClassName} strokeWidth={3.5} />
        </span>
    )
}

function RunningIcon({className}: IconProps) {
    return (
        <svg
            viewBox="0 0 16 16"
            fill="none"
            aria-hidden="true"
            className={cn('size-3.5 shrink-0 text-live', className)}
        >
            <circle
                cx="8"
                cy="8"
                r="6"
                stroke="currentColor"
                strokeOpacity="0.15"
                strokeWidth="2"
            />
            <circle
                cx="8"
                cy="8"
                r="6"
                stroke="currentColor"
                strokeWidth="2"
                strokeLinecap="round"
                strokeDasharray="9.4 28.3"
                className="spin-arc"
            />
        </svg>
    )
}

function DoneIcon(props: IconProps) {
    return (
        <Disc {...props} fill="bg-state-done" glyph={Check} glyphClassName="size-2.5 text-white" />
    )
}

function FailedIcon(props: IconProps) {
    return <Disc {...props} fill="bg-state-failed" glyph={X} glyphClassName="size-2.5 text-white" />
}

function CancelledIcon(props: IconProps) {
    return (
        <Disc
            {...props}
            fill="bg-state-idle"
            glyph={Square}
            glyphClassName="size-1.5 fill-white text-white"
        />
    )
}

function AwaitingAcceptIcon(props: IconProps) {
    return (
        <Disc
            {...props}
            fill="bg-state-approval"
            glyph={Pause}
            glyphClassName="size-2 fill-white text-white"
        />
    )
}

function AwaitingApprovalIcon({className}: IconProps) {
    return (
        <span
            aria-hidden="true"
            className={cn(
                'flex size-3.5 shrink-0 items-center justify-center rounded-full bg-state-approval text-[0.5625rem] leading-none font-bold text-white',
                className,
            )}
        >
            !
        </span>
    )
}

const RING_CLASSES: Partial<Record<StepState, string>> = {
    [StepState.Queued]: 'border-info/40',
    [StepState.Blocked]: 'border-dashed border-muted-foreground/50',
    [StepState.Idle]: 'border-muted-foreground/40',
}

function ringIcon(state: StepState) {
    return function RingIcon({className}: IconProps) {
        return (
            <span
                aria-hidden="true"
                className={cn(
                    'block size-3.5 shrink-0 rounded-full border-[1.5px]',
                    RING_CLASSES[state],
                    className,
                )}
            />
        )
    }
}

const STATE_ICONS: Record<StepState, ComponentType<IconProps>> = {
    [StepState.Idle]: ringIcon(StepState.Idle),
    [StepState.Blocked]: ringIcon(StepState.Blocked),
    [StepState.Queued]: ringIcon(StepState.Queued),
    [StepState.Running]: RunningIcon,
    [StepState.AwaitingApproval]: AwaitingApprovalIcon,
    [StepState.AwaitingReview]: AwaitingAcceptIcon,
    [StepState.Done]: DoneIcon,
    [StepState.Failed]: FailedIcon,
    [StepState.Cancelled]: CancelledIcon,
}

export function StateIcon({state, className}: {state: StepState; className?: string}) {
    const Icon = STATE_ICONS[state]
    return <Icon className={className} />
}
