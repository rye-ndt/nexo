import type {ReactNode} from 'react'

import {
    ContextMenuContent,
    ContextMenuItem,
    ContextMenuLabel,
    ContextMenuSeparator,
} from '@/shared/ui/context-menu'
import {findStep} from '@/features/workflows/graph'
import {t} from '@/shared/lib/i18n'
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
            <Menu subject={t('canvas.menu.pane')}>
                {!locked && (
                    <>
                        <ContextMenuItem onSelect={() => onNewStep(target.at)}>
                            {t('canvas.menu.newStepHere')}
                        </ContextMenuItem>
                        <ContextMenuSeparator />
                    </>
                )}
                <ContextMenuItem onSelect={onFitView}>{t('canvas.menu.fitView')}</ContextMenuItem>
            </Menu>
        )

    if (target.kind === 'step') {
        const step = findStep(workflow, target.stepId)
        if (!step) return null

        return (
            <Menu subject={step.title || t('step.untitled')}>
                <ContextMenuItem onSelect={() => onOpenStep(step.id)}>
                    {t('canvas.menu.openStep')}
                </ContextMenuItem>
                {!locked && (
                    <>
                        <ContextMenuSeparator />
                        <ContextMenuItem onSelect={() => onDuplicateStep(step.id)}>
                            {t('canvas.menu.duplicateStep')}
                        </ContextMenuItem>
                        <ContextMenuItem
                            variant="destructive"
                            onSelect={() => onDeleteStep(step.id)}
                        >
                            {t('canvas.menu.deleteStep')}
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
        <Menu
            subject={t('canvas.menu.edgeSubject', {
                source: source.title || t('step.untitled'),
                target: downstream.title || t('step.untitled'),
            })}
        >
            <ContextMenuItem
                variant="destructive"
                onSelect={() => onUnlink(source.id, downstream.id)}
            >
                {t('canvas.menu.unlink')}
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
