import {StatusChip} from '@/shared/components/status-chip'
import {pluralize} from '@/shared/lib/format'

export function MissingInputs({count}: {count: number}) {
    if (count === 0) return null

    return <StatusChip tone="attention">{pluralize(count, 'required input')} left</StatusChip>
}

export function MissingInputsNote({count}: {count: number}) {
    if (count === 0) return null

    return (
        <span className="text-sm text-muted-foreground">
            {pluralize(count, 'input')} still empty — the run goes ahead with the prompt as written.
        </span>
    )
}
