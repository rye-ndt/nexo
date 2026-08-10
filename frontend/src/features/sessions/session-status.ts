import type {ChipTone} from '@/shared/components/status-chip'
import {SessionStatus} from '@/shared/lib/enums'

export const SESSION_CHIP_TONES: Record<SessionStatus, ChipTone> = {
    [SessionStatus.Empty]: 'muted',
    [SessionStatus.Draft]: 'muted',
    [SessionStatus.Ready]: 'info',
    [SessionStatus.Running]: 'running',
    [SessionStatus.Done]: 'done',
    [SessionStatus.Failed]: 'failed',
    [SessionStatus.Cancelled]: 'stopped',
}

export const SESSION_TITLE_CLASSES: Record<SessionStatus, string> = {
    [SessionStatus.Empty]: 'text-muted-foreground',
    [SessionStatus.Draft]: 'text-state-approval',
    [SessionStatus.Ready]: 'text-state-approval',
    [SessionStatus.Running]: 'text-info',
    [SessionStatus.Done]: 'text-state-done',
    [SessionStatus.Failed]: 'text-state-failed',
    [SessionStatus.Cancelled]: 'text-muted-foreground',
}
