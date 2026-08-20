import type {ReactNode} from 'react'

import {
    ContextMenuContent,
    ContextMenuItem,
    ContextMenuLabel,
    ContextMenuSeparator,
} from '@/shared/ui/context-menu'
import {UNTITLED} from '@/features/workflows/components/canvas/step-card'
import {findStep} from '@/features/workflows/graph'
import type {Point, Workflow} from '@/features/workflows/types'

export type CanvasTarget =
    | {kind: 'step'; stepId: string}
    | {kind: 'edge'; sourceId: string; targetId: string}
    | {kind: 'pane'; at: Point}

export function CanvasMenu({
    workflow,
    target,
    onOpenStep,
    onDuplicateStep,
    onDeleteStep,
    onUnlink,
    onNewStep,
    onFitView,
}: {
    workflow: Workflow
    target: CanvasTarget
    onOpenStep: (stepId: string) => void
    onDuplicateStep: (stepId: string) => void
    onDeleteStep: (stepId: string) => void
    onUnlink: (sourceId: string, targetId: string) => void
    onNewStep: (at: Point) => void
    onFitView: () => void
}) {
    const locked = workflow.locked

    if (target.kind === 'pane')
        return (
            <Menu subject="Canvas">
                {!locked && (
                    <>
                        <ContextMenuItem onSelect={() => onNewStep(target.at)}>
                            New step here
                        </ContextMenuItem>
                        <ContextMenuSeparator />
                    </>
                )}
                <ContextMenuItem onSelect={onFitView}>Fit view</ContextMenuItem>
            </Menu>
        )

    if (target.kind === 'step') {
        const step = findStep(workflow, target.stepId)
        if (!step) return null

        return (
            <Menu subject={step.title || UNTITLED}>
                <ContextMenuItem onSelect={() => onOpenStep(step.id)}>Open step</ContextMenuItem>
                {!locked && (
                    <>
                        <ContextMenuSeparator />
                        <ContextMenuItem onSelect={() => onDuplicateStep(step.id)}>
                            Duplicate step
                        </ContextMenuItem>
                        <ContextMenuItem
                            variant="destructive"
                            onSelect={() => onDeleteStep(step.id)}
                        >
                            Delete step
                        </ContextMenuItem>
                    </>
                )}
            </Menu>
        )
    }

    const source = findStep(workflow, target.sourceId)
    const downstream = findStep(workflow, target.targetId)
    if (!source || !downstream) return null

    return (
        <Menu subject={`${source.title || UNTITLED} → ${downstream.title || UNTITLED}`}>
            <ContextMenuItem
                variant="destructive"
                onSelect={() => onUnlink(source.id, downstream.id)}
            >
                Unlink
            </ContextMenuItem>
        </Menu>
    )
}

function Menu({subject, children}: {subject: string; children: ReactNode}) {
    return (
        <ContextMenuContent className="max-w-64">
            <ContextMenuLabel>{subject}</ContextMenuLabel>
            {children}
        </ContextMenuContent>
    )
}
