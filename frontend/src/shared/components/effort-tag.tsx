import {StatusChip} from '@/shared/components/status-chip'
import {EFFORT_LABELS, type Effort} from '@/shared/lib/enums'
import {t} from '@/shared/lib/i18n'

export function EffortTag({effort, className}: {effort: Effort; className?: string}) {
    return (
        <StatusChip tone="outline" className={className}>
            {t(EFFORT_LABELS[effort])}
        </StatusChip>
    )
}
