import {useState} from 'react'

import {ConfirmDialog} from '@/shared/components/confirm-dialog'
import {DialogShell} from '@/shared/components/dialog-shell'
import {SessionLocationsFields} from '@/features/sessions/components/session-locations'
import {Button} from '@/shared/ui/button'
import {useToggle} from '@/shared/hooks/use-toggle'
import type {Session, SessionLocations} from '@/features/sessions/types'

export function EditLocationsDialog({
    session,
    onSave,
    onClose,
}: {
    session: Session
    onSave: (locations: SessionLocations) => void
    onClose: () => void
}) {
    const [workingDir, setWorkingDir] = useState(session.workingDir)
    const [contextDir, setContextDir] = useState(session.contextDir)
    const confirming = useToggle()

    const changed = workingDir !== session.workingDir || contextDir !== session.contextDir
    const ready = changed && workingDir.length > 0 && contextDir.length > 0

    const save = () => {
        if (ready) confirming.open()
    }

    const commit = () => {
        confirming.close()
        onSave({workingDir, contextDir})
        onClose()
    }

    return (
        <DialogShell
            onClose={onClose}
            title="Session directories"
            description="Where this session runs. Finalizing locks them."
            footer={
                <>
                    <span className="flex-1" />
                    <Button variant="outline" size="sm" onClick={onClose}>
                        Cancel
                    </Button>
                    <Button size="sm" disabled={!ready} onClick={save}>
                        Save
                    </Button>
                </>
            }
        >
            <SessionLocationsFields
                workingDir={workingDir}
                contextDir={contextDir}
                onWorkingDirChange={setWorkingDir}
                onContextDirChange={setContextDir}
            />

            {confirming.on && (
                <ConfirmDialog
                    title="Move this session?"
                    description="Nodes that already ran keep their reports, but every node from here on runs against the new directories."
                    confirmLabel="Save directories"
                    onConfirm={commit}
                    onClose={confirming.close}
                />
            )}
        </DialogShell>
    )
}
