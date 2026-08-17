import {Settings} from 'lucide-react'

import {Button} from '@/shared/ui/button'
import {ConfirmDialog} from '@/shared/components/confirm-dialog'
import {HeaderAction} from '@/features/sessions/components/canvas/header/header-action'
import {RailToggle} from '@/features/sessions/components/canvas/header/rail-toggle'
import {RunCost} from '@/features/sessions/components/canvas/header/run-cost'
import {SessionDirectories} from '@/features/sessions/components/canvas/header/session-directories'
import {SessionIdentity} from '@/features/sessions/components/canvas/header/session-identity'
import {SessionMenu} from '@/features/sessions/components/canvas/session-menu'
import {SessionName} from '@/features/sessions/components/canvas/header/session-name'
import {Tooltip, TooltipContent, TooltipTrigger} from '@/shared/ui/tooltip'
import {
    SessionActionId,
    sessionActions,
    type SessionActionHandlers,
} from '@/features/sessions/session-actions'
import {CANCEL_CONFIRM, FINALIZE_CONFIRM} from '@/features/sessions/session-copy'
import {useToggle} from '@/shared/hooks/use-toggle'
import type {Session} from '@/features/sessions/types'

export function SessionHeader({
    session,
    cancelling,
    onRename,
    onEditLocations,
    onFinalize,
    onRun,
    onCancel,
    onClone,
    onNewNode,
    onOpenSettings,
    railOpen,
    onToggleRail,
}: {
    session: Session | null
    cancelling: boolean
    onRename: (name: string) => void
    onEditLocations: () => void
    onFinalize: () => void
    onRun: () => void
    onCancel: (onSettled: () => void) => void
    onClone: () => void
    onNewNode: () => void
    onOpenSettings: () => void
    railOpen: boolean
    onToggleRail: () => void
}) {
    const confirmingFinalize = useToggle()
    const confirmingCancel = useToggle()

    const handlers: SessionActionHandlers = {
        [SessionActionId.NewNode]: onNewNode,
        [SessionActionId.Clone]: onClone,
        [SessionActionId.Finalize]: confirmingFinalize.open,
        [SessionActionId.Run]: onRun,
        [SessionActionId.Cancel]: confirmingCancel.open,
    }

    const actions = sessionActions(session)

    const finalize = () => {
        confirmingFinalize.close()
        onFinalize()
    }

    const cancelRun = () => onCancel(confirmingCancel.close)

    return (
        <header className="shrink-0">
            <div className="flex h-14 items-center gap-3 border-b border-border-strong bg-card px-4">
                <RailToggle open={railOpen} onToggle={onToggleRail} />

                <SessionMenu
                    session={session}
                    actions={actions}
                    handlers={handlers}
                    onEditLocations={onEditLocations}
                    onOpenSettings={onOpenSettings}
                />

                {session && (
                    <div className="flex min-w-0 flex-1 items-center gap-3">
                        <SessionName session={session} onRename={onRename} />
                        <SessionDirectories session={session} onEdit={onEditLocations} />
                        <SessionIdentity session={session} />
                        <RunCost session={session} />
                    </div>
                )}

                {!session && <span className="flex-1" />}

                <div className="hidden shrink-0 items-center gap-2 lg:flex">
                    {actions.map((action) => (
                        <HeaderAction
                            key={action.id}
                            action={action}
                            onClick={handlers[action.id]}
                        />
                    ))}

                    {actions.length > 0 && <span className="mx-1 h-5 w-px bg-border" />}

                    <Tooltip>
                        <TooltipTrigger asChild>
                            <Button
                                variant="ghost"
                                size="icon"
                                aria-label="Settings"
                                onClick={onOpenSettings}
                            >
                                <Settings />
                            </Button>
                        </TooltipTrigger>
                        <TooltipContent side="bottom">Settings</TooltipContent>
                    </Tooltip>
                </div>
            </div>

            {confirmingFinalize.on && session && (
                <ConfirmDialog
                    title={`Finalize “${session.name}”?`}
                    description={FINALIZE_CONFIRM.description}
                    confirmLabel={FINALIZE_CONFIRM.confirmLabel}
                    onConfirm={finalize}
                    onClose={confirmingFinalize.close}
                />
            )}

            {confirmingCancel.on && session && (
                <ConfirmDialog
                    title={CANCEL_CONFIRM.title}
                    description={CANCEL_CONFIRM.description}
                    confirmLabel={cancelling ? 'Cancelling…' : 'Cancel run'}
                    dismissLabel={CANCEL_CONFIRM.dismissLabel}
                    destructive
                    busy={cancelling}
                    onConfirm={cancelRun}
                    onClose={confirmingCancel.close}
                />
            )}
        </header>
    )
}
