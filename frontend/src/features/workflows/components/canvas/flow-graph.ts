import type {Dimensions} from '@xyflow/react'

import {StepState, type Effort} from '@/shared/lib/enums'
import type {StepNodeType} from '@/features/workflows/components/canvas/step-card'
import type {UnlinkEdgeType} from '@/features/workflows/components/canvas/unlink-edge'
import {effortOf} from '@/features/workflows/step-spec'
import {t} from '@/shared/lib/i18n'
import type {Point, Workflow, Step} from '@/features/workflows/types'
import type {Role} from '@/features/roles/types'

export type GraphEdge = UnlinkEdgeType

const EDGE_TONES: Partial<Record<StepState, string>> = {
    [StepState.Running]: 'is-live',
    [StepState.Done]: 'is-done',
}

type FlowContext = {
    workflow: Workflow
    selectedStepId: string | null
    needsInputIds: Set<string>
    /** Ids that would form a cycle if linked to the handle being dragged from. */
    unlinkable: Set<string> | null
    effortByStep: Map<string, Effort | null>
    dragPositions: Record<string, Point>
    sizes: Record<string, Dimensions>
}

type EdgeContext = {
    workflow: Workflow
    selectedEdgeId: string | null
    hoveredEdgeId: string | null
    onHold: (edgeId: string) => void
    onRelease: () => void
    onUnlink: (sourceId: string, targetId: string) => void
}

export function toFlowNodes(context: FlowContext): StepNodeType[] {
    const {workflow, selectedStepId, needsInputIds, unlinkable, effortByStep} = context

    return workflow.steps.map((step) => ({
        id: step.id,
        type: 'step' as const,
        position: context.dragPositions[step.id] ?? step.position,
        measured: context.sizes[step.id],
        data: {
            step,
            workflow,
            unlinkable: unlinkable?.has(step.id) ?? false,
            needsInput: needsInputIds.has(step.id),
            effort: effortByStep.get(step.id) ?? null,
        },
        selected: step.id === selectedStepId,
        // Rearranging the board is always allowed; locking freezes the graph, not the layout.
        draggable: true,
        deletable: false,
        className:
            'rounded-xl focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-live',
    }))
}

export function toFlowEdges(context: EdgeContext): GraphEdge[] {
    const {workflow, selectedEdgeId, hoveredEdgeId, onHold, onRelease, onUnlink} = context
    const byId = new Map(workflow.steps.map((step) => [step.id, step]))
    const locked = workflow.locked

    return workflow.steps.flatMap((step) =>
        step.dependsOn.flatMap((sourceId) => {
            const source = byId.get(sourceId)
            if (!source) return []

            const id = `${sourceId}->${step.id}`

            return [
                {
                    id,
                    source: sourceId,
                    target: step.id,
                    type: 'unlink' as const,
                    className: EDGE_TONES[source.state],
                    selected: id === selectedEdgeId,
                    deletable: !locked,
                    data: {
                        revealed: !locked && id === hoveredEdgeId,
                        label: unlinkLabel(source, step),
                        onHold: () => onHold(id),
                        onRelease,
                        onUnlink: () => onUnlink(sourceId, step.id),
                    },
                },
            ]
        }),
    )
}

function unlinkLabel(source: Step, target: Step) {
    return t('canvas.edge.unlink', {
        source: source.title || t('step.untitled'),
        target: target.title || t('step.untitled'),
    })
}

export function efforts(steps: Step[], roles: Role[]) {
    return new Map<string, Effort | null>(steps.map((step) => [step.id, effortOf(step, roles)]))
}
