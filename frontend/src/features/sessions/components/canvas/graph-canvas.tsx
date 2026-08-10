import {useMemo, useState, type MouseEvent} from 'react'
import {
    Background,
    BackgroundVariant,
    ReactFlow,
    ReactFlowProvider,
    useReactFlow,
    type Connection,
    type Edge,
    type HandleType,
    type OnConnectStartParams,
} from '@xyflow/react'

import {EmptyCanvas} from '@/features/sessions/components/canvas/empty-canvas'
import {ZoomCluster} from '@/features/sessions/components/canvas/zoom-cluster'
import {TaskNode} from '@/features/sessions/components/canvas/task-node'
import {
    taskLevels,
    toFlowEdges,
    toFlowNodes,
    type GraphEdge,
} from '@/features/sessions/components/canvas/flow-graph'
import {useNodeGestures} from '@/features/sessions/components/canvas/use-node-gestures'
import {
    CANVAS_CHROME,
    FIT_VIEW_OPTIONS,
    MAX_ZOOM,
    MIN_ZOOM,
    NODE_CURSOR_OFFSET,
    ORIGIN,
} from '@/features/sessions/components/canvas/view'
import {useTemplates} from '@/features/templates/use-templates'
import {createsCycle, unlinkableFrom, unlinkableInto} from '@/features/sessions/graph'
import type {TaskNodeType} from '@/features/sessions/components/canvas/task-node'
import type {Point, Session} from '@/features/sessions/types'

type GraphCanvasProps = {
    session: Session
    selectedTaskId: string | null
    needsInputIds: Set<string>
    onSelectTask: (taskId: string | null) => void
    onMoveTask: (taskId: string, position: Point) => void
    onConnect: (sourceId: string, targetId: string) => void
    onDisconnect: (sourceId: string, targetId: string) => void
    onNewNode: (position: Point) => void
}

type ConnectionOrigin = {nodeId: string; handleType: HandleType}

const nodeTypes = {task: TaskNode}

export function GraphCanvas(props: GraphCanvasProps) {
    return (
        <ReactFlowProvider>
            <Canvas {...props} />
        </ReactFlowProvider>
    )
}

function Canvas({
    session,
    selectedTaskId,
    needsInputIds,
    onSelectTask,
    onMoveTask,
    onConnect,
    onDisconnect,
    onNewNode,
}: GraphCanvasProps) {
    const {screenToFlowPosition} = useReactFlow()
    const {templates} = useTemplates()

    const [selectedEdgeId, setSelectedEdgeId] = useState<string | null>(null)
    const [connectingFrom, setConnectingFrom] = useState<ConnectionOrigin | null>(null)
    const gestures = useNodeGestures(session.tasks, onMoveTask)

    const locked = session.finalized

    const unlinkable = useMemo(() => {
        if (!connectingFrom) return null

        return connectingFrom.handleType === 'source'
            ? unlinkableFrom(session.tasks, connectingFrom.nodeId)
            : unlinkableInto(session.tasks, connectingFrom.nodeId)
    }, [session.tasks, connectingFrom])

    const levelByTask = useMemo(
        () => taskLevels(session.tasks, templates),
        [session.tasks, templates],
    )

    const nodes = useMemo(
        () =>
            toFlowNodes({
                session,
                selectedTaskId,
                needsInputIds,
                unlinkable,
                levelByTask,
                dragPositions: gestures.dragPositions,
                sizes: gestures.sizes,
            }),
        [
            session,
            selectedTaskId,
            needsInputIds,
            unlinkable,
            levelByTask,
            gestures.dragPositions,
            gestures.sizes,
        ],
    )

    const edges = useMemo(() => toFlowEdges(session, selectedEdgeId), [session, selectedEdgeId])

    const selectNode = (_: unknown, node: TaskNodeType) => {
        setSelectedEdgeId(null)
        onSelectTask(node.id)
    }

    const selectEdge = (_: unknown, edge: GraphEdge) => setSelectedEdgeId(edge.id)

    const clearSelection = () => {
        setSelectedEdgeId(null)
        onSelectTask(null)
    }

    const startConnecting = (_: unknown, {nodeId, handleType}: OnConnectStartParams) =>
        setConnectingFrom(nodeId && handleType ? {nodeId, handleType} : null)

    const stopConnecting = () => setConnectingFrom(null)

    const connect = (connection: Connection) => {
        if (connection.source && connection.target) onConnect(connection.source, connection.target)
    }

    const canConnect = (connection: Connection | Edge) =>
        Boolean(connection.source) &&
        Boolean(connection.target) &&
        !createsCycle(session.tasks, connection.source, connection.target)

    const disconnectAll = (deleted: Edge[]) => {
        for (const edge of deleted) onDisconnect(edge.source, edge.target)
    }

    const addAtOrigin = () => onNewNode(ORIGIN)

    const addAtCursor = (event: MouseEvent<HTMLDivElement>) => {
        if (locked || (event.target as HTMLElement).closest(CANVAS_CHROME)) return

        const at = screenToFlowPosition({x: event.clientX, y: event.clientY})
        onNewNode({x: at.x - NODE_CURSOR_OFFSET.x, y: at.y - NODE_CURSOR_OFFSET.y})
    }

    return (
        <div className="relative min-h-0 flex-1 bg-background" onDoubleClick={addAtCursor}>
            <ReactFlow<TaskNodeType, GraphEdge>
                nodes={nodes}
                edges={edges}
                nodeTypes={nodeTypes}
                fitView
                fitViewOptions={FIT_VIEW_OPTIONS}
                minZoom={MIN_ZOOM}
                maxZoom={MAX_ZOOM}
                zoomOnDoubleClick={false}
                nodesConnectable={!locked}
                deleteKeyCode={['Backspace', 'Delete']}
                proOptions={{hideAttribution: true}}
                onNodesChange={gestures.trackChanges}
                onNodeDragStart={gestures.lift}
                onNodeDragStop={gestures.drop}
                onNodeClick={selectNode}
                onEdgeClick={selectEdge}
                onPaneClick={clearSelection}
                onConnectStart={startConnecting}
                onConnectEnd={stopConnecting}
                onConnect={connect}
                isValidConnection={canConnect}
                onEdgesDelete={disconnectAll}
            >
                <Background
                    variant={BackgroundVariant.Dots}
                    gap={16}
                    size={1}
                    color="var(--border)"
                />
                <ZoomCluster />
            </ReactFlow>

            {session.tasks.length === 0 && <EmptyCanvas locked={locked} onNewNode={addAtOrigin} />}
        </div>
    )
}
