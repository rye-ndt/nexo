import {useState} from 'react'

import {Button} from '@/shared/ui/button'
import {DialogShell} from '@/shared/components/dialog-shell'
import {WorkflowLocationsFields} from '@/features/workflows/components/workflow-locations'
import {EffortTag} from '@/shared/components/effort-tag'
import {pluralize} from '@/shared/lib/format'
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
            title="Import workflow"
            description="Where this workflow runs on this machine."
            footer={
                <>
                    <span className="flex-1" />
                    <Button variant="outline" size="sm" onClick={onClose}>
                        Cancel
                    </Button>
                    <Button size="sm" disabled={!located} onClick={confirm}>
                        Import workflow
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
            <span className="micro-label">From the file</span>

            <div className="flex flex-col gap-1">
                <span className="text-base font-medium break-words">{workflow.name}</span>
                <span className="text-sm text-muted-foreground">
                    {pluralize(workflow.steps.length, 'step')}
                </span>
            </div>

            <div className="flex max-h-56 flex-col gap-1 overflow-y-auto">
                {workflow.steps.map((step) => (
                    <StepRow key={step.id} step={step} />
                ))}
            </div>

            <p className="text-sm text-muted-foreground">
                These steps run on their own: the roles they were built from stay on the machine
                that exported them.
            </p>
        </section>
    )
}

function StepRow({step}: {step: Step}) {
    return (
        <div className="flex items-center gap-3 rounded-lg border border-border bg-card px-3 py-2">
            <span className="min-w-0 flex-1 truncate text-base">
                {step.title || 'Untitled step'}
            </span>
            {step.spec && <EffortTag effort={step.spec.effort} />}
        </div>
    )
}
