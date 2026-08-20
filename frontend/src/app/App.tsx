import {useState} from 'react'

import {DeleteWorkflowDialog} from '@/features/workflows/components/delete-workflow-dialog'
import {ImportWorkflowDialog} from '@/features/workflows/components/import-workflow-dialog'
import {NewWorkflowDialog} from '@/features/workflows/components/new-workflow-dialog'
import {NoticeDialog} from '@/shared/components/notice-dialog'
import {PathPickerHost} from '@/shared/components/path-picker'
import {WorkflowWorkspace} from '@/features/workflows/components/workflow-workspace'
import {WorkflowsRail} from '@/features/workflows/components/workflows-rail'
import {SettingsDialog} from '@/features/settings/components/settings-dialog'
import {WorkingDialog} from '@/shared/components/working-dialog'
import {OnboardingDialog} from '@/features/onboarding/components/onboarding-dialog'
import {TourLayer} from '@/features/onboarding/components/tour-layer'
import {TourProvider} from '@/features/onboarding/components/tour-provider'
import {OnboardingPhase} from '@/features/onboarding/types'
import {useOnboarding} from '@/features/onboarding/use-onboarding'
import {useWorkflowStore} from '@/features/workflows/use-workflow-store'
import {useWorkflowTransfer} from '@/features/workflows/use-workflow-transfer'
import {useLanguage} from '@/shared/hooks/use-language'
import {useToggle} from '@/shared/hooks/use-toggle'
import {t} from '@/shared/lib/i18n'
import type {WorkflowLocations} from '@/features/workflows/types'

function App() {
    useLanguage()

    const store = useWorkflowStore()
    const transfer = useWorkflowTransfer(store.importWorkflow)
    const onboarding = useOnboarding()

    const rail = useToggle(true)
    const settings = useToggle()
    const newWorkflow = useToggle()
    const [pendingDeleteId, setPendingDeleteId] = useState<string | null>(null)

    const pendingDelete = store.workflows.find((workflow) => workflow.id === pendingDeleteId)

    const createWorkflow = (locations: WorkflowLocations) => {
        store.addWorkflow({
            name: t('app.workflow.defaultName', {count: store.workflows.length + 1}),
            ...locations,
        })
        newWorkflow.close()
    }

    const cancelDelete = () => setPendingDeleteId(null)

    const confirmDelete = () => {
        if (pendingDelete) store.deleteWorkflow(pendingDelete.id)
        cancelDelete()
    }

    return (
        <TourProvider active={onboarding.phase === OnboardingPhase.Tour} onDone={onboarding.finish}>
            <div className="flex h-screen gap-3 overflow-hidden bg-background p-3 text-foreground">
                {rail.on && (
                    <WorkflowsRail
                        workflows={store.workflows}
                        activeWorkflowId={store.activeWorkflowId}
                        onSelect={store.selectWorkflow}
                        onCreate={newWorkflow.open}
                        onImport={transfer.beginImport}
                        onDuplicate={store.duplicateWorkflow}
                        onExport={transfer.exportWorkflow}
                        onDelete={setPendingDeleteId}
                        onReorder={store.reorderWorkflow}
                    />
                )}

                <WorkflowWorkspace
                    store={store}
                    railOpen={rail.on}
                    onToggleRail={rail.toggle}
                    onOpenSettings={settings.open}
                />

                {pendingDelete && (
                    <DeleteWorkflowDialog
                        key={pendingDelete.id}
                        workflow={pendingDelete}
                        onConfirm={confirmDelete}
                        onClose={cancelDelete}
                    />
                )}

                {newWorkflow.on && (
                    <NewWorkflowDialog onCreate={createWorkflow} onClose={newWorkflow.close} />
                )}

                {transfer.pending && (
                    <ImportWorkflowDialog
                        key={transfer.pending.id}
                        workflow={transfer.pending}
                        onImport={transfer.confirmImport}
                        onClose={transfer.cancelImport}
                    />
                )}

                {transfer.reading && (
                    <WorkingDialog
                        title={t('app.transfer.importing.title')}
                        description={t('app.transfer.importing.description')}
                    />
                )}

                {transfer.writing && (
                    <WorkingDialog
                        title={t('app.transfer.exporting.title')}
                        description={t('app.transfer.exporting.description')}
                    />
                )}

                {transfer.notice && (
                    <NoticeDialog
                        title={transfer.notice.title}
                        description={transfer.notice.description}
                        detail={transfer.notice.detail}
                        onClose={transfer.dismissNotice}
                    />
                )}

                <SettingsDialog open={settings.on} onOpenChange={settings.set} />

                <PathPickerHost />

                <OnboardingDialog onboarding={onboarding} />

                <TourLayer />
            </div>
        </TourProvider>
    )
}

export default App
