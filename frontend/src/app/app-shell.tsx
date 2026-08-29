import {useState} from 'react'
import {Store as StoreIcon, Workflow as WorkflowIcon} from 'lucide-react'

import {AppView} from '@/app/view'
import {DeleteWorkflowDialog} from '@/features/workflows/components/delete-workflow-dialog'
import {DuplicateWorkflowDialog} from '@/features/workflows/components/duplicate-workflow-dialog'
import {ImportWorkflowDialog} from '@/features/workflows/components/import-workflow-dialog'
import {NewWorkflowDialog} from '@/features/workflows/components/new-workflow-dialog'
import {NoticeDialog} from '@/shared/components/notice-dialog'
import {PathPickerHost} from '@/shared/components/path-picker'
import {RailNav, type RailNavItem} from '@/shared/components/rail-nav'
import {SettingsDialog} from '@/features/settings/components/settings-dialog'
import {StoreRail} from '@/features/store/components/store-rail'
import {StoreSection} from '@/features/store/types'
import {StoreView} from '@/features/store/components/store-view'
import {WorkflowWorkspace} from '@/features/workflows/components/workflow-workspace'
import {WorkflowsRail} from '@/features/workflows/components/workflows-rail'
import {WorkingDialog} from '@/shared/components/working-dialog'
import {useStore} from '@/features/store/use-store'
import {useTour} from '@/features/onboarding/tour-context'
import {useToggle} from '@/shared/hooks/use-toggle'
import {useWorkflowStore} from '@/features/workflows/use-workflow-store'
import {useWorkflowTransfer} from '@/features/workflows/use-workflow-transfer'
import {t} from '@/shared/lib/i18n'
import type {Workflow, WorkflowLocations} from '@/features/workflows/types'

const NAV: RailNavItem<AppView>[] = [
    {id: AppView.Workflows, label: 'store.nav.workflows', icon: WorkflowIcon},
    {id: AppView.Store, label: 'store.nav.store', icon: StoreIcon, tour: 'store'},
]

/**
 * Everything under the tour, which is why this sits below TourProvider rather
 * than in App: the tour opens the store itself, and it can only do that from
 * inside its own context.
 */
export function AppShell() {
    const workflows = useWorkflowStore()
    const transfer = useWorkflowTransfer(workflows.importWorkflow)
    const tour = useTour()

    const rail = useToggle(true)
    const settings = useToggle()
    const newWorkflow = useToggle()
    const [chosenView, setChosenView] = useState<AppView>(AppView.Workflows)
    const [section, setSection] = useState<StoreSection>(StoreSection.Workflows)
    const [pendingDeleteId, setPendingDeleteId] = useState<string | null>(null)

    const view = tour.openStore ? AppView.Store : chosenView
    const inStore = view === AppView.Store

    /** Added work is the point of the store, so the app follows it back to the canvas. */
    const openAdded = (workflow: Workflow, locations: WorkflowLocations) => {
        workflows.importWorkflow(workflow, locations)
        setChosenView(AppView.Workflows)
    }

    const store = useStore(openAdded)

    const pendingDelete = workflows.workflows.find((workflow) => workflow.id === pendingDeleteId)

    const createWorkflow = (locations: WorkflowLocations) => {
        workflows.addWorkflow({
            name: t('app.workflow.defaultName', {count: workflows.workflows.length + 1}),
            ...locations,
        })
        newWorkflow.close()
    }

    const cancelDelete = () => setPendingDeleteId(null)

    const confirmDelete = () => {
        if (pendingDelete) workflows.deleteWorkflow(pendingDelete.id)
        cancelDelete()
    }

    // The tour opens the store rather than describing it, so leaving takes the tour
    // with you: the nav is never a control that looks live and does nothing.
    const changeView = (next: AppView) => {
        if (next !== AppView.Store) tour.leaveStore()
        setChosenView(next)
    }

    const nav = <RailNav items={NAV} activeId={view} onSelect={changeView} />

    return (
        <div className="flex h-screen gap-3 overflow-hidden bg-background p-3 text-foreground">
            {rail.on &&
                (inStore ? (
                    <StoreRail
                        nav={nav}
                        section={section}
                        workflowCount={store.templates.length}
                        roleCount={store.roles.length}
                        onSelect={setSection}
                    />
                ) : (
                    <WorkflowsRail
                        nav={nav}
                        workflows={workflows.workflows}
                        activeWorkflowId={workflows.activeWorkflowId}
                        onSelect={workflows.selectWorkflow}
                        onCreate={newWorkflow.open}
                        onImport={transfer.beginImport}
                        onDuplicate={workflows.requestDuplicate}
                        onExport={transfer.exportWorkflow}
                        onDelete={setPendingDeleteId}
                        onReorder={workflows.reorderWorkflow}
                    />
                ))}

            {inStore ? (
                <StoreView
                    store={store}
                    section={section}
                    railOpen={rail.on}
                    onToggleRail={rail.toggle}
                />
            ) : (
                <WorkflowWorkspace
                    store={workflows}
                    railOpen={rail.on}
                    onToggleRail={rail.toggle}
                    onOpenSettings={settings.open}
                />
            )}

            {pendingDelete && (
                <DeleteWorkflowDialog
                    key={pendingDelete.id}
                    workflow={pendingDelete}
                    onConfirm={confirmDelete}
                    onClose={cancelDelete}
                />
            )}

            {workflows.duplicating && (
                <DuplicateWorkflowDialog
                    key={workflows.duplicating.id}
                    workflow={workflows.duplicating}
                    onConfirm={workflows.confirmDuplicate}
                    onClose={workflows.cancelDuplicate}
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
        </div>
    )
}
