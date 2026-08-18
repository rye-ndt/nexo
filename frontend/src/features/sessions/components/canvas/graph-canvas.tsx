import {useCallback, useMemo, useRef, useState, type MouseEvent} from 'react'
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

import {CanvasMenu, type CanvasTarget} from '@/features/sessions/components/canvas/canvas-menu'
import {ContextMenu, ContextMenuTrigger} from '@/shared/ui/context-menu'
import {EmptyCanvas} from '@/features/sessions/components/canvas/empty-canvas'
import {ZoomCluster} from '@/features/sessions/components/canvas/zoom-cluster'
import {TaskNode} from '@/features/sessions/components/canvas/task-node'
import {
    taskLevels,
    toFlowEdges,
    toFlowNodes,
    type GraphEdge,
} from '@/features/sessions/components/canvas/flow-graph'
import {UnlinkEdge} from '@/features/sessions/components/canvas/unlink-edge'
import {useEdgeHover} from '@/features/sessions/components/canvas/use-edge-hover'
import {useNodeGestures} from '@/features/sessions/components/canvas/use-node-gestures'
import {
    CANVAS_CHROME,
    FIT_MS,
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
    onDeleteTask: (taskId: string) => void
}

type ConnectionOrigin = {nodeId: string; handleType: HandleType}

type Cursor = {clientX: number; clientY: number}

const nodeTypes = {task: TaskNode}

const edgeTypes = {unlink: UnlinkEdge}

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
    onDeleteTask,
}: GraphCanvasProps) {
    const {screenToFlowPosition, fitView} = useReactFlow()
    const {templates} = useTemplates()

    const [selectedEdgeId, setSelectedEdgeId] = useState<string | null>(null)
    const [connectingFrom, setConnectingFrom] = useState<ConnectionOrigin | null>(null)
    const [menuTarget, setMenuTarget] = useState<CanvasTarget | null>(null)
    const armedTarget = useRef<CanvasTarget | null>(null)
    const gestures = useNodeGestures(session.tasks, onMoveTask)
    const {hoveredEdgeId, hold, release} = useEdgeHover()

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

    const unlink = useCallback(
        (sourceId: string, targetId: string) => {
            release()
            onDisconnect(sourceId, targetId)
        },
        [release, onDisconnect],
    )

    const edges = useMemo(
        () =>
            toFlowEdges({
                session,
                selectedEdgeId,
                hoveredEdgeId,
                onHold: hold,
                onRelease: release,
                onUnlink: unlink,
            }),
        [session, selectedEdgeId, hoveredEdgeId, hold, release, unlink],
    )

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

    const nodePositionAt = (event: Cursor) => {
        const at = screenToFlowPosition({x: event.clientX, y: event.clientY})
        return {x: at.x - NODE_CURSOR_OFFSET.x, y: at.y - NODE_CURSOR_OFFSET.y}
    }

    const addAtCursor = (event: MouseEvent<HTMLDivElement>) => {
        if (locked || (event.target as HTMLElement).closest(CANVAS_CHROME)) return
        onNewNode(nodePositionAt(event))
    }

    const arm = (target: CanvasTarget | null) => {
        armedTarget.current = target
    }

    const openMenu = (event: MouseEvent<HTMLDivElement>) => {
        const target = armedTarget.current
        arm(null)

        if (!target) return event.preventDefault()

        setMenuTarget(target)
    }

    const fitToView = () => fitView({...FIT_VIEW_OPTIONS, duration: FIT_MS})

    return (
        <ContextMenu>
            <ContextMenuTrigger asChild>
                <div
                    className="relative min-h-0 flex-1 bg-background"
                    onDoubleClick={addAtCursor}
                    onContextMenu={openMenu}
                >
                    <ReactFlow<TaskNodeType, GraphEdge>
                        nodes={nodes}
                        edges={edges}
                        nodeTypes={nodeTypes}
                        edgeTypes={edgeTypes}
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
                        onEdgeMouseEnter={(_, edge) => hold(edge.id)}
                        onEdgeMouseLeave={release}
                        onPaneClick={clearSelection}
                        onNodeContextMenu={(_, node) => arm({kind: 'node', taskId: node.id})}
                        onEdgeContextMenu={(_, edge) =>
                            arm(
                                locked
                                    ? null
                                    : {kind: 'edge', sourceId: edge.source, targetId: edge.target},
                            )
                        }
                        onPaneContextMenu={(event: Cursor) =>
                            arm({kind: 'pane', at: nodePositionAt(event)})
                        }
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

                    {session.tasks.length === 0 && (
                        <EmptyCanvas locked={locked} onNewNode={addAtOrigin} />
                    )}
                </div>
            </ContextMenuTrigger>

            {menuTarget && (
                <CanvasMenu
                    session={session}
                    target={menuTarget}
                    onOpenTask={onSelectTask}
                    onDeleteTask={onDeleteTask}
                    onUnlink={unlink}
                    onNewNode={onNewNode}
                    onFitView={fitToView}
                />
            )}
        </ContextMenu>
    )
}
