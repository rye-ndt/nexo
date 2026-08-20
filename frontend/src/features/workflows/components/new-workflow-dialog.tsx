import {useState} from 'react'

import {DialogShell} from '@/shared/components/dialog-shell'
import {RequiredAgents} from '@/features/workflows/components/required-agents'
import {WorkflowLocationsFields} from '@/features/workflows/components/workflow-locations'
import {Button} from '@/shared/ui/button'
import {useRequiredAgents} from '@/features/workflows/use-required-agents'
import type {WorkflowLocations} from '@/features/workflows/types'

export function NewWorkflowDialog({
    onCreate,
    onClose,
}: {
    onCreate: (locations: WorkflowLocations) => void
    onClose: () => void
}) {
    const agents = useRequiredAgents()

    const [projectDir, setProjectDir] = useState('')

    const locations = {projectDir: projectDir.trim()}
    const located = locations.projectDir.length > 0

    const create = () => {
        if (!located || !agents.ready) return
        onCreate(locations)
    }

    return (
        <DialogShell
            onClose={onClose}
            title="New workflow"
            description="Where this workflow runs, and the agents it needs."
            footer={
                <>
                    <span className="flex-1" />
                    <Button variant="outline" size="sm" onClick={onClose}>
                        Cancel
                    </Button>
                    <Button size="sm" disabled={!located || !agents.ready} onClick={create}>
                        Create workflow
                    </Button>
                </>
            }
        >
            <div className="border-b border-border">
                <WorkflowLocationsFields
                    projectDir={projectDir}
                    onProjectDirChange={setProjectDir}
                />
            </div>

            <RequiredAgents
                required={agents.required}
                loading={agents.loading}
                controls={agents.controls}
            />
        </DialogShell>
    )
}
