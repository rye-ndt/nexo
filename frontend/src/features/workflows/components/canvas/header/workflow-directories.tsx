import {Folder, FolderOpen} from 'lucide-react'

import {Tooltip, TooltipContent, TooltipTrigger} from '@/shared/ui/tooltip'
import {t} from '@/shared/lib/i18n'
import type {Workflow} from '@/features/workflows/types'

const WIDTH = 'hidden min-w-0 shrink lg:max-w-40 xl:max-w-72'

export function WorkflowDirectories({workflow, onEdit}: {workflow: Workflow; onEdit: () => void}) {
    if (!workflow.projectDir) {
        if (workflow.locked)
            return (
                <span className={`${WIDTH} truncate text-sm text-muted-foreground lg:block`}>
                    {t('canvas.dirs.none')}
                </span>
            )

        return (
            <Tooltip>
                <TooltipTrigger asChild>
                    <button
                        type="button"
                        onClick={onEdit}
                        className="hidden shrink-0 items-center gap-2 rounded-md border border-live/40 bg-live-tint px-2 py-1 text-sm font-medium text-live outline-none transition-colors hover:border-live/70 focus-visible:ring-2 focus-visible:ring-live lg:flex"
                    >
                        <FolderOpen className="size-3.5 shrink-0" />
                        {t('canvas.dirs.choose')}
                    </button>
                </TooltipTrigger>
                <TooltipContent side="bottom">{t('canvas.dirs.chooseHint')}</TooltipContent>
            </Tooltip>
        )
    }

    if (workflow.locked)
        return (
            <span className={`${WIDTH} truncate font-mono text-sm text-muted-foreground lg:block`}>
                {workflow.projectDir}
            </span>
        )

    return (
        <Tooltip>
            <TooltipTrigger asChild>
                <button
                    type="button"
                    aria-label={t('canvas.dirs.label')}
                    onClick={onEdit}
                    className={`${WIDTH} items-center gap-2 rounded-md px-2 py-1 outline-none transition-colors hover:bg-muted focus-visible:ring-2 focus-visible:ring-live lg:flex`}
                >
                    <Folder className="size-3.5 shrink-0 text-muted-foreground" />
                    <span className="hidden truncate font-mono text-sm text-muted-foreground lg:block">
                        {workflow.projectDir}
                    </span>
                </button>
            </TooltipTrigger>
            <TooltipContent side="bottom">
                <span className="flex flex-col gap-1">
                    <span className="font-mono">{workflow.projectDir}</span>
                    <span>{t('canvas.dirs.change')}</span>
                </span>
            </TooltipContent>
        </Tooltip>
    )
}
