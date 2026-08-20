import {useState} from 'react'

import {ConfirmDialog} from '@/shared/components/confirm-dialog'
import {DialogShell} from '@/shared/components/dialog-shell'
import {WorkflowLocationsFields} from '@/features/workflows/components/workflow-locations'
import {Button} from '@/shared/ui/button'
import {useToggle} from '@/shared/hooks/use-toggle'
import type {Workflow, WorkflowLocations} from '@/features/workflows/types'

export function EditLocationsDialog({
    workflow,
    onSave,
    onClose,
}: {
    workflow: Workflow
    onSave: (locations: WorkflowLocations) => void
    onClose: () => void
}) {
    const [projectDir, setProjectDir] = useState(workflow.projectDir)
    const confirming = useToggle()

    const changed = projectDir !== workflow.projectDir
    const ready = changed && projectDir.length > 0

    const save = () => {
        if (ready) confirming.open()
    }

    const commit = () => {
        confirming.close()
        onSave({projectDir})
        onClose()
    }

    return (
        <DialogShell
            onClose={onClose}
            title="Project folder"
            description="Where this workflow runs. Locking fixes it for good."
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
            <WorkflowLocationsFields projectDir={projectDir} onProjectDirChange={setProjectDir} />

            {confirming.on && (
                <ConfirmDialog
                    title="Move this workflow?"
                    description="Steps that already ran keep their results, but every step from here on runs against the new folder."
                    confirmLabel="Save folder"
                    onConfirm={commit}
                    onClose={confirming.close}
                />
            )}
        </DialogShell>
    )
}
