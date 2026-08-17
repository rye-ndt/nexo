import {SlidersHorizontal} from 'lucide-react'

import {TaskLevelTag} from '@/shared/components/task-level-tag'
import {type TaskLevel} from '@/shared/lib/enums'

const CHOSEN_UNDER = 'Which agent runs this level is set under Settings → Preferences.'

function explain(taskLevel: TaskLevel | null | undefined, fromTemplate: boolean) {
    if (!taskLevel) return 'Set once this node has a template.'
    if (fromTemplate) return `Inherited from the template. ${CHOSEN_UNDER}`

    return `Set when this session was exported. ${CHOSEN_UNDER}`
}

export function InheritedAgent({
    taskLevel,
    fromTemplate,
}: {
    taskLevel: TaskLevel | null | undefined
    fromTemplate: boolean
}) {
    return (
        <div className="flex flex-col gap-2 border-b border-border px-4 py-3">
            <div className="flex items-center justify-between gap-3">
                <span className="micro-label">Task level</span>
                <span className="flex min-w-0 items-center gap-2 text-muted-foreground">
                    <SlidersHorizontal className="size-3 shrink-0" aria-hidden="true" />
                    {taskLevel ? (
                        <TaskLevelTag taskLevel={taskLevel} />
                    ) : (
                        <span className="truncate text-base text-foreground">Not set</span>
                    )}
                </span>
            </div>
            <p className="text-sm text-muted-foreground">{explain(taskLevel, fromTemplate)}</p>
        </div>
    )
}
