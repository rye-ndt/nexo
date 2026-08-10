import {StatusChip} from '@/shared/components/status-chip'
import {TASK_LEVEL_LABELS, type TaskLevel} from '@/shared/lib/enums'

export function TaskLevelTag({taskLevel, className}: {taskLevel: TaskLevel; className?: string}) {
    return (
        <StatusChip tone="outline" className={className}>
            {TASK_LEVEL_LABELS[taskLevel]}
        </StatusChip>
    )
}
