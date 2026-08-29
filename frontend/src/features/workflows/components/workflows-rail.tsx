import type {ReactNode} from 'react'
import {FolderInput, Plus} from 'lucide-react'

import {HelpTip} from '@/shared/components/help-tip'
import {WorkflowRow} from '@/features/workflows/components/workflow-row'
import {cn} from '@/shared/lib/utils'
import {t} from '@/shared/lib/i18n'
import {useRailReorder} from '@/features/workflows/use-rail-reorder'
import {Button} from '@/shared/ui/button'
import {ScrollArea} from '@/shared/ui/scroll-area'
import {Tooltip, TooltipContent, TooltipTrigger} from '@/shared/ui/tooltip'
import type {Workflow} from '@/features/workflows/types'

export function WorkflowsRail({
    nav,
    workflows,
    activeWorkflowId,
    onSelect,
    onCreate,
    onImport,
    onDuplicate,
    onExport,
    onDelete,
    onReorder,
}: {
    nav: ReactNode
    workflows: Workflow[]
    activeWorkflowId: string | null
    onSelect: (workflowId: string) => void
    onCreate: () => void
    onImport: () => void
    onDuplicate: (workflowId: string) => void
    onExport: (workflowId: string) => void
    onDelete: (workflowId: string) => void
    onReorder: (workflowId: string, toIndex: number) => void
}) {
    const reorder = useRailReorder(workflows.length, onReorder)

    return (
        <aside className="surface-card flex h-full w-[280px] shrink-0 flex-col overflow-hidden ring-1 ring-border-strong">
            {nav}

            <div className="flex h-14 shrink-0 items-center justify-between border-b border-border pr-2 pl-4">
                <span className="flex items-center gap-2">
                    <span className="micro-label">{t('workflow.rail.title')}</span>
                    <HelpTip term="workflow" side="bottom" />
                </span>
                <div className="flex items-center gap-1">
                    <Tooltip>
                        <TooltipTrigger asChild>
                            <Button
                                variant="ghost"
                                size="icon-sm"
                                aria-label={t('workflow.rail.import')}
                                onClick={onImport}
                                className="text-muted-foreground hover:bg-muted hover:text-foreground"
                            >
                                <FolderInput />
                            </Button>
                        </TooltipTrigger>
                        <TooltipContent side="bottom">{t('workflow.rail.import')}</TooltipContent>
                    </Tooltip>

                    <Tooltip>
                        <TooltipTrigger asChild>
                            <Button
                                variant="ghost"
                                size="icon-sm"
                                aria-label={t('workflow.rail.new')}
                                onClick={onCreate}
                                className="text-muted-foreground hover:bg-muted hover:text-foreground"
                            >
                                <Plus />
                            </Button>
                        </TooltipTrigger>
                        <TooltipContent side="bottom">{t('workflow.rail.new')}</TooltipContent>
                    </Tooltip>
                </div>
            </div>

            {workflows.length === 0 ? (
                <div className="flex flex-1 flex-col items-center justify-center gap-2 px-6 text-center">
                    <p className="text-base text-muted-foreground">{t('workflow.rail.empty')}</p>
                    <button
                        type="button"
                        onClick={onCreate}
                        className="rounded-md text-base font-medium underline underline-offset-4 outline-none hover:text-foreground/70 focus-visible:ring-2 focus-visible:ring-ring/50"
                    >
                        {t('workflow.rail.createFirst')}
                    </button>
                </div>
            ) : (
                <ScrollArea className="min-h-0 flex-1">
                    <div className="flex min-w-0 flex-col gap-1 p-2">
                        {workflows.map((workflow, index) => (
                            <div
                                key={workflow.id}
                                className={cn(
                                    'relative min-w-0',
                                    reorder.dragging(workflow.id) &&
                                        'z-10 rounded-xl bg-card opacity-95 shadow-[0_8px_24px_rgba(27,28,30,0.16)] ring-1 ring-border-strong select-none [&_*]:cursor-grabbing',
                                )}
                                style={reorder.liftStyle(workflow.id)}
                                {...reorder.rowProps(workflow.id, index)}
                            >
                                {reorder.dropSlot === index && <DropLine />}

                                <WorkflowRow
                                    workflow={workflow}
                                    active={workflow.id === activeWorkflowId}
                                    onSelect={onSelect}
                                    onDuplicate={onDuplicate}
                                    onExport={onExport}
                                    onDelete={onDelete}
                                />

                                {reorder.dropSlot === workflows.length &&
                                    index === workflows.length - 1 && <DropLine last />}
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
