import {Meter} from '@/features/sessions/components/canvas/header/meter'
import {sessionRunWindow} from '@/features/sessions/graph'
import {useElapsed} from '@/shared/hooks/use-elapsed'
import type {Session} from '@/features/sessions/types'

/** How long the run has been going, and how long it took once it settles. */
export function SessionClock({session}: {session: Session}) {
    const {startedAt, finishedAt} = sessionRunWindow(session)
    const elapsed = useElapsed(startedAt, finishedAt)

    if (!elapsed) return null

    return (
        <Meter
            label="Elapsed"
            value={elapsed}
            hint={finishedAt ? `Ran for ${elapsed}.` : `Running for ${elapsed}.`}
        />
    )
}
