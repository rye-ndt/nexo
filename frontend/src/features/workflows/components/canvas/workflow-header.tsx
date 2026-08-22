import {Settings} from 'lucide-react'

import {Button} from '@/shared/ui/button'
import {ConfirmDialog} from '@/shared/components/confirm-dialog'
import {HeaderAction} from '@/features/workflows/components/canvas/header/header-action'
import {RailToggle} from '@/features/workflows/components/canvas/header/rail-toggle'
import {RunCost} from '@/features/workflows/components/canvas/header/run-cost'
import {WorkflowDirectories} from '@/features/workflows/components/canvas/header/workflow-directories'
import {WorkflowIdentity} from '@/features/workflows/components/canvas/header/workflow-identity'
import {WorkflowMenu} from '@/features/workflows/components/canvas/workflow-menu'
import {WorkflowName} from '@/features/workflows/components/canvas/header/workflow-name'
import {Tooltip, TooltipContent, TooltipTrigger} from '@/shared/ui/tooltip'
import {
    WorkflowActionId,
    workflowActions,
    type WorkflowActionHandlers,
} from '@/features/workflows/workflow-actions'
import {
    cancelConfirm,
    lockConfirm,
    pauseConfirm,
    unlockConfirm,
} from '@/features/workflows/workflow-copy'
import {useToggle} from '@/shared/hooks/use-toggle'
import {t} from '@/shared/lib/i18n'
import type {Workflow} from '@/features/workflows/types'

export function WorkflowHeader({
    workflow,
    cancelling,
    pausing,
    onRename,
    onEditLocations,
    onLock,
    onUnlock,
    onRun,
    onPause,
    onResume,
    onCancel,
    onDuplicate,
    onNewStep,
    onOpenSettings,
    railOpen,
    onToggleRail,
}: {
    workflow: Workflow | null
    cancelling: boolean
    pausing: boolean
    onRename: (name: string) => void
    onEditLocations: () => void
    onLock: () => void
    onUnlock: () => void
    onRun: () => void
    onPause: (onSettled: () => void) => void
    onResume: () => void
    onCancel: (onSettled: () => void) => void
    onDuplicate: () => void
    onNewStep: () => void
    onOpenSettings: () => void
    railOpen: boolean
    onToggleRail: () => void
}) {
    const confirmingLock = useToggle()
    const confirmingUnlock = useToggle()
    const confirmingPause = useToggle()
    const confirmingCancel = useToggle()

    const handlers: WorkflowActionHandlers = {
        [WorkflowActionId.NewStep]: onNewStep,
        [WorkflowActionId.Duplicate]: onDuplicate,
        [WorkflowActionId.Lock]: confirmingLock.open,
        [WorkflowActionId.Unlock]: confirmingUnlock.open,
        [WorkflowActionId.Run]: onRun,
        [WorkflowActionId.Pause]: confirmingPause.open,
        [WorkflowActionId.Resume]: onResume,
        [WorkflowActionId.Cancel]: confirmingCancel.open,
    }

    const actions = workflowActions(workflow)

    const lock = () => {
        confirmingLock.close()
        onLock()
    }

    const unlock = () => {
        confirmingUnlock.close()
        onUnlock()
    }

    const pauseRun = () => onPause(confirmingPause.close)

    const cancelRun = () => onCancel(confirmingCancel.close)

    return (
        <header className="shrink-0">
            <div className="flex h-14 items-center gap-3 border-b border-border-strong bg-card px-4">
                <RailToggle open={railOpen} onToggle={onToggleRail} />

                <WorkflowMenu
                    workflow={workflow}
                    actions={actions}
                    handlers={handlers}
                    onEditLocations={onEditLocations}
                    onOpenSettings={onOpenSettings}
                />

                {workflow && (
                    <div className="flex min-w-0 flex-1 items-center gap-3">
                        <WorkflowName workflow={workflow} onRename={onRename} />
                        <WorkflowDirectories workflow={workflow} onEdit={onEditLocations} />
                        <WorkflowIdentity workflow={workflow} />
                        <RunCost workflow={workflow} />
                    </div>
                )}

                {!workflow && <span className="flex-1" />}

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
                                aria-label={t('canvas.header.settings')}
                                onClick={onOpenSettings}
                            >
                                <Settings />
                            </Button>
                        </TooltipTrigger>
                        <TooltipContent side="bottom">{t('canvas.header.settings')}</TooltipContent>
                    </Tooltip>
                </div>
            </div>

            {confirmingLock.on && workflow && (
                <ConfirmDialog
                    title={t('canvas.header.lockTitle', {name: workflow.name})}
                    description={lockConfirm().description}
                    confirmLabel={lockConfirm().confirmLabel}
                    onConfirm={lock}
                    onClose={confirmingLock.close}
                />
            )}

            {confirmingUnlock.on && workflow && (
                <ConfirmDialog
                    title={unlockConfirm().title}
                    description={unlockConfirm().description}
                    confirmLabel={unlockConfirm().confirmLabel}
                    dismissLabel={unlockConfirm().dismissLabel}
                    destructive
                    onConfirm={unlock}
                    onClose={confirmingUnlock.close}
                />
            )}

            {confirmingPause.on && workflow && (
                <ConfirmDialog
                    title={pauseConfirm().title}
                    description={pauseConfirm().description}
                    confirmLabel={
                        pausing ? t('canvas.header.pausing') : pauseConfirm().confirmLabel
                    }
                    dismissLabel={pauseConfirm().dismissLabel}
                    busy={pausing}
                    onConfirm={pauseRun}
                    onClose={confirmingPause.close}
                />
            )}

            {confirmingCancel.on && workflow && (
                <ConfirmDialog
                    title={cancelConfirm().title}
                    description={cancelConfirm().description}
                    confirmLabel={
                        cancelling ? t('canvas.header.cancelling') : t('canvas.header.cancelRun')
                    }
                    dismissLabel={cancelConfirm().dismissLabel}
                    destructive
                    busy={cancelling}
                    onConfirm={cancelRun}
                    onClose={confirmingCancel.close}
                />
            )}
        </header>
    )
}
