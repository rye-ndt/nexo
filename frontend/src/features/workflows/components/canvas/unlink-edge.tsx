import {useState} from 'react'
import {X} from 'lucide-react'
import {
    BaseEdge,
    EdgeLabelRenderer,
    getSmoothStepPath,
    type Edge,
    type EdgeProps,
} from '@xyflow/react'

import {cn} from '@/shared/lib/utils'

const BORDER_RADIUS = 12

export type UnlinkEdgeData = {
    /** The pointer is on this edge, so the step it links may be unlinked from here. */
    revealed: boolean
    label: string
    onHold: () => void
    onRelease: () => void
    onUnlink: () => void
}

export type UnlinkEdgeType = Edge<UnlinkEdgeData, 'unlink'>

/**
 * Hovering the line offers a cut at its middle. Arming that cut turns the whole
 * connection red, so a dense graph never leaves you guessing which of two crossing
 * lines you are about to remove.
 */
export function UnlinkEdge({
    id,
    sourceX,
    sourceY,
    sourcePosition,
    targetX,
    targetY,
    targetPosition,
    data,
}: EdgeProps<UnlinkEdgeType>) {
    const [onControl, setOnControl] = useState(false)

    const [path, labelX, labelY] = getSmoothStepPath({
        sourceX,
        sourceY,
        sourcePosition,
        targetX,
        targetY,
        targetPosition,
        borderRadius: BORDER_RADIUS,
    })

    const revealed = data?.revealed ?? false
    const armed = revealed && onControl

    return (
        <>
            <BaseEdge id={id} path={path} className={cn(armed && 'is-armed')} />

            {revealed && data && (
                <EdgeLabelRenderer>
                    <button
                        type="button"
                        aria-label={data.label}
                        title={data.label}
                        style={{
                            transform: `translate(-50%, -50%) translate(${labelX}px, ${labelY}px)`,
                        }}
                        onMouseEnter={() => {
                            data.onHold()
                            setOnControl(true)
                        }}
                        onMouseLeave={() => {
                            data.onRelease()
                            setOnControl(false)
                        }}
                        onClick={data.onUnlink}
                        className={cn(
                            'nodrag nopan pointer-events-auto absolute flex size-5 items-center justify-center rounded-full border-2 transition-colors',
                            'animate-in duration-150 fade-in zoom-in-75 motion-reduce:animate-none',
                            'focus-visible:ring-2 focus-visible:ring-live focus-visible:outline-none',
                            armed
                                ? 'border-state-failed bg-state-failed text-background'
                                : 'border-border-strong bg-card text-muted-foreground',
                        )}
                    >
                        <X aria-hidden="true" className="size-2.5" strokeWidth={3} />
                    </button>
                </EdgeLabelRenderer>
            )}
        </>
    )
}
