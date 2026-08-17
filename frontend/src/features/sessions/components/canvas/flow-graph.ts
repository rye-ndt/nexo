import type {Dimensions} from '@xyflow/react'

import {TaskState, type TaskLevel} from '@/shared/lib/enums'
import {UNTITLED, type TaskNodeType} from '@/features/sessions/components/canvas/task-node'
import type {UnlinkEdgeType} from '@/features/sessions/components/canvas/unlink-edge'
import {taskLevelOf} from '@/features/sessions/task-spec'
import type {Point, Session, Task} from '@/features/sessions/types'
import type {Template} from '@/features/templates/types'

export type GraphEdge = UnlinkEdgeType

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

type EdgeContext = {
    session: Session
    selectedEdgeId: string | null
    hoveredEdgeId: string | null
    onHold: (edgeId: string) => void
    onRelease: () => void
    onUnlink: (sourceId: string, targetId: string) => void
}

export function toFlowNodes(context: FlowContext): TaskNodeType[] {
    const {session, selectedTaskId, needsInputIds, unlinkable, levelByTask} = context

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
        // Rearranging the board is always allowed; finalizing freezes the graph, not the layout.
        draggable: true,
        deletable: false,
        className:
            'rounded-xl focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-live',
    }))
}

export function toFlowEdges(context: EdgeContext): GraphEdge[] {
    const {session, selectedEdgeId, hoveredEdgeId, onHold, onRelease, onUnlink} = context
    const byId = new Map(session.tasks.map((task) => [task.id, task]))
    const locked = session.finalized

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
                    type: 'unlink' as const,
                    className: EDGE_TONES[source.state],
                    selected: id === selectedEdgeId,
                    deletable: !locked,
                    data: {
                        revealed: !locked && id === hoveredEdgeId,
                        label: unlinkLabel(source, task),
                        onHold: () => onHold(id),
                        onRelease,
                        onUnlink: () => onUnlink(sourceId, task.id),
                    },
                },
            ]
        }),
    )
}

function unlinkLabel(source: Task, target: Task) {
    return `Unlink ${source.title || UNTITLED} from ${target.title || UNTITLED}`
}

export function taskLevels(tasks: Task[], templates: Template[]) {
    return new Map<string, TaskLevel | null>(
        tasks.map((task) => [task.id, taskLevelOf(task, templates)]),
    )
}
