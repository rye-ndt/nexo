import {Plus, Upload} from 'lucide-react'

import {SessionRow} from '@/features/sessions/components/session-row'
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
}: {
    sessions: Session[]
    activeSessionId: string | null
    onSelect: (sessionId: string) => void
    onCreate: () => void
    onImport: () => void
    onClone: (sessionId: string) => void
    onExport: (sessionId: string) => void
    onDelete: (sessionId: string) => void
}) {
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
                    <div className="flex flex-col gap-1 p-2">
                        {sessions.map((session) => (
                            <SessionRow
                                key={session.id}
                                session={session}
                                active={session.id === activeSessionId}
                                onSelect={onSelect}
                                onClone={onClone}
                                onExport={onExport}
                                onDelete={onDelete}
                            />
                        ))}
                    </div>
                </ScrollArea>
            )}
        </aside>
    )
}
