import {useState} from 'react'

import {useWorkflows} from '@/features/workflows/use-workflows'
import * as graph from '@/features/workflows/graph'
import type {InputValue} from '@/features/roles/types'
import type {
    Point,
    Workflow,
    WorkflowDraft,
    WorkflowLocations,
    Step,
    StepDraft,
} from '@/features/workflows/types'

/**
 * Single source of truth for the workflow graph. Workflows are server state and
 * live in the query cache; the active workflow and the selected step are view
 * state and stay local.
 *
 * Rail-level writes name the workflow they touch. Everything under `active` acts
 * on the open workflow and is inert when there is none, so no caller repeats
 * that guard.
 */
export function useWorkflowStore() {
    const store = useWorkflows()
    const {workflows} = store

    const [selectedWorkflowId, setSelectedWorkflowId] = useState<string | null>(null)
    const [selectedStepId, setSelectedStepId] = useState<string | null>(null)

    const activeWorkflow = graph.findWorkflow(workflows, selectedWorkflowId) ?? workflows[0] ?? null
    const activeWorkflowId = activeWorkflow?.id ?? null

    const selectWorkflow = (workflowId: string | null) => {
        setSelectedWorkflowId(workflowId)
        setSelectedStepId(null)
    }

    /** Lifts a write that needs a workflow id into one the open workflow supplies. */
    const onActive =
        <TArgs extends unknown[]>(write: (workflowId: string, ...args: TArgs) => void) =>
        (...args: TArgs) => {
            if (activeWorkflowId) write(activeWorkflowId, ...args)
        }

    /** A draft is the only thing structural edits may touch; a locked graph is frozen. */
    const openWorkflow = (workflowId: string) => {
        const workflow = graph.findWorkflow(workflows, workflowId)
        return workflow && !workflow.locked ? workflow : null
    }

    const addWorkflow = (draft: WorkflowDraft) => {
        const workflowId = crypto.randomUUID()
        store.createWorkflow({workflowId, draft})
        selectWorkflow(workflowId)
    }

    const duplicateWorkflow = (sourceId: string) => {
        const workflowId = crypto.randomUUID()
        store.duplicateWorkflow({sourceId, workflowId})
        selectWorkflow(workflowId)
    }

    const importWorkflow = (workflow: Workflow, locations: WorkflowLocations) => {
        store.importWorkflow({workflow, locations})
        selectWorkflow(workflow.id)
    }

    const reorderWorkflow = (workflowId: string, toIndex: number) => {
        store.reorderWorkflow({workflowId, toIndex})
    }

    const deleteWorkflow = (workflowId: string) => {
        store.deleteWorkflow({workflowId})
        if (selectedWorkflowId === workflowId) selectWorkflow(null)
    }

    const addStep = (workflowId: string, draft: StepDraft, position: Point) => {
        const workflow = openWorkflow(workflowId)
        if (!workflow) return

        const stepId = crypto.randomUUID()
        store.createStep({
            workflowId,
            stepId,
            draft,
            position: graph.freePosition(workflow, position),
        })
        setSelectedStepId(stepId)
    }

    const duplicateStep = (workflowId: string, stepId: string) => {
        const workflow = openWorkflow(workflowId)
        if (!workflow) return

        const source = graph.findStep(workflow, stepId)
        if (!source) return

        const copyId = crypto.randomUUID()
        store.duplicateStep({
            workflowId,
            stepId,
            copyId,
            position: graph.freePosition(workflow, source.position),
        })
        setSelectedStepId(copyId)
    }

    const removeStep = (workflowId: string, stepId: string) => {
        store.deleteStep({workflowId, stepId})
        if (selectedStepId === stepId) setSelectedStepId(null)
    }

    const connectSteps = (workflowId: string, sourceId: string, targetId: string) => {
        const workflow = openWorkflow(workflowId)
        if (!workflow || graph.createsCycle(workflow.steps, sourceId, targetId)) return

        store.connectSteps({workflowId, sourceId, targetId})
    }

    const active = {
        workflow: activeWorkflow,

        rename: onActive((workflowId, name: string) => store.renameWorkflow({workflowId, name})),
        setLocations: onActive((workflowId, locations: WorkflowLocations) =>
            store.setWorkflowLocations({workflowId, ...locations}),
        ),
        lock: onActive((workflowId) => store.lockWorkflow({workflowId})),
        start: onActive((workflowId) => store.startWorkflow({workflowId})),
        pause: onActive((workflowId, onSettled?: () => void) =>
            store.pauseWorkflow({workflowId}, {onSettled}),
        ),
        resume: onActive((workflowId) => store.resumeWorkflow({workflowId})),
        cancel: onActive((workflowId, onSettled?: () => void) =>
            store.cancelWorkflow({workflowId}, {onSettled}),
        ),
        duplicate: onActive((workflowId) => duplicateWorkflow(workflowId)),

        addStep: onActive((workflowId, draft: StepDraft, position: Point) =>
            addStep(workflowId, draft, position),
        ),
        updateStep: onActive((workflowId, stepId: string, patch: Partial<Step>) =>
            store.updateStep({workflowId, stepId, patch}),
        ),
        moveStep: onActive((workflowId, stepId: string, position: Point) =>
            store.moveStep({workflowId, stepId, position}),
        ),
        duplicateStep: onActive((workflowId, stepId: string) => duplicateStep(workflowId, stepId)),
        removeStep: onActive((workflowId, stepId: string) => removeStep(workflowId, stepId)),
        saveStepInputs: onActive((workflowId, stepId: string, values: Record<string, InputValue>) =>
            store.saveStepInputs({workflowId, stepId, values}),
        ),
        answerStepAcceptance: onActive((workflowId, stepId: string, accepted: boolean) =>
            store.answerStepAcceptance({workflowId, stepId, accepted}),
        ),
        revertToStep: onActive((workflowId, stepId: string, onSettled?: () => void) =>
            store.revertToStep({workflowId, stepId}, {onSettled}),
        ),

        connectSteps: onActive((workflowId, sourceId: string, targetId: string) =>
            connectSteps(workflowId, sourceId, targetId),
        ),
        disconnectSteps: onActive((workflowId, sourceId: string, targetId: string) =>
            store.disconnectSteps({workflowId, sourceId, targetId}),
        ),
    }

    return {
        workflows,
        activeWorkflowId,
        active,

        selectedStepId,
        selectWorkflow,
        selectStep: setSelectedStepId,

        addWorkflow,
        duplicateWorkflow,
        importWorkflow,
        reorderWorkflow,
        deleteWorkflow,

        pausing: store.pausingWorkflow,
        resuming: store.resumingWorkflow,
        cancelling: store.cancellingWorkflow,
        savingStepInputs: store.savingStepInputs,
        answeringStepAcceptance: store.answeringStepAcceptance,
        revertingToStep: store.revertingToStep,
    }
}

export type WorkflowStore = ReturnType<typeof useWorkflowStore>
