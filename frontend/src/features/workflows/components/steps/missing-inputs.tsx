import {StatusChip} from '@/shared/components/status-chip'
import {tn} from '@/shared/lib/i18n'

export function MissingInputs({count}: {count: number}) {
    if (count === 0) return null

    return (
        <StatusChip tone="attention">
            {tn('step.inputs.missing.one', 'step.inputs.missing.other', count)}
        </StatusChip>
    )
}

export function MissingInputsNote({count}: {count: number}) {
    if (count === 0) return null

    return (
        <span className="min-w-0 truncate text-sm text-muted-foreground">
            {tn('step.inputs.empty.one', 'step.inputs.empty.other', count)}
        </span>
    )
}
