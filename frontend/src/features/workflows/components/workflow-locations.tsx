import {DirectoryField} from '@/shared/components/directory-field'
import {t} from '@/shared/lib/i18n'
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
                label={t('workflow.locations.label')}
                term="projectFolder"
                hint={t('workflow.locations.hint')}
                title={t('workflow.locations.picker')}
                value={projectDir}
                onChange={onProjectDirChange}
            />
        </div>
    )
}
