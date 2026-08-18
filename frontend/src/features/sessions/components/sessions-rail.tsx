import {Plus, Upload} from 'lucide-react'

import {SessionRow} from '@/features/sessions/components/session-row'
import {cn} from '@/shared/lib/utils'
import {useRailReorder} from '@/features/sessions/use-rail-reorder'
import {Button} from '@/shared/ui/button'
import {ScrollArea} from '@/shared/ui/scroll-area'
import {Tooltip, TooltipContent, TooltipTrigger} from '@/shared/ui/tooltip'
import type {Session} from '@/features/sessions/types'

export function SessionsRail({
    sessions,
    activeSessionId,
    onSelect,
    onCreate,
    onImport,
    onClone,
    onExport,
    onDelete,
    onReorder,
}: {
    sessions: Session[]
    activeSessionId: string | null
    onSelect: (sessionId: string) => void
    onCreate: () => void
    onImport: () => void
    onClone: (sessionId: string) => void
    onExport: (sessionId: string) => void
    onDelete: (sessionId: string) => void
    onReorder: (sessionId: string, toIndex: number) => void
}) {
    const reorder = useRailReorder(sessions.length, onReorder)

    return (
        <aside className="surface-card flex h-full w-[280px] shrink-0 flex-col overflow-hidden ring-1 ring-border-strong">
            <div className="flex h-14 shrink-0 items-center justify-between border-b border-border pr-2 pl-4">
                <span className="micro-label">Sessions</span>
                <div className="flex items-center gap-1">
                    <Tooltip>
                        <TooltipTrigger asChild>
                            <Button
                                variant="ghost"
                                size="icon-sm"
                                aria-label="Import session"
                                onClick={onImport}
                                className="text-muted-foreground hover:bg-muted hover:text-foreground"
                            >
                                <Upload />
                            </Button>
                        </TooltipTrigger>
                        <TooltipContent side="bottom">Import session</TooltipContent>
                    </Tooltip>

                    <Tooltip>
                        <TooltipTrigger asChild>
                            <Button
                                variant="ghost"
                                size="icon-sm"
                                aria-label="New session"
                                onClick={onCreate}
                                className="text-muted-foreground hover:bg-muted hover:text-foreground"
                            >
                                <Plus />
                            </Button>
                        </TooltipTrigger>
                        <TooltipContent side="bottom">New session</TooltipContent>
                    </Tooltip>
                </div>
            </div>

            {sessions.length === 0 ? (
                <div className="flex flex-1 flex-col items-center justify-center gap-2 px-6 text-center">
                    <p className="text-base text-muted-foreground">No sessions yet.</p>
                    <button
                        type="button"
                        onClick={onCreate}
                        className="rounded-md text-base font-medium underline underline-offset-4 outline-none hover:text-foreground/70 focus-visible:ring-2 focus-visible:ring-ring/50"
                    >
                        Create your first session
                    </button>
                </div>
            ) : (
                <ScrollArea className="min-h-0 flex-1">
                    <div className="flex min-w-0 flex-col gap-1 p-2">
                        {sessions.map((session, index) => (
                            <div
                                key={session.id}
                                className={cn(
                                    'relative min-w-0',
                                    reorder.dragging(session.id) &&
                                        'z-10 rounded-xl bg-card opacity-95 shadow-[0_8px_24px_rgba(27,28,30,0.16)] ring-1 ring-border-strong select-none [&_*]:cursor-grabbing',
                                )}
                                style={reorder.liftStyle(session.id)}
                                {...reorder.rowProps(session.id, index)}
                            >
                                {reorder.dropSlot === index && <DropLine />}

                                <SessionRow
                                    session={session}
                                    active={session.id === activeSessionId}
                                    onSelect={onSelect}
                                    onClone={onClone}
                                    onExport={onExport}
                                    onDelete={onDelete}
                                />

                                {reorder.dropSlot === sessions.length &&
                                    index === sessions.length - 1 && <DropLine last />}
                            </div>
                        ))}
                    </div>
                </ScrollArea>
            )}
        </aside>
    )
}

function DropLine({last}: {last?: boolean}) {
    return (
        <span
            aria-hidden
            className={cn(
                'pointer-events-none absolute inset-x-1 h-0.5 rounded-full bg-live',
                last ? '-bottom-[3px]' : '-top-[3px]',
            )}
        />
    )
}
