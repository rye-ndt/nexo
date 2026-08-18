import type {ReactNode} from 'react'

import {
    ContextMenuContent,
    ContextMenuItem,
    ContextMenuLabel,
    ContextMenuSeparator,
} from '@/shared/ui/context-menu'
import {UNTITLED} from '@/features/sessions/components/canvas/task-node'
import {findTask} from '@/features/sessions/graph'
import type {Point, Session} from '@/features/sessions/types'

export type CanvasTarget =
    | {kind: 'node'; taskId: string}
    | {kind: 'edge'; sourceId: string; targetId: string}
    | {kind: 'pane'; at: Point}

export function CanvasMenu({
    session,
    target,
    onOpenTask,
    onDeleteTask,
    onUnlink,
    onNewNode,
    onFitView,
}: {
    session: Session
    target: CanvasTarget
    onOpenTask: (taskId: string) => void
    onDeleteTask: (taskId: string) => void
    onUnlink: (sourceId: string, targetId: string) => void
    onNewNode: (at: Point) => void
    onFitView: () => void
}) {
    const locked = session.finalized

    if (target.kind === 'pane')
        return (
            <Menu subject="Canvas">
                {!locked && (
                    <>
                        <ContextMenuItem onSelect={() => onNewNode(target.at)}>
                            New node here
                        </ContextMenuItem>
                        <ContextMenuSeparator />
                    </>
                )}
                <ContextMenuItem onSelect={onFitView}>Fit view</ContextMenuItem>
            </Menu>
        )

    if (target.kind === 'node') {
        const task = findTask(session, target.taskId)
        if (!task) return null

        return (
            <Menu subject={task.title || UNTITLED}>
                <ContextMenuItem onSelect={() => onOpenTask(task.id)}>Open node</ContextMenuItem>
                {!locked && (
                    <>
                        <ContextMenuSeparator />
                        <ContextMenuItem
                            variant="destructive"
                            onSelect={() => onDeleteTask(task.id)}
                        >
                            Delete node
                        </ContextMenuItem>
                    </>
                )}
            </Menu>
        )
    }

    const source = findTask(session, target.sourceId)
    const downstream = findTask(session, target.targetId)
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
