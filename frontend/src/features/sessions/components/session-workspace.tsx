import {useState} from 'react'

import {AcceptGateDialog} from '@/features/sessions/components/accept-gate-dialog'
import {ApprovalDialog} from '@/features/approvals/components/approval-dialog'
import {EditLocationsDialog} from '@/features/sessions/components/edit-locations-dialog'
import {GraphCanvas} from '@/features/sessions/components/canvas/graph-canvas'
import {MissingInputsDialog} from '@/features/sessions/components/nodes/missing-inputs-dialog'
import {NewNodeDialog} from '@/features/sessions/components/nodes/new-node-dialog'
import {SessionHeader} from '@/features/sessions/components/canvas/session-header'
import {TaskDialog} from '@/features/sessions/components/nodes/task-dialog'
import {useInterrupts} from '@/features/sessions/use-interrupts'
import {useMissingInputs} from '@/features/sessions/use-missing-inputs'
import {useToggle} from '@/shared/hooks/use-toggle'
import type {Point, TaskDraft} from '@/features/sessions/types'
import type {SessionStore} from '@/features/sessions/use-session-store'

const ORIGIN: Point = {x: 0, y: 0}

export function SessionWorkspace({
    store,
    railOpen,
    onToggleRail,
    onOpenSettings,
}: {
    store: SessionStore
    railOpen: boolean
    onToggleRail: () => void
    onOpenSettings: () => void
}) {
    const {active} = store
    const {session, gate, approval} = useInterrupts(store)
    const missing = useMissingInputs(session)

    const locations = useToggle()
    const blockedRun = useToggle()
    const [newNodeAt, setNewNodeAt] = useState<Point | null>(null)

    const selectedTask = session?.tasks.find((task) => task.id === store.selectedTaskId)

    const run = () => {
        if (missing.entries.length > 0) blockedRun.open()
        else active.start()
    }

    const fillBlockedNode = (taskId: string) => {
        blockedRun.close()
        store.selectTask(taskId)
    }

    const runAnyway = () => {
        blockedRun.close()
        active.start()
    }

    const closeNewNode = () => setNewNodeAt(null)

    const createTask = (draft: TaskDraft) => {
        if (newNodeAt) active.addTask(draft, newNodeAt)
        closeNewNode()
    }

    const closeTask = () => store.selectTask(null)

    const inspecting = selectedTask && selectedTask.id !== gate?.subject.id && !approval

    return (
        <>
            <main className="surface-card flex min-w-0 flex-1 flex-col overflow-hidden ring-1 ring-border-strong">
                <SessionHeader
                    session={session}
                    cancelling={store.cancelling}
                    onRename={active.rename}
                    onEditLocations={locations.open}
                    onFinalize={active.finalize}
                    onRun={run}
                    onCancel={active.cancel}
                    onClone={active.clone}
                    onNewNode={() => setNewNodeAt(ORIGIN)}
                    onOpenSettings={onOpenSettings}
                    railOpen={railOpen}
                    onToggleRail={onToggleRail}
                />

                {session ? (
                    <GraphCanvas
                        session={session}
                        needsInputIds={missing.taskIds}
                        selectedTaskId={store.selectedTaskId}
                        onSelectTask={store.selectTask}
                        onMoveTask={active.moveTask}
                        onConnect={active.connectTasks}
                        onDisconnect={active.disconnectTasks}
                        onNewNode={setNewNodeAt}
                    />
                ) : (
                    <EmptyWorkspace />
                )}
            </main>

            {session && inspecting && selectedTask && (
                <TaskDialog
                    key={selectedTask.id}
                    session={session}
                    task={selectedTask}
                    savingInputs={store.savingTaskInputs}
                    reverting={store.revertingToTask}
                    onSave={(patch) => active.updateTask(selectedTask.id, patch)}
                    onSaveInputs={(values) => active.saveTaskInputs(selectedTask.id, values)}
                    onRevert={() => active.revertToTask(selectedTask.id, closeTask)}
                    onDelete={() => active.removeTask(selectedTask.id)}
                    onClose={closeTask}
                />
            )}

            {blockedRun.on && (
                <MissingInputsDialog
                    entries={missing.entries}
                    onSelectTask={fillBlockedNode}
                    onRunAnyway={runAnyway}
                    onClose={blockedRun.close}
                />
            )}

            {approval && (
                <ApprovalDialog
                    key={approval.subject.id}
                    approval={approval.subject}
                    waiting={approval.waiting}
                    busy={approval.busy}
                    onAnswer={approval.answer}
                    onClose={approval.dismiss}
                />
            )}

            {gate && (
                <AcceptGateDialog
                    key={gate.subject.id}
                    task={gate.subject}
                    waiting={gate.waiting}
                    busy={gate.busy}
                    onAnswer={gate.answer}
                    onClose={gate.dismiss}
                />
            )}

            {newNodeAt && <NewNodeDialog onCreate={createTask} onClose={closeNewNode} />}

            {locations.on && session && !session.finalized && (
                <EditLocationsDialog
                    key={session.id}
                    session={session}
                    onSave={active.setLocations}
                    onClose={locations.close}
                />
            )}
        </>
    )
}

function EmptyWorkspace() {
    return (
        <div className="flex flex-1 items-center justify-center bg-background">
            <p className="text-base text-muted-foreground">
                No session open. Create one to start a chain of tasks.
            </p>
        </div>
    )
}
