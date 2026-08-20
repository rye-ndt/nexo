import {Menu, Settings} from 'lucide-react'
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
                    aria-label="Workflow menu"
                    className="lg:hidden"
                >
                    <Menu />
                </Button>
            </SheetTrigger>

            <SheetContent side="left" className="gap-0">
                <SheetHeader className="h-14 shrink-0 justify-center border-b border-border px-4 py-0">
                    <SheetTitle className="micro-label">Menu</SheetTitle>
                    <SheetDescription className="sr-only">
                        Directories, progress and actions for this workflow.
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
                                    onClick={pick(handlers[action.id])}
                                />
                            ))}

                            <Divider />
                        </>
                    )}

                    <MenuAction label="Settings" icon={Settings} onClick={pick(onOpenSettings)} />
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
    onClick,
}: {
    label: string
    icon: ComponentType<{className?: string}>
    onClick: () => void
}) {
    return (
        <Button variant="ghost" className="w-full justify-start" onClick={onClick}>
            <Icon />
            {label}
        </Button>
    )
}

function WorkflowLocations({workflow, onEdit}: {workflow: Workflow; onEdit: () => void}) {
    if (!workflow.projectDir) return null

    const body = (
        <>
            <span className="micro-label">Project folder</span>
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
            <span className="micro-label">Progress</span>
            <span className="font-mono text-sm text-muted-foreground">
                {done}/{total}
            </span>
        </span>
    )
}
