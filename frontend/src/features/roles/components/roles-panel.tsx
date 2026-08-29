import {useState} from 'react'
import {Download, Plus, Upload} from 'lucide-react'

import {Button} from '@/shared/ui/button'
import {NoticeDialog} from '@/shared/components/notice-dialog'
import {RoleExportDialog} from '@/features/roles/components/role-export-dialog'
import {RoleFormDialog} from '@/features/roles/components/role-form-dialog'
import {RoleList} from '@/features/roles/components/role-list'
import {WorkingDialog} from '@/shared/components/working-dialog'
import {chooseFile, chooseSaveFile} from '@/shared/api/dialogs'
import {reportError} from '@/shared/lib/error-bus'
import {t, tn} from '@/shared/lib/i18n'
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
            reportError(cause, t('role.error.filePicker'))
            return ''
        }
    }

    const runExport = async (roleIds: string[]) => {
        const path = await pickPath(() =>
            chooseSaveFile(t('role.export.title'), exportFileName(), JSON_FILES),
        )
        if (!path) return

        const count = await transfer.exportRoles({roleIds, path}).catch(() => null)
        if (count === null) return

        exporting.close()
        setNotice({
            title: tn('role.export.doneOne', 'role.export.doneOther', count),
            description: t('role.export.doneBody'),
            detail: path,
        })
    }

    const runImport = async () => {
        const path = await pickPath(() => chooseFile(t('role.import.title'), JSON_FILES))
        if (!path) return

        const count = await transfer.importRoles(path).catch(() => null)
        if (count === null) return

        setNotice({
            title: tn('role.import.doneOne', 'role.import.doneOther', count),
            description: t('role.import.doneBody'),
            detail: path,
        })
    }

    return (
        <section className="flex flex-col">
            <div className="flex items-start gap-3 px-4 pt-4 pb-3">
                <div className="flex min-w-0 flex-1 flex-col gap-1">
                    <h3 className="text-lg font-medium">
                        {onPick ? t('role.panel.pickTitle') : t('role.panel.title')}
                    </h3>
                    <p className="text-sm text-muted-foreground">{t('role.panel.subtitle')}</p>
                </div>

                {roles.length > 0 && (
                    <Button variant="outline" size="sm" className="shrink-0" onClick={newRole}>
                        <Plus />
                        {t('role.action.new')}
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
                        {t('role.panel.transfer')}
                    </p>

                    <Button variant="ghost" size="sm" onClick={runImport}>
                        <Upload />
                        {t('role.action.import')}
                    </Button>

                    {roles.length > 0 && (
                        <Button variant="ghost" size="sm" onClick={exporting.open}>
                            <Download />
                            {t('role.action.export')}
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
                <WorkingDialog
                    title={t('role.export.working')}
                    description={t('role.export.workingBody')}
                />
            )}

            {transfer.receiving && (
                <WorkingDialog
                    title={t('role.import.working')}
                    description={t('role.import.workingBody')}
                />
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
