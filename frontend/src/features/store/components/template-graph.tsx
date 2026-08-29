import {tn} from '@/shared/lib/i18n'
import type {StoreStep} from '@/features/store/types'

const CELL_X = 40
const CELL_Y = 30
const PAD = 6
const PIP = 4

type Placed = {id: string; x: number; y: number}

/** Column and row from the step's own canvas position, so the drawing is the graph. */
function place(steps: StoreStep[]): Placed[] {
    const columns = [...new Set(steps.map((step) => step.position.x))].sort((a, b) => a - b)
    const rows = [...new Set(steps.map((step) => step.position.y))].sort((a, b) => a - b)

    return steps.map((step) => ({
        id: step.id,
        x: PAD + columns.indexOf(step.position.x) * CELL_X,
        y: PAD + rows.indexOf(step.position.y) * CELL_Y,
    }))
}

/**
 * The shape of the graph a template will give you: one pip per step, one line per
 * dependency, drawn from the steps' real positions. It says at a glance whether
 * this is a straight chain or something that fans out, which no icon would.
 *
 * Drawn at its natural size rather than stretched to the plate, so a pip means the
 * same thing on every card and a longer chain reads as a wider drawing.
 */
export function TemplateGraph({steps}: {steps: StoreStep[]}) {
    const placed = place(steps)
    const at = new Map(placed.map((node) => [node.id, node]))

    const width = Math.max(...placed.map((node) => node.x)) + PAD
    const height = Math.max(...placed.map((node) => node.y)) + PAD

    return (
        <svg
            role="img"
            aria-label={tn('store.card.steps.one', 'store.card.steps.other', steps.length)}
            width={width}
            height={height}
            viewBox={`0 0 ${width} ${height}`}
            className="max-h-full max-w-full"
        >
            {steps.map((step) =>
                step.dependsOn.map((parentId) => {
                    const from = at.get(parentId)
                    const to = at.get(step.id)
                    if (!from || !to) return null

                    return (
                        <line
                            key={`${parentId}-${step.id}`}
                            x1={from.x}
                            y1={from.y}
                            x2={to.x}
                            y2={to.y}
                            strokeWidth={1.5}
                            className="stroke-muted-foreground/35 transition-colors duration-[120ms] group-hover:stroke-live/50"
                        />
                    )
                }),
            )}

            {placed.map((node) => (
                <circle
                    key={node.id}
                    cx={node.x}
                    cy={node.y}
                    r={PIP}
                    className="fill-muted-foreground transition-colors duration-[120ms] group-hover:fill-live"
                />
            ))}
        </svg>
    )
}
