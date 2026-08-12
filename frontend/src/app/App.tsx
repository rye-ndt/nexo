import {useState} from 'react'

import {DeleteSessionDialog} from '@/features/sessions/components/delete-session-dialog'
import {NewSessionDialog} from '@/features/sessions/components/new-session-dialog'
import {PathPickerHost} from '@/shared/components/path-picker'
import {SessionWorkspace} from '@/features/sessions/components/session-workspace'
import {SessionsRail} from '@/features/sessions/components/sessions-rail'
import {SettingsDialog} from '@/features/settings/components/settings-dialog'
import {WelcomeDialog} from '@/features/onboarding/components/welcome-dialog'
import {useDependencies} from '@/features/onboarding/use-dependencies'
import {useSessionStore} from '@/features/sessions/use-session-store'
import {useToggle} from '@/shared/hooks/use-toggle'
import type {SessionLocations} from '@/features/sessions/types'

function App() {
    const store = useSessionStore()
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
                    onClone={store.cloneSession}
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
