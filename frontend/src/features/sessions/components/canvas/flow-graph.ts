import type {Dimensions, Edge} from '@xyflow/react'

import {TaskState, type TaskLevel} from '@/shared/lib/enums'
import {cn} from '@/shared/lib/utils'
import type {TaskNodeType} from '@/features/sessions/components/canvas/task-node'
import type {Point, Session, Task} from '@/features/sessions/types'
import type {Template} from '@/features/templates/types'

export type GraphEdge = Edge & {pathOptions?: {borderRadius?: number}}

const EDGE_TONES: Partial<Record<TaskState, string>> = {
    [TaskState.Running]: 'is-live',
    [TaskState.Done]: 'is-done',
}

type FlowContext = {
    session: Session
    selectedTaskId: string | null
    needsInputIds: Set<string>
    /** Ids that would form a cycle if linked to the handle being dragged from. */
    unlinkable: Set<string> | null
    levelByTask: Map<string, TaskLevel | null>
    dragPositions: Record<string, Point>
    sizes: Record<string, Dimensions>
}

export function toFlowNodes(context: FlowContext): TaskNodeType[] {
    const {session, selectedTaskId, needsInputIds, unlinkable, levelByTask} = context
    const locked = session.finalized

    return session.tasks.map((task) => ({
        id: task.id,
        type: 'task' as const,
        position: context.dragPositions[task.id] ?? task.position,
        measured: context.sizes[task.id],
        data: {
            task,
            session,
            unlinkable: unlinkable?.has(task.id) ?? false,
            needsInput: needsInputIds.has(task.id),
            taskLevel: levelByTask.get(task.id) ?? null,
        },
        selected: task.id === selectedTaskId,
        draggable: !locked,
        deletable: false,
        className: cn(
            'rounded-xl focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-live',
            locked && 'is-locked',
        ),
    }))
}

export function toFlowEdges(session: Session, selectedEdgeId: string | null): GraphEdge[] {
    const byId = new Map(session.tasks.map((task) => [task.id, task]))

    return session.tasks.flatMap((task) =>
        task.dependsOn.flatMap((sourceId) => {
            const source = byId.get(sourceId)
            if (!source) return []

            const id = `${sourceId}->${task.id}`

            return [
                {
                    id,
                    source: sourceId,
                    target: task.id,
                    type: 'smoothstep',
                    pathOptions: {borderRadius: 12},
                    className: EDGE_TONES[source.state],
                    selected: id === selectedEdgeId,
                    deletable: !session.finalized,
                },
            ]
        }),
    )
}

export function taskLevels(tasks: Task[], templates: Template[]) {
    const levelOf = new Map(templates.map((template) => [template.id, template.taskLevel]))

    return new Map<string, TaskLevel | null>(
        tasks.map((task) => [task.id, levelOf.get(task.templateId ?? '') ?? null]),
    )
}
