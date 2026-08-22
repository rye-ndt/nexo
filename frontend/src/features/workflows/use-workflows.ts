import {useMutation, useQuery, useQueryClient} from '@tanstack/react-query'

import * as api from '@/features/workflows/api'
import * as graph from '@/features/workflows/graph'
import {StepState} from '@/shared/lib/enums'
import {t, type MessageKey} from '@/shared/lib/i18n'
import type {InputValue} from '@/features/roles/types'
import type {
    Point,
    Workflow,
    WorkflowDraft,
    WorkflowLocations,
    Step,
    StepDraft,
} from '@/features/workflows/types'

const WORKFLOWS_KEY = ['workflows']

const RUN_POLL_MS = 900

type WorkflowEdit = (workflow: Workflow) => Workflow

function editWorkflow(workflows: Workflow[], workflowId: string, edit: WorkflowEdit) {
    return workflows.map((workflow) => (workflow.id === workflowId ? edit(workflow) : workflow))
}

function editDraft(workflows: Workflow[], workflowId: string, edit: WorkflowEdit) {
    return workflows.map((workflow) =>
        workflow.id === workflowId && !workflow.locked ? edit(workflow) : workflow,
    )
}

function useWorkflowMutation<TArgs>(
    action: MessageKey,
    mutationFn: (args: TArgs) => Promise<unknown>,
    optimistic: (workflows: Workflow[], args: TArgs) => Workflow[],
) {
    const queryClient = useQueryClient()

    return useMutation({
        meta: {action: t(action)},
        mutationFn,
        onMutate: async (args: TArgs) => {
            await queryClient.cancelQueries({queryKey: WORKFLOWS_KEY})
            const previous = queryClient.getQueryData<Workflow[]>(WORKFLOWS_KEY) ?? []
            queryClient.setQueryData<Workflow[]>(WORKFLOWS_KEY, optimistic(previous, args))
            return {previous}
        },
        onError: (_error, _args, context) => {
            if (context) queryClient.setQueryData<Workflow[]>(WORKFLOWS_KEY, context.previous)
        },
        onSettled: () => {
            queryClient.invalidateQueries({queryKey: WORKFLOWS_KEY})
        },
    })
}

/** Workflows are server state: every write goes out optimistically and is reconciled by a refetch. */
export function useWorkflows() {
    const {data} = useQuery({
        queryKey: WORKFLOWS_KEY,
        queryFn: api.listWorkflows,
        meta: {action: t('workflow.error.load')},
        refetchInterval: (query) =>
            query.state.data?.some(graph.hasActiveStep) ? RUN_POLL_MS : false,
        refetchIntervalInBackground: true,
    })

    const create = useWorkflowMutation(
        'workflow.error.create',
        (args: {workflowId: string; draft: WorkflowDraft}) =>
            api.createWorkflow(args.workflowId, args.draft),
        (workflows, {workflowId, draft}) => [
            {...graph.createWorkflow(draft), id: workflowId},
            ...workflows,
        ],
    )

    const duplicate = useWorkflowMutation(
        'workflow.error.duplicate',
        (args: {sourceId: string; workflowId: string; copyInputs: boolean}) =>
            api.duplicateWorkflow(args.sourceId, args.workflowId, args.copyInputs),
        (workflows, {sourceId, workflowId, copyInputs}) => {
            const source = graph.findWorkflow(workflows, sourceId)
            if (!source) return workflows

            return [{...graph.duplicateWorkflow(source, copyInputs), id: workflowId}, ...workflows]
        },
    )

    const importWorkflow = useWorkflowMutation(
        'workflow.error.import',
        (args: {workflow: Workflow; locations: WorkflowLocations}) =>
            api.importWorkflow(args.workflow, args.locations),
        (workflows, {workflow, locations}) => [{...workflow, ...locations}, ...workflows],
    )

    const rename = useWorkflowMutation(
        'workflow.error.rename',
        (args: {workflowId: string; name: string}) =>
            api.updateWorkflow(args.workflowId, {name: args.name}),
        (workflows, {workflowId, name}) =>
            editDraft(workflows, workflowId, (workflow) => ({...workflow, name})),
    )

    const setLocations = useWorkflowMutation(
        'workflow.error.locations',
        (args: {workflowId: string; projectDir: string}) =>
            api.updateWorkflow(args.workflowId, {projectDir: args.projectDir}),
        (workflows, {workflowId, projectDir}) =>
            editDraft(workflows, workflowId, (workflow) => ({...workflow, projectDir})),
    )

    const lock = useWorkflowMutation(
        'workflow.error.lock',
        (args: {workflowId: string}) => api.updateWorkflow(args.workflowId, {locked: true}),
        (workflows, {workflowId}) =>
            editWorkflow(workflows, workflowId, (workflow) => ({...workflow, locked: true})),
    )

    const unlock = useWorkflowMutation(
        'workflow.error.unlock',
        (args: {workflowId: string}) => api.updateWorkflow(args.workflowId, {locked: false}),
        (workflows, {workflowId}) => editWorkflow(workflows, workflowId, graph.archiveRun),
    )

    const start = useWorkflowMutation(
        'workflow.error.start',
        (args: {workflowId: string}) => api.updateWorkflow(args.workflowId, {started: true}),
        (workflows, {workflowId}) =>
            editWorkflow(workflows, workflowId, (workflow) => ({...workflow, started: true})),
    )

    const pause = useWorkflowMutation(
        'workflow.error.pause',
        (args: {workflowId: string}) => api.pauseWorkflow(args.workflowId),
        (workflows, {workflowId}) => editWorkflow(workflows, workflowId, graph.pauseRun),
    )

    const resume = useWorkflowMutation(
        'workflow.error.resume',
        (args: {workflowId: string}) => api.resumeWorkflow(args.workflowId),
        (workflows, {workflowId}) => editWorkflow(workflows, workflowId, graph.resumeRun),
    )

    const cancel = useWorkflowMutation(
        'workflow.error.cancel',
        (args: {workflowId: string}) => api.cancelWorkflow(args.workflowId),
        (workflows, {workflowId}) => editWorkflow(workflows, workflowId, graph.cancelRun),
    )

    const reorder = useWorkflowMutation(
        'workflow.error.reorder',
        (args: {workflowId: string; toIndex: number}) =>
            api.reorderWorkflow(args.workflowId, args.toIndex),
        (workflows, {workflowId, toIndex}) => graph.moveWorkflow(workflows, workflowId, toIndex),
    )

    const remove = useWorkflowMutation(
        'workflow.error.delete',
        (args: {workflowId: string}) => api.deleteWorkflow(args.workflowId),
        (workflows, {workflowId}) => workflows.filter((workflow) => workflow.id !== workflowId),
    )

    const createStep = useWorkflowMutation(
        'workflow.error.addStep',
        (args: {workflowId: string; stepId: string; draft: StepDraft; position: Point}) =>
            api.createStep(args.workflowId, args.stepId, args.draft, args.position),
        (workflows, {workflowId, stepId, draft, position}) =>
            editDraft(workflows, workflowId, (workflow) =>
                graph.withStep(workflow, {...graph.createStep(draft, position), id: stepId}),
            ),
    )

    const duplicateStep = useWorkflowMutation(
        'workflow.error.duplicateStep',
        (args: {workflowId: string; stepId: string; copyId: string; position: Point}) =>
            api.duplicateStep(args.workflowId, args.stepId, args.copyId, args.position),
        (workflows, {workflowId, stepId, copyId, position}) =>
            editDraft(workflows, workflowId, (workflow) => {
                const source = graph.findStep(workflow, stepId)
                if (!source) return workflow

                return graph.withStep(workflow, graph.copyStep(source, copyId, position))
            }),
    )

    const updateStep = useWorkflowMutation(
        'workflow.error.saveStep',
        (args: {workflowId: string; stepId: string; patch: Partial<Step>}) =>
            api.updateStep(args.workflowId, args.stepId, args.patch),
        (workflows, {workflowId, stepId, patch}) =>
            editDraft(workflows, workflowId, (workflow) =>
                graph.withStepPatch(workflow, stepId, patch),
            ),
    )

    // Layout is a view concern, not part of the graph, so it stays editable on a
    // locked workflow — hence editWorkflow rather than editDraft.
    const moveStep = useWorkflowMutation(
        'workflow.error.moveStep',
        (args: {workflowId: string; stepId: string; position: Point}) =>
            api.moveStep(args.workflowId, args.stepId, args.position),
        (workflows, {workflowId, stepId, position}) =>
            editWorkflow(workflows, workflowId, (workflow) =>
                graph.withStepPatch(workflow, stepId, {position}),
            ),
    )

    const saveInputs = useWorkflowMutation(
        'workflow.error.saveInputs',
        (args: {workflowId: string; stepId: string; values: Record<string, InputValue>}) =>
            api.setStepInputs(args.workflowId, args.stepId, args.values),
        (workflows, {workflowId, stepId, values}) =>
            editWorkflow(workflows, workflowId, (workflow) =>
                graph.withStepPatch(workflow, stepId, {
                    values: {...graph.findStep(workflow, stepId)?.values, ...values},
                }),
            ),
    )

    const answerAcceptance = useWorkflowMutation(
        'workflow.error.decision',
        (args: {workflowId: string; stepId: string; accepted: boolean}) =>
            api.answerStepAcceptance(args.workflowId, args.stepId, args.accepted),
        (workflows, {workflowId, stepId, accepted}) =>
            editWorkflow(workflows, workflowId, (workflow) =>
                graph.withStepPatch(workflow, stepId, {
                    state: accepted ? StepState.Done : StepState.Failed,
                }),
            ),
    )

    const revert = useWorkflowMutation(
        'workflow.error.revert',
        (args: {workflowId: string; stepId: string}) =>
            api.revertWorkflowTo(args.workflowId, args.stepId),
        (workflows, {workflowId, stepId}) =>
            editWorkflow(workflows, workflowId, (workflow) => graph.rewindTo(workflow, stepId)),
    )

    const deleteStep = useWorkflowMutation(
        'workflow.error.deleteStep',
        (args: {workflowId: string; stepId: string}) =>
            api.deleteStep(args.workflowId, args.stepId),
        (workflows, {workflowId, stepId}) =>
            editDraft(workflows, workflowId, (workflow) => graph.withoutStep(workflow, stepId)),
    )

    const connect = useWorkflowMutation(
        'workflow.error.connect',
        (args: {workflowId: string; sourceId: string; targetId: string}) =>
            api.addDependency(args.workflowId, args.sourceId, args.targetId),
        (workflows, {workflowId, sourceId, targetId}) =>
            editDraft(workflows, workflowId, (workflow) =>
                graph.withDependency(workflow, sourceId, targetId),
            ),
    )

    const disconnect = useWorkflowMutation(
        'workflow.error.disconnect',
        (args: {workflowId: string; sourceId: string; targetId: string}) =>
            api.removeDependency(args.workflowId, args.sourceId, args.targetId),
        (workflows, {workflowId, sourceId, targetId}) =>
            editDraft(workflows, workflowId, (workflow) =>
                graph.withoutDependency(workflow, sourceId, targetId),
            ),
    )

    return {
        workflows: data ?? [],
        createWorkflow: create.mutate,
        duplicateWorkflow: duplicate.mutate,
        importWorkflow: importWorkflow.mutate,
        renameWorkflow: rename.mutate,
        setWorkflowLocations: setLocations.mutate,
        lockWorkflow: lock.mutate,
        unlockWorkflow: unlock.mutate,
        startWorkflow: start.mutate,
        startingWorkflow: start.isPending,
        pauseWorkflow: pause.mutate,
        pausingWorkflow: pause.isPending,
        resumeWorkflow: resume.mutate,
        resumingWorkflow: resume.isPending,
        cancelWorkflow: cancel.mutate,
        cancellingWorkflow: cancel.isPending,
        reorderWorkflow: reorder.mutate,
        deleteWorkflow: remove.mutate,
        createStep: createStep.mutate,
        duplicateStep: duplicateStep.mutate,
        updateStep: updateStep.mutate,
        moveStep: moveStep.mutate,
        saveStepInputs: saveInputs.mutate,
        savingStepInputs: saveInputs.isPending,
        answerStepAcceptance: answerAcceptance.mutate,
        answeringStepAcceptance: answerAcceptance.isPending,
        revertToStep: revert.mutate,
        revertingToStep: revert.isPending,
        deleteStep: deleteStep.mutate,
        connectSteps: connect.mutate,
        disconnectSteps: disconnect.mutate,
    }
}
