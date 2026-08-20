import {useMemo} from 'react'

import {missingInputs} from '@/features/workflows/step-inputs'
import {useRoles} from '@/features/roles/use-roles'
import type {Workflow} from '@/features/workflows/types'

const NO_ENTRIES: ReturnType<typeof missingInputs> = []

/** A started workflow can no longer be filled in, so it reports nothing outstanding. */
export function useMissingInputs(workflow: Workflow | null) {
    const {roles} = useRoles()

    const entries = useMemo(
        () => (workflow && !workflow.started ? missingInputs(workflow, roles) : NO_ENTRIES),
        [workflow, roles],
    )

    const stepIds = useMemo(() => new Set(entries.map((entry) => entry.step.id)), [entries])

    return {entries, stepIds}
}
