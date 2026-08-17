import {useState} from 'react'

import {Button} from '@/shared/ui/button'
import {DialogShell} from '@/shared/components/dialog-shell'
import {SessionLocationsFields} from '@/features/sessions/components/session-locations'
import {TaskLevelTag} from '@/shared/components/task-level-tag'
import {pluralize} from '@/shared/lib/format'
import type {Session, SessionLocations, Task} from '@/features/sessions/types'

export function ImportSessionDialog({
    session,
    onImport,
    onClose,
}: {
    session: Session
    onImport: (locations: SessionLocations) => void
    onClose: () => void
}) {
    const [workingDir, setWorkingDir] = useState(session.workingDir ?? '')
    const [contextDir, setContextDir] = useState(session.contextDir ?? '')

    const locations = {workingDir: workingDir.trim(), contextDir: contextDir.trim()}
    const located = locations.workingDir.length > 0 && locations.contextDir.length > 0

    const confirm = () => {
        if (!located) return

        onImport(locations)
        onClose()
    }

    return (
        <DialogShell
            onClose={onClose}
            title="Import session"
            description="Where this session runs on this machine."
            footer={
                <>
                    <span className="flex-1" />
                    <Button variant="outline" size="sm" onClick={onClose}>
                        Cancel
                    </Button>
                    <Button size="sm" disabled={!located} onClick={confirm}>
                        Import session
                    </Button>
                </>
            }
        >
            <FileSummary session={session} />

            <SessionLocationsFields
                workingDir={workingDir}
                contextDir={contextDir}
                onWorkingDirChange={setWorkingDir}
                onContextDirChange={setContextDir}
            />
        </DialogShell>
    )
}

function FileSummary({session}: {session: Session}) {
    return (
        <section className="flex flex-col gap-3 border-b border-border px-4 py-4">
            <span className="micro-label">From the file</span>

            <div className="flex flex-col gap-1">
                <span className="text-base font-medium break-words">{session.name}</span>
                <span className="text-sm text-muted-foreground">
                    {pluralize(session.tasks.length, 'node')}
                </span>
            </div>

            <div className="flex max-h-56 flex-col gap-1 overflow-y-auto">
                {session.tasks.map((task) => (
                    <NodeRow key={task.id} task={task} />
                ))}
            </div>

            <p className="text-sm text-muted-foreground">
                These nodes run on their own: the templates they were built from stay on the machine
                that exported them.
            </p>
        </section>
    )
}

function NodeRow({task}: {task: Task}) {
    return (
        <div className="flex items-center gap-3 rounded-lg border border-border bg-card px-3 py-2">
            <span className="min-w-0 flex-1 truncate text-base">
                {task.title || 'Untitled node'}
            </span>
            {task.spec && <TaskLevelTag taskLevel={task.spec.taskLevel} />}
        </div>
    )
}
