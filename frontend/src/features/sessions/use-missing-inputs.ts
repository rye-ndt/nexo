import {useMemo} from 'react'

import {missingInputs} from '@/features/sessions/task-inputs'
import {useTemplates} from '@/features/templates/use-templates'
import type {Session} from '@/features/sessions/types'

const NO_ENTRIES: ReturnType<typeof missingInputs> = []

/** A started session can no longer be filled in, so it reports nothing outstanding. */
export function useMissingInputs(session: Session | null) {
    const {templates} = useTemplates()

    const entries = useMemo(
        () => (session && !session.started ? missingInputs(session, templates) : NO_ENTRIES),
        [session, templates],
    )

    const taskIds = useMemo(() => new Set(entries.map((entry) => entry.task.id)), [entries])

    return {entries, taskIds}
}
