import {useState} from 'react'

import {ReviewDialog} from '@/features/workflows/components/review-dialog'
import {ApprovalDialog} from '@/features/approvals/components/approval-dialog'
import {ConfirmDialog} from '@/shared/components/confirm-dialog'
import {EditLocationsDialog} from '@/features/workflows/components/edit-locations-dialog'
import {GraphCanvas} from '@/features/workflows/components/canvas/graph-canvas'
import {MissingInputsDialog} from '@/features/workflows/components/steps/missing-inputs-dialog'
import {NewStepDialog} from '@/features/workflows/components/steps/new-step-dialog'
import {WorkflowHeader} from '@/features/workflows/components/canvas/workflow-header'
import {StepDialog} from '@/features/workflows/components/steps/step-dialog'
import {draftContext} from '@/features/workflows/graph'
import {useInterrupts} from '@/features/workflows/use-interrupts'
import {useMissingInputs} from '@/features/workflows/use-missing-inputs'
import {useToggle} from '@/shared/hooks/use-toggle'
import type {Point, StepDraft} from '@/features/workflows/types'
import type {WorkflowStore} from '@/features/workflows/use-workflow-store'

const ORIGIN: Point = {x: 0, y: 0}

export function WorkflowWorkspace({
    store,
    railOpen,
    onToggleRail,
    onOpenSettings,
}: {
    store: WorkflowStore
    railOpen: boolean
    onToggleRail: () => void
    onOpenSettings: () => void
}) {
    const {active} = store
    const {workflow, gate, approval} = useInterrupts(store)
    const missing = useMissingInputs(workflow)

    const locations = useToggle()
    const blockedRun = useToggle()
    const [newStepAt, setNewStepAt] = useState<Point | null>(null)
    const [deletingStepId, setDeletingStepId] = useState<string | null>(null)

    const selectedStep = workflow?.steps.find((step) => step.id === store.selectedStepId)
    const deletingStep = workflow?.steps.find((step) => step.id === deletingStepId)

    const run = () => {
        if (missing.entries.length > 0) blockedRun.open()
        else active.start()
    }

    const fillBlockedStep = (stepId: string) => {
        blockedRun.close()
        store.selectStep(stepId)
    }

    const runAnyway = () => {
        blockedRun.close()
        active.start()
    }

    const closeNewStep = () => setNewStepAt(null)

    const createStep = (draft: StepDraft) => {
        if (newStepAt) active.addStep(draft, newStepAt)
        closeNewStep()
    }

    const closeStep = () => store.selectStep(null)

    const stopDeleting = () => setDeletingStepId(null)

    const deleteStep = () => {
        if (deletingStepId) active.removeStep(deletingStepId)
        stopDeleting()
    }

    const inspecting = selectedStep && selectedStep.id !== gate?.subject.id && !approval

    return (
        <>
            <main className="surface-card flex min-w-0 flex-1 flex-col overflow-hidden ring-1 ring-border-strong">
                <WorkflowHeader
                    workflow={workflow}
                    cancelling={store.cancelling}
                    pausing={store.pausing}
                    onRename={active.rename}
                    onEditLocations={locations.open}
                    onLock={active.lock}
                    onRun={run}
                    onPause={active.pause}
                    onResume={active.resume}
                    onCancel={active.cancel}
                    onDuplicate={active.duplicate}
                    onNewStep={() => setNewStepAt(ORIGIN)}
                    onOpenSettings={onOpenSettings}
                    railOpen={railOpen}
                    onToggleRail={onToggleRail}
                />

                {workflow ? (
                    <GraphCanvas
                        workflow={workflow}
                        needsInputIds={missing.stepIds}
                        selectedStepId={store.selectedStepId}
                        onSelectStep={store.selectStep}
                        onMoveStep={active.moveStep}
                        onConnect={active.connectSteps}
                        onDisconnect={active.disconnectSteps}
                        onNewStep={setNewStepAt}
                        onDuplicateStep={active.duplicateStep}
                        onDeleteStep={setDeletingStepId}
                    />
                ) : (
                    <EmptyWorkspace />
                )}
            </main>

            {workflow && inspecting && selectedStep && (
                <StepDialog
                    key={selectedStep.id}
                    workflow={workflow}
                    step={selectedStep}
                    savingInputs={store.savingStepInputs}
                    reverting={store.revertingToStep}
                    onSave={(patch) => active.updateStep(selectedStep.id, patch)}
                    onSaveInputs={(values) => active.saveStepInputs(selectedStep.id, values)}
                    onRevert={() => active.revertToStep(selectedStep.id, closeStep)}
                    onDelete={() => active.removeStep(selectedStep.id)}
                    onClose={closeStep}
                />
            )}

            {blockedRun.on && (
                <MissingInputsDialog
                    entries={missing.entries}
                    onSelectStep={fillBlockedStep}
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
                <ReviewDialog
                    key={gate.subject.id}
                    step={gate.subject}
                    waiting={gate.waiting}
                    busy={gate.busy}
                    onAnswer={gate.answer}
                    onClose={gate.dismiss}
                />
            )}

            {newStepAt && workflow && (
                <NewStepDialog
                    context={draftContext(workflow)}
                    onCreate={createStep}
                    onClose={closeNewStep}
                />
            )}

            {deletingStep && (
                <ConfirmDialog
                    title={`Delete “${deletingStep.title || 'this step'}”?`}
                    description="Its prompt, inputs and any result it produced go with it. This cannot be undone."
                    confirmLabel="Delete step"
                    destructive
                    onConfirm={deleteStep}
                    onClose={stopDeleting}
                />
            )}

            {locations.on && workflow && !workflow.locked && (
                <EditLocationsDialog
                    key={workflow.id}
                    workflow={workflow}
                    onSave={active.setLocations}
                    onClose={locations.close}
                />
            )}
        </>
    )
}

function EmptyWorkspace() {
    return (
        <div className="flex flex-1 flex-col items-center justify-center gap-3 bg-background">
            <img
                src="/empty_bg.png"
                alt=""
                aria-hidden
                draggable={false}
                className="mb-4 w-[min(60%,34rem)] max-w-none opacity-35 select-none dark:opacity-45 [mask-image:radial-gradient(ellipse_at_center,black_45%,transparent_85%)]"
            />
            <p className="text-base text-muted-foreground">
                No workflow open. Create one to start a chain of steps.
            </p>
        </div>
    )
}
