import {FileChanges} from '@/features/sessions/components/inspector/file-changes'
import {useTaskDiff} from '@/features/sessions/use-task-diff'
import {Button} from '@/shared/ui/button'
import type {ReactNode} from 'react'

/** Same row the collapsed section draws when it has nothing, so the header never jumps. */
function DiffNote({text, action}: {text: string; action?: ReactNode}) {
    return (
        <div className="flex items-center gap-2 py-1">
            <span className="size-3.5 shrink-0" />
            <span className="micro-label">Files changed</span>
            <span className="min-w-0 flex-1 truncate text-sm text-muted-foreground">{text}</span>
            {action}
        </div>
    )
}

export function TaskDiff({sessionId, taskId}: {sessionId: string; taskId: string}) {
    const {changes, loading, failed, retry} = useTaskDiff(sessionId, taskId)

    if (loading) return <DiffNote text="Reading what this step changed…" />

    if (failed)
        return (
            <DiffNote
                text="Could not read what this step changed."
                action={
                    <Button variant="ghost" size="xs" onClick={retry}>
                        Try again
                    </Button>
                }
            />
        )

    return <FileChanges changes={changes} />
}
