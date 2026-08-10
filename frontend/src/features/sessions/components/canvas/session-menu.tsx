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
import {sessionPaths, sessionProgress} from '@/features/sessions/graph'
import {useToggle} from '@/shared/hooks/use-toggle'
import type {SessionAction, SessionActionHandlers} from '@/features/sessions/session-actions'
import type {Session} from '@/features/sessions/types'

export function SessionMenu({
    session,
    actions,
    handlers,
    onEditLocations,
    onOpenSettings,
}: {
    session: Session | null
    actions: SessionAction[]
    handlers: SessionActionHandlers
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
                <Button variant="ghost" size="icon" aria-label="Session menu" className="lg:hidden">
                    <Menu />
                </Button>
            </SheetTrigger>

            <SheetContent side="left" className="gap-0">
                <SheetHeader className="h-14 shrink-0 justify-center border-b border-border px-4 py-0">
                    <SheetTitle className="micro-label">Menu</SheetTitle>
                    <SheetDescription className="sr-only">
                        Directories, progress and actions for this session.
                    </SheetDescription>
                </SheetHeader>

                <div className="flex min-h-0 flex-1 flex-col gap-1 overflow-x-hidden overflow-y-auto p-2">
                    {session && (
                        <>
                            <SessionLocations session={session} onEdit={pick(onEditLocations)} />
                            <SessionProgress session={session} />

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

function SessionLocations({session, onEdit}: {session: Session; onEdit: () => void}) {
    const paths = sessionPaths(session)
    if (paths.length === 0) return null

    const body = (
        <>
            <span className="micro-label">Directories</span>
            {paths.map((path) => (
                <span key={path} className="truncate font-mono text-sm text-muted-foreground">
                    {path}
                </span>
            ))}
        </>
    )

    if (session.finalized) return <span className="flex flex-col gap-2 px-3 py-2">{body}</span>

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

function SessionProgress({session}: {session: Session}) {
    const {done, total} = sessionProgress(session)

    return (
        <span className="flex items-center justify-between gap-3 px-3 py-2">
            <span className="micro-label">Progress</span>
            <span className="font-mono text-sm text-muted-foreground">
                {done}/{total}
            </span>
        </span>
    )
}
