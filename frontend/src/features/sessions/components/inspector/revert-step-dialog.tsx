import {ConfirmDialog} from '@/shared/components/confirm-dialog'
import {descendantsOf} from '@/features/sessions/graph'
import type {Session, Task} from '@/features/sessions/types'

function undoneAfter(session: Session, task: Task) {
    const later = descendantsOf(session.tasks, task.id)
    const started = session.tasks.filter((other) => later.has(other.id) && other.run).length

    if (started === 0) return 'No step after it has run, so nothing else is undone.'
    if (started === 1) return 'The one step after it is undone and goes back to not started.'

    return `The ${started} steps after it are undone and go back to not started.`
}

export function RevertStepDialog({
    session,
    task,
    busy,
    onConfirm,
    onClose,
}: {
    session: Session
    task: Task
    busy: boolean
    onConfirm: () => void
    onClose: () => void
}) {
    return (
        <ConfirmDialog
            title={`Revert to “${task.title || 'Untitled node'}”?`}
            description={`Files in ${session.workingDir} go back to how this step left them, and anything written after it is lost. ${undoneAfter(session, task)} The run stops, and you start it again when you are ready.`}
            confirmLabel="Revert to this step"
            destructive
            busy={busy}
            onConfirm={onConfirm}
            onClose={onClose}
        />
    )
}
