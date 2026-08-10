import {TaskStatusDialog} from '@/features/sessions/components/inspector/task-status-dialog'
import {EditNodeDialog} from '@/features/sessions/components/nodes/edit-node-dialog'
import {NodeInputsDialog} from '@/features/sessions/components/nodes/node-inputs-dialog'
import type {Session, Task} from '@/features/sessions/types'
import type {ParamValue} from '@/features/templates/types'

export function TaskDialog({
    session,
    task,
    savingInputs,
    reverting,
    onSave,
    onSaveInputs,
    onRevert,
    onDelete,
    onClose,
}: {
    session: Session
    task: Task
    savingInputs: boolean
    reverting: boolean
    onSave: (patch: Partial<Task>) => void
    onSaveInputs: (values: Record<string, ParamValue>) => void
    onRevert: () => void
    onDelete: () => void
    onClose: () => void
}) {
    if (session.started)
        return (
            <TaskStatusDialog
                session={session}
                task={task}
                reverting={reverting}
                onRevert={onRevert}
                onClose={onClose}
            />
        )

    if (session.finalized)
        return (
            <NodeInputsDialog
                task={task}
                busy={savingInputs}
                onSave={onSaveInputs}
                onClose={onClose}
            />
        )

    return <EditNodeDialog task={task} onSave={onSave} onDelete={onDelete} onClose={onClose} />
}
