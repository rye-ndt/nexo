import {useEffect, useRef, useState} from 'react'
import type {Dimensions, NodeChange} from '@xyflow/react'

import type {TaskNodeType} from '@/features/sessions/components/canvas/task-node'
import type {Point, Task} from '@/features/sessions/types'

const NO_DRAG: Record<string, Point> = {}

const NO_SIZES: Record<string, Dimensions> = {}

function collect<T>(
    changes: NodeChange<TaskNodeType>[],
    pick: (change: NodeChange<TaskNodeType>) => readonly [string, T] | null,
) {
    const picked = changes.flatMap((change) => {
        const entry = pick(change)
        return entry ? [entry] : []
    })

    return picked.length > 0 ? Object.fromEntries(picked) : null
}

/** A drag renders from local state and crosses the bridge once, on drop. */
export function useNodeGestures(tasks: Task[], onMoveTask: (taskId: string, at: Point) => void) {
    const [dragPositions, setDragPositions] = useState<Record<string, Point>>(NO_DRAG)
    const [sizes, setSizes] = useState<Record<string, Dimensions>>(NO_SIZES)
    const dragging = useRef(false)

    // Once the store has caught up with the drop, the local override is stale.
    useEffect(() => {
        if (dragging.current) return
        setDragPositions((current) => (current === NO_DRAG ? current : NO_DRAG))
    }, [tasks])

    const trackChanges = (changes: NodeChange<TaskNodeType>[]) => {
        const measured = collect(changes, (change) =>
            change.type === 'dimensions' && change.dimensions
                ? ([change.id, change.dimensions] as const)
                : null,
        )
        if (measured) setSizes((current) => ({...current, ...measured}))

        const moved = collect(changes, (change) =>
            change.type === 'position' && change.position
                ? ([change.id, change.position] as const)
                : null,
        )
        if (moved) setDragPositions((current) => ({...current, ...moved}))
    }

    const lift = () => {
        dragging.current = true
    }

    const drop = (_: unknown, node: TaskNodeType) => {
        dragging.current = false
        onMoveTask(node.id, node.position)
    }

    return {dragPositions, sizes, trackChanges, lift, drop}
}
