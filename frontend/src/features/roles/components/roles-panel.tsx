import {useState} from 'react'
import {Download, Plus, Upload} from 'lucide-react'

import {Button} from '@/shared/ui/button'
import {NoticeDialog} from '@/shared/components/notice-dialog'
import {RoleExportDialog} from '@/features/roles/components/role-export-dialog'
import {RoleFormDialog} from '@/features/roles/components/role-form-dialog'
import {RoleList} from '@/features/roles/components/role-list'
import {WorkingDialog} from '@/shared/components/working-dialog'
import {chooseFile, chooseSaveFile} from '@/shared/api/dialogs'
import {pluralize} from '@/shared/lib/format'
import {reportError} from '@/shared/lib/error-bus'
import {useRoles, useRoleTransfer} from '@/features/roles/use-roles'
import {useToggle} from '@/shared/hooks/use-toggle'
import type {DraftContext, Role} from '@/features/roles/types'

const JSON_FILES = '*.json'

type RoleEdit = {role: Role | null}

type Notice = {title: string; description: string; detail: string}

function exportFileName() {
    return `nexo-roles-${new Date().toISOString().slice(0, 10)}.json`
}

export function RolesPanel({
    context,
    onPick,
}: {
    context?: DraftContext
    onPick?: (role: Role) => void
}) {
    const {roles, loading, removeRole} = useRoles()
    const transfer = useRoleTransfer()

    const [editing, setEditing] = useState<RoleEdit | null>(null)
    const [notice, setNotice] = useState<Notice | null>(null)
    const exporting = useToggle()

    const newRole = () => setEditing({role: null})
    const editRole = (role: Role) => setEditing({role})
    const closeForm = () => setEditing(null)
    const dismissNotice = () => setNotice(null)

    const pickPath = async (pick: () => Promise<string>) => {
        try {
            return await pick()
        } catch (cause) {
            reportError(cause, 'Could not open the file picker')
            return ''
        }
    }

    const runExport = async (roleIds: string[]) => {
        const path = await pickPath(() =>
            chooseSaveFile('Export roles', exportFileName(), JSON_FILES),
        )
        if (!path) return

        const count = await transfer.exportRoles({roleIds, path}).catch(() => null)
        if (count === null) return

        exporting.close()
        setNotice({
            title: `Exported ${pluralize(count, 'role')}`,
            description: 'The file is yours to keep, move, or hand to someone else.',
            detail: path,
        })
    }

    const runImport = async () => {
        const path = await pickPath(() => chooseFile('Import roles', JSON_FILES))
        if (!path) return

        const count = await transfer.importRoles(path).catch(() => null)
        if (count === null) return

        setNotice({
            title: `Imported ${pluralize(count, 'role')}`,
            description: 'They are ready to build steps from.',
            detail: path,
        })
    }

    return (
        <section className="flex flex-col">
            <div className="flex items-start gap-3 px-4 pt-4 pb-3">
                <div className="flex min-w-0 flex-1 flex-col gap-1">
                    <h3 className="text-lg font-medium">
                        {onPick ? 'Pick a role' : 'Roles your steps start from'}
                    </h3>
                    <p className="text-sm text-muted-foreground">
                        One kind of work: the agent's role, its inputs, and how hard to try.
                    </p>
                </div>

                {roles.length > 0 && (
                    <Button variant="outline" size="sm" className="shrink-0" onClick={newRole}>
                        <Plus />
                        New role
                    </Button>
                )}
            </div>

            <div className="border-t border-border">
                <RoleList
                    roles={roles}
                    loading={loading}
                    onPick={onPick ?? editRole}
                    onEdit={editRole}
                    onRemove={removeRole}
                    onCreate={newRole}
                />
            </div>

            {!onPick && (
                <div className="flex items-center gap-1 border-t border-border px-4 py-3">
                    <p className="min-w-0 flex-1 truncate text-sm text-muted-foreground">
                        Roles travel as a .json file.
                    </p>

                    <Button variant="ghost" size="sm" onClick={runImport}>
                        <Upload />
                        Import
                    </Button>

                    {roles.length > 0 && (
                        <Button variant="ghost" size="sm" onClick={exporting.open}>
                            <Download />
                            Export
                        </Button>
                    )}
                </div>
            )}

            {editing && (
                <RoleFormDialog role={editing.role} context={context} onClose={closeForm} />
            )}

            {exporting.on && (
                <RoleExportDialog roles={roles} onExport={runExport} onClose={exporting.close} />
            )}

            {transfer.sending && (
                <WorkingDialog title="Exporting roles" description="Writing the file. Hold on." />
            )}

            {transfer.receiving && (
                <WorkingDialog title="Importing roles" description="Reading the file. Hold on." />
            )}

            {notice && (
                <NoticeDialog
                    title={notice.title}
                    description={notice.description}
                    detail={notice.detail}
                    onClose={dismissNotice}
                />
            )}
        </section>
    )
}
