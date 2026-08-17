import {SessionClock} from '@/features/sessions/components/canvas/header/session-clock'
import {SessionTokens} from '@/features/sessions/components/canvas/header/session-tokens'
import type {Session} from '@/features/sessions/types'

/**
 * What the run costs, read together: one inset cluster so the numbers stop
 * blending into the session's name and paths.
 */
export function RunMeters({session}: {session: Session}) {
    if (!session.started) return null

    return (
        <span className="hidden h-7 shrink-0 items-center divide-x divide-border-strong rounded-md border border-border bg-muted lg:flex">
            <SessionClock session={session} />
            <SessionTokens session={session} />
        </span>
    )
}
