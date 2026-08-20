import {useState} from 'react'

import {Button} from '@/shared/ui/button'
import {DialogShell} from '@/shared/components/dialog-shell'
import {WorkflowLocationsFields} from '@/features/workflows/components/workflow-locations'
import {EffortTag} from '@/shared/components/effort-tag'
import {t, tn} from '@/shared/lib/i18n'
import type {Workflow, WorkflowLocations, Step} from '@/features/workflows/types'

export function ImportWorkflowDialog({
    workflow,
    onImport,
    onClose,
}: {
    workflow: Workflow
    onImport: (locations: WorkflowLocations) => void
    onClose: () => void
}) {
    const [projectDir, setProjectDir] = useState(workflow.projectDir ?? '')

    const locations = {projectDir: projectDir.trim()}
    const located = locations.projectDir.length > 0

    const confirm = () => {
        if (!located) return

        onImport(locations)
        onClose()
    }

    return (
        <DialogShell
            onClose={onClose}
            title={t('workflow.import.title')}
            description={t('workflow.import.description')}
            footer={
                <>
                    <span className="flex-1" />
                    <Button variant="outline" size="sm" onClick={onClose}>
                        {t('workflow.dialog.cancel')}
                    </Button>
                    <Button size="sm" disabled={!located} onClick={confirm}>
                        {t('workflow.import.confirm')}
                    </Button>
                </>
            }
        >
            <FileSummary workflow={workflow} />

            <WorkflowLocationsFields projectDir={projectDir} onProjectDirChange={setProjectDir} />
        </DialogShell>
    )
}

function FileSummary({workflow}: {workflow: Workflow}) {
    return (
        <section className="flex flex-col gap-3 border-b border-border px-4 py-4">
            <span className="micro-label">{t('workflow.import.fromFile')}</span>

            <div className="flex flex-col gap-1">
                <span className="text-base font-medium break-words">{workflow.name}</span>
                <span className="text-sm text-muted-foreground">
                    {tn(
                        'workflow.import.steps.one',
                        'workflow.import.steps.other',
                        workflow.steps.length,
                    )}
                </span>
            </div>

            <div className="flex max-h-56 flex-col gap-1 overflow-y-auto">
                {workflow.steps.map((step) => (
                    <StepRow key={step.id} step={step} />
                ))}
            </div>

            <p className="text-sm text-muted-foreground">{t('workflow.import.rolesStayBehind')}</p>
        </section>
    )
}

function StepRow({step}: {step: Step}) {
    return (
        <div className="flex items-center gap-3 rounded-lg border border-border bg-card px-3 py-2">
            <span className="min-w-0 flex-1 truncate text-base">
                {step.title || t('workflow.step.untitled')}
            </span>
            {step.spec && <EffortTag effort={step.spec.effort} />}
        </div>
    )
}
