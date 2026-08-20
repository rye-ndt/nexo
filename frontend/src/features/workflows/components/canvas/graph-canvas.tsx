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

import {CanvasMenu, type CanvasTarget} from '@/features/workflows/components/canvas/canvas-menu'
import {ContextMenu, ContextMenuTrigger} from '@/shared/ui/context-menu'
import {EmptyCanvas} from '@/features/workflows/components/canvas/empty-canvas'
import {ZoomCluster} from '@/features/workflows/components/canvas/zoom-cluster'
import {StepCard} from '@/features/workflows/components/canvas/step-card'
import {
    efforts,
    toFlowEdges,
    toFlowNodes,
    type GraphEdge,
} from '@/features/workflows/components/canvas/flow-graph'
import {UnlinkEdge} from '@/features/workflows/components/canvas/unlink-edge'
import {useEdgeHover} from '@/features/workflows/components/canvas/use-edge-hover'
import {useStepGestures} from '@/features/workflows/components/canvas/use-step-gestures'
import {
    CANVAS_CHROME,
    FIT_MS,
    FIT_VIEW_OPTIONS,
    MAX_ZOOM,
    MIN_ZOOM,
    STEP_CURSOR_OFFSET,
    ORIGIN,
} from '@/features/workflows/components/canvas/view'
import {useRoles} from '@/features/roles/use-roles'
import {createsCycle, unlinkableFrom, unlinkableInto} from '@/features/workflows/graph'
import type {StepNodeType} from '@/features/workflows/components/canvas/step-card'
import type {Point, Workflow} from '@/features/workflows/types'

type GraphCanvasProps = {
    workflow: Workflow
    selectedStepId: string | null
    needsInputIds: Set<string>
    onSelectStep: (stepId: string | null) => void
    onMoveStep: (stepId: string, position: Point) => void
    onConnect: (sourceId: string, targetId: string) => void
    onDisconnect: (sourceId: string, targetId: string) => void
    onNewStep: (position: Point) => void
    onDeleteStep: (stepId: string) => void
}

type ConnectionOrigin = {stepId: string; handleType: HandleType}

type Cursor = {clientX: number; clientY: number}

const nodeTypes = {step: StepCard}

const edgeTypes = {unlink: UnlinkEdge}

export function GraphCanvas(props: GraphCanvasProps) {
    return (
        <ReactFlowProvider>
            <Canvas {...props} />
        </ReactFlowProvider>
    )
}

function Canvas({
    workflow,
    selectedStepId,
    needsInputIds,
    onSelectStep,
    onMoveStep,
    onConnect,
    onDisconnect,
    onNewStep,
    onDeleteStep,
}: GraphCanvasProps) {
    const {screenToFlowPosition, fitView} = useReactFlow()
    const {roles} = useRoles()

    const [selectedEdgeId, setSelectedEdgeId] = useState<string | null>(null)
    const [connectingFrom, setConnectingFrom] = useState<ConnectionOrigin | null>(null)
    const [menuTarget, setMenuTarget] = useState<CanvasTarget | null>(null)
    const armedTarget = useRef<CanvasTarget | null>(null)
    const gestures = useStepGestures(workflow.steps, onMoveStep)
    const {hoveredEdgeId, hold, release} = useEdgeHover()

    const locked = workflow.locked

    const unlinkable = useMemo(() => {
        if (!connectingFrom) return null

        return connectingFrom.handleType === 'source'
            ? unlinkableFrom(workflow.steps, connectingFrom.stepId)
            : unlinkableInto(workflow.steps, connectingFrom.stepId)
    }, [workflow.steps, connectingFrom])

    const effortByStep = useMemo(() => efforts(workflow.steps, roles), [workflow.steps, roles])

    const nodes = useMemo(
        () =>
            toFlowNodes({
                workflow,
                selectedStepId,
                needsInputIds,
                unlinkable,
                effortByStep,
                dragPositions: gestures.dragPositions,
                sizes: gestures.sizes,
            }),
        [
            workflow,
            selectedStepId,
            needsInputIds,
            unlinkable,
            effortByStep,
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
                workflow,
                selectedEdgeId,
                hoveredEdgeId,
                onHold: hold,
                onRelease: release,
                onUnlink: unlink,
            }),
        [workflow, selectedEdgeId, hoveredEdgeId, hold, release, unlink],
    )

    const selectStep = (_: unknown, step: StepNodeType) => {
        setSelectedEdgeId(null)
        onSelectStep(step.id)
    }

    const selectEdge = (_: unknown, edge: GraphEdge) => setSelectedEdgeId(edge.id)

    const clearSelection = () => {
        setSelectedEdgeId(null)
        onSelectStep(null)
    }

    const startConnecting = (_: unknown, {nodeId, handleType}: OnConnectStartParams) =>
        setConnectingFrom(nodeId && handleType ? {stepId: nodeId, handleType} : null)

    const stopConnecting = () => setConnectingFrom(null)

    const connect = (connection: Connection) => {
        if (connection.source && connection.target) onConnect(connection.source, connection.target)
    }

    const canConnect = (connection: Connection | Edge) =>
        Boolean(connection.source) &&
        Boolean(connection.target) &&
        !createsCycle(workflow.steps, connection.source, connection.target)

    const disconnectAll = (deleted: Edge[]) => {
        for (const edge of deleted) onDisconnect(edge.source, edge.target)
    }

    const addAtOrigin = () => onNewStep(ORIGIN)

    const stepPositionAt = (event: Cursor) => {
        const at = screenToFlowPosition({x: event.clientX, y: event.clientY})
        return {x: at.x - STEP_CURSOR_OFFSET.x, y: at.y - STEP_CURSOR_OFFSET.y}
    }

    const addAtCursor = (event: MouseEvent<HTMLDivElement>) => {
        if (locked || (event.target as HTMLElement).closest(CANVAS_CHROME)) return
        onNewStep(stepPositionAt(event))
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
                    <ReactFlow<StepNodeType, GraphEdge>
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
                        onNodeClick={selectStep}
                        onEdgeClick={selectEdge}
                        onEdgeMouseEnter={(_, edge) => hold(edge.id)}
                        onEdgeMouseLeave={release}
                        onPaneClick={clearSelection}
                        onNodeContextMenu={(_, step) => arm({kind: 'step', stepId: step.id})}
                        onEdgeContextMenu={(_, edge) =>
                            arm(
                                locked
                                    ? null
                                    : {kind: 'edge', sourceId: edge.source, targetId: edge.target},
                            )
                        }
                        onPaneContextMenu={(event: Cursor) =>
                            arm({kind: 'pane', at: stepPositionAt(event)})
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

                    {workflow.steps.length === 0 && (
                        <EmptyCanvas locked={locked} onNewStep={addAtOrigin} />
                    )}
                </div>
            </ContextMenuTrigger>

            {menuTarget && (
                <CanvasMenu
                    workflow={workflow}
                    target={menuTarget}
                    onOpenStep={onSelectStep}
                    onDeleteStep={onDeleteStep}
                    onUnlink={unlink}
                    onNewStep={onNewStep}
                    onFitView={fitToView}
                />
            )}
        </ContextMenu>
    )
}
