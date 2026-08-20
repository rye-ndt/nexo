import {FolderOpen, Menu, Settings} from 'lucide-react'
import type {ComponentType} from 'react'

import {Button} from '@/shared/ui/button'
import {
    Sheet,
    SheetContent,
    SheetDescription,
    SheetHeader,
    SheetTitle,
    SheetTrigger,
} from '@/shared/ui/sheet'
import {workflowProgress} from '@/features/workflows/graph'
import {useToggle} from '@/shared/hooks/use-toggle'
import type {WorkflowAction, WorkflowActionHandlers} from '@/features/workflows/workflow-actions'
import {cn} from '@/shared/lib/utils'
import {t} from '@/shared/lib/i18n'
import type {Workflow} from '@/features/workflows/types'

export function WorkflowMenu({
    workflow,
    actions,
    handlers,
    onEditLocations,
    onOpenSettings,
}: {
    workflow: Workflow | null
    actions: WorkflowAction[]
    handlers: WorkflowActionHandlers
    onEditLocations: () => void
    onOpenSettings: () => void
}) {
    const sheet = useToggle()

    /** Every pick closes the sheet first, so the dialog it opens is not behind it. */
    const pick = (action: () => void) => () => {
        sheet.close()
        action()
    }

    return (
        <Sheet open={sheet.on} onOpenChange={sheet.set}>
            <SheetTrigger asChild>
                <Button
                    variant="ghost"
                    size="icon"
                    aria-label={t('canvas.sheet.open')}
                    className="lg:hidden"
                >
                    <Menu />
                </Button>
            </SheetTrigger>

            <SheetContent side="left" className="gap-0">
                <SheetHeader className="h-14 shrink-0 justify-center border-b border-border px-4 py-0">
                    <SheetTitle className="micro-label">{t('canvas.sheet.title')}</SheetTitle>
                    <SheetDescription className="sr-only">
                        {t('canvas.sheet.description')}
                    </SheetDescription>
                </SheetHeader>

                <div className="flex min-h-0 flex-1 flex-col gap-1 overflow-x-hidden overflow-y-auto p-2">
                    {workflow && (
                        <>
                            <WorkflowLocations workflow={workflow} onEdit={pick(onEditLocations)} />
                            <WorkflowProgress workflow={workflow} />

                            <Divider />

                            {actions.map((action) => (
                                <MenuAction
                                    key={action.id}
                                    label={action.label}
                                    icon={action.icon}
                                    disabledReason={action.disabledReason}
                                    onClick={pick(handlers[action.id])}
                                />
                            ))}

                            <Divider />
                        </>
                    )}

                    <MenuAction
                        label={t('canvas.sheet.settings')}
                        icon={Settings}
                        onClick={pick(onOpenSettings)}
                    />
                </div>
            </SheetContent>
        </Sheet>
    )
}

function Divider() {
    return <span className="my-1 h-px shrink-0 bg-border" />
}

function MenuAction({
    label,
    icon: Icon,
    disabledReason,
    onClick,
}: {
    label: string
    icon: ComponentType<{className?: string}>
    disabledReason?: string
    onClick: () => void
}) {
    return (
        <span className="flex flex-col">
            <Button
                variant="ghost"
                className={cn('w-full justify-start', disabledReason && 'opacity-50')}
                aria-disabled={disabledReason ? true : undefined}
                onClick={disabledReason ? undefined : onClick}
            >
                <Icon />
                {label}
            </Button>

            {disabledReason && (
                <span className="px-3 pb-1 text-sm text-muted-foreground">{disabledReason}</span>
            )}
        </span>
    )
}

function WorkflowLocations({workflow, onEdit}: {workflow: Workflow; onEdit: () => void}) {
    if (!workflow.projectDir) {
        if (workflow.locked)
            return (
                <span className="flex flex-col gap-2 px-3 py-2">
                    <span className="micro-label">{t('canvas.sheet.projectDir')}</span>
                    <span className="text-sm text-muted-foreground">
                        {t('canvas.sheet.noProjectDir')}
                    </span>
                </span>
            )

        return (
            <button
                type="button"
                onClick={onEdit}
                className="flex items-center gap-2 rounded-lg border border-live/40 bg-live-tint px-3 py-2 text-left text-sm font-medium text-live outline-none transition-colors hover:border-live/70 focus-visible:ring-2 focus-visible:ring-ring/50"
            >
                <FolderOpen className="size-4 shrink-0" />
                {t('canvas.sheet.chooseProjectDir')}
            </button>
        )
    }

    const body = (
        <>
            <span className="micro-label">{t('canvas.sheet.projectDir')}</span>
            <span className="truncate font-mono text-sm text-muted-foreground">
                {workflow.projectDir}
            </span>
        </>
    )

    if (workflow.locked) return <span className="flex flex-col gap-2 px-3 py-2">{body}</span>

    return (
        <button
            type="button"
            onClick={onEdit}
            className="flex flex-col gap-2 rounded-lg px-3 py-2 text-left outline-none transition-colors hover:bg-muted focus-visible:ring-2 focus-visible:ring-ring/50"
        >
            {body}
        </button>
    )
}

function WorkflowProgress({workflow}: {workflow: Workflow}) {
    const {done, total} = workflowProgress(workflow)

    return (
        <span className="flex items-center justify-between gap-3 px-3 py-2">
            <span className="micro-label">{t('canvas.sheet.progress')}</span>
            <span className="font-mono text-sm text-muted-foreground">
                {done}/{total}
            </span>
        </span>
    )
}
