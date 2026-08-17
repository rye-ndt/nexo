import {useState} from 'react'

import {DeleteSessionDialog} from '@/features/sessions/components/delete-session-dialog'
import {ImportSessionDialog} from '@/features/sessions/components/import-session-dialog'
import {NewSessionDialog} from '@/features/sessions/components/new-session-dialog'
import {NoticeDialog} from '@/shared/components/notice-dialog'
import {PathPickerHost} from '@/shared/components/path-picker'
import {SessionWorkspace} from '@/features/sessions/components/session-workspace'
import {SessionsRail} from '@/features/sessions/components/sessions-rail'
import {SettingsDialog} from '@/features/settings/components/settings-dialog'
import {WorkingDialog} from '@/shared/components/working-dialog'
import {WelcomeDialog} from '@/features/onboarding/components/welcome-dialog'
import {useDependencies} from '@/features/onboarding/use-dependencies'
import {useSessionStore} from '@/features/sessions/use-session-store'
import {useSessionTransfer} from '@/features/sessions/use-session-transfer'
import {useToggle} from '@/shared/hooks/use-toggle'
import type {SessionLocations} from '@/features/sessions/types'

function App() {
    const store = useSessionStore()
    const transfer = useSessionTransfer(store.importSession)
    const dependencies = useDependencies()

    const rail = useToggle(true)
    const settings = useToggle()
    const newSession = useToggle()
    const [pendingDeleteId, setPendingDeleteId] = useState<string | null>(null)

    const pendingDelete = store.sessions.find((session) => session.id === pendingDeleteId)

    const createSession = (locations: SessionLocations) => {
        store.addSession({name: `Session ${store.sessions.length + 1}`, ...locations})
        newSession.close()
    }

    const cancelDelete = () => setPendingDeleteId(null)

    const confirmDelete = () => {
        if (pendingDelete) store.deleteSession(pendingDelete.id)
        cancelDelete()
    }

    return (
        <div className="flex h-screen gap-3 overflow-hidden bg-background p-3 text-foreground">
            {rail.on && (
                <SessionsRail
                    sessions={store.sessions}
                    activeSessionId={store.activeSessionId}
                    onSelect={store.selectSession}
                    onCreate={newSession.open}
                    onImport={transfer.beginImport}
                    onClone={store.cloneSession}
                    onExport={transfer.exportSession}
                    onDelete={setPendingDeleteId}
                />
            )}

            <SessionWorkspace
                store={store}
                railOpen={rail.on}
                onToggleRail={rail.toggle}
                onOpenSettings={settings.open}
            />

            {pendingDelete && (
                <DeleteSessionDialog
                    key={pendingDelete.id}
                    session={pendingDelete}
                    onConfirm={confirmDelete}
                    onClose={cancelDelete}
                />
            )}

            {newSession.on && (
                <NewSessionDialog onCreate={createSession} onClose={newSession.close} />
            )}

            {transfer.pending && (
                <ImportSessionDialog
                    key={transfer.pending.id}
                    session={transfer.pending}
                    onImport={transfer.confirmImport}
                    onClose={transfer.cancelImport}
                />
            )}

            {transfer.reading && (
                <WorkingDialog title="Importing session" description="Reading the file. Hold on." />
            )}

            {transfer.writing && (
                <WorkingDialog title="Exporting session" description="Writing the file. Hold on." />
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

            {dependencies.ready && dependencies.required && (
                <WelcomeDialog
                    dependencies={dependencies.dependencies}
                    ratio={dependencies.ratio}
                    settled={dependencies.settled}
                    canContinue={dependencies.canContinue}
                    failed={dependencies.failed}
                    onRetry={dependencies.retry}
                    onStart={dependencies.dismiss}
                />
            )}
        </div>
    )
}

export default App
