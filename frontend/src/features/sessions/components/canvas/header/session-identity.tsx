import {StatusChip} from '@/shared/components/status-chip'
import {SESSION_STATUS_LABELS} from '@/shared/lib/enums'
import {SESSION_CHIP_TONES} from '@/features/sessions/session-status'
import {sessionProgress, sessionStatus} from '@/features/sessions/graph'
import type {Session} from '@/features/sessions/types'

export function SessionIdentity({session}: {session: Session}) {
    const status = sessionStatus(session)
    const {done, total} = sessionProgress(session)

    return (
        <span className="flex shrink-0 items-center gap-3">
            <StatusChip tone={SESSION_CHIP_TONES[status]}>
                {SESSION_STATUS_LABELS[status]}
            </StatusChip>

            <span className="hidden font-mono text-sm text-muted-foreground lg:block">
                {done}/{total}
            </span>
        </span>
    )
}
