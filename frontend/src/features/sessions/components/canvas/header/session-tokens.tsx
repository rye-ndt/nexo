import {Meter} from '@/features/sessions/components/canvas/header/meter'
import {formatTokens} from '@/shared/lib/format'
import type {Session} from '@/features/sessions/types'

/** The benchmark reading: what the whole run cost, ticking up while it runs. */
export function SessionTokens({session}: {session: Session}) {
    const tokens = session.tokensUsed ?? 0

    return (
        <Meter
            label="Spent"
            value={formatTokens(tokens)}
            hint={`${tokens.toLocaleString()} tokens spent across every node in this run. Retries count too, so this only goes up.`}
        />
    )
}
