import {StatusChip} from '@/shared/components/status-chip'
import {EFFORT_LABELS, type Effort} from '@/shared/lib/enums'

export function EffortTag({effort, className}: {effort: Effort; className?: string}) {
    return (
        <StatusChip tone="outline" className={className}>
            {EFFORT_LABELS[effort]}
        </StatusChip>
    )
}
