import {useState} from 'react'

import {ConfirmDialog} from '@/shared/components/confirm-dialog'
import {DialogShell} from '@/shared/components/dialog-shell'
import {WorkflowLocationsFields} from '@/features/workflows/components/workflow-locations'
import {Button} from '@/shared/ui/button'
import {useToggle} from '@/shared/hooks/use-toggle'
import {t} from '@/shared/lib/i18n'
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
            title={t('workflow.locations.title')}
            description={t('workflow.locations.description')}
            footer={
                <>
                    <span className="flex-1" />
                    <Button variant="outline" size="sm" onClick={onClose}>
                        {t('workflow.dialog.cancel')}
                    </Button>
                    <Button size="sm" disabled={!ready} onClick={save}>
                        {t('workflow.dialog.save')}
                    </Button>
                </>
            }
        >
            <WorkflowLocationsFields projectDir={projectDir} onProjectDirChange={setProjectDir} />

            {confirming.on && (
                <ConfirmDialog
                    title={t('workflow.locations.moveTitle')}
                    description={t('workflow.locations.moveDescription')}
                    confirmLabel={t('workflow.locations.moveConfirm')}
                    onConfirm={commit}
                    onClose={confirming.close}
                />
            )}
        </DialogShell>
    )
}
