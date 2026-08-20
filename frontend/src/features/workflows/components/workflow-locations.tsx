import {DirectoryField} from '@/shared/components/directory-field'
import type {WorkflowLocations} from '@/features/workflows/types'

export function WorkflowLocationsFields({
    projectDir,
    onProjectDirChange,
}: WorkflowLocations & {
    onProjectDirChange: (path: string) => void
}) {
    return (
        <div className="flex flex-col gap-6 p-4">
            <DirectoryField
                label="Project folder"
                term="projectFolder"
                hint="Pick it now — it is fixed once the workflow is locked."
                title="Choose the project folder"
                value={projectDir}
                onChange={onProjectDirChange}
            />
        </div>
    )
}
