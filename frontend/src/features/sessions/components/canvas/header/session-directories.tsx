import {Folder} from 'lucide-react'

import {Tooltip, TooltipContent, TooltipTrigger} from '@/shared/ui/tooltip'
import {sessionPaths} from '@/features/sessions/graph'
import type {Session} from '@/features/sessions/types'

const WIDTH = 'hidden min-w-0 shrink lg:max-w-40 xl:max-w-72'

export function SessionDirectories({session, onEdit}: {session: Session; onEdit: () => void}) {
    const paths = sessionPaths(session)
    if (paths.length === 0) return null

    if (session.finalized)
        return (
            <span className={`${WIDTH} truncate font-mono text-sm text-muted-foreground lg:block`}>
                {session.workingDir}
            </span>
        )

    return (
        <Tooltip>
            <TooltipTrigger asChild>
                <button
                    type="button"
                    aria-label="Session directories"
                    onClick={onEdit}
                    className={`${WIDTH} items-center gap-2 rounded-md px-2 py-1 outline-none transition-colors hover:bg-muted focus-visible:ring-2 focus-visible:ring-live lg:flex`}
                >
                    <Folder className="size-3.5 shrink-0 text-muted-foreground" />
                    <span className="hidden truncate font-mono text-sm text-muted-foreground lg:block">
                        {session.workingDir}
                    </span>
                </button>
            </TooltipTrigger>
            <TooltipContent side="bottom">
                <span className="flex flex-col gap-1">
                    {paths.map((path) => (
                        <span key={path} className="font-mono">
                            {path}
                        </span>
                    ))}
                    <span>Click to change either one.</span>
                </span>
            </TooltipContent>
        </Tooltip>
    )
}
