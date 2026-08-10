import type {ReactNode} from 'react'

import {cn} from '@/shared/lib/utils'

export type ChipTone =
    | 'muted'
    | 'info'
    | 'running'
    | 'attention'
    | 'attention-solid'
    | 'done'
    | 'failed'
    | 'stopped'
    | 'outline'

const TONE_CLASSES: Record<ChipTone, string> = {
    muted: 'bg-state-idle-tint text-muted-foreground',
    info: 'bg-info-tint text-info',
    running: 'bg-state-running-tint text-state-running',
    attention: 'bg-state-approval-tint text-state-approval',
    'attention-solid': 'bg-state-approval text-white',
    done: 'bg-state-done-tint text-state-done',
    failed: 'bg-state-failed-tint text-state-failed',
    stopped: 'bg-state-idle text-white',
    outline: 'border border-border-strong text-muted-foreground',
}

const BASE =
    'inline-flex shrink-0 items-center rounded-sm px-2.5 py-1 text-xs leading-none font-bold tracking-[0.05em] uppercase'

export function StatusChip({
    tone,
    className,
    children,
}: {
    tone: ChipTone
    className?: string
    children: ReactNode
}) {
    return <span className={cn(BASE, TONE_CLASSES[tone], className)}>{children}</span>
}
