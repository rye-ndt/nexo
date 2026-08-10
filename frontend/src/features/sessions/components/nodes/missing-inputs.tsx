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
            {pluralize(count, 'input')} still empty — fill {count === 1 ? 'it' : 'them'} before you
            run.
        </span>
    )
}
