import {useState} from 'react'
import {Check} from 'lucide-react'

import {Button} from '@/shared/ui/button'
import {DialogShell} from '@/shared/components/dialog-shell'
import {EffortTag} from '@/shared/components/effort-tag'
import {cn} from '@/shared/lib/utils'
import {t, tn} from '@/shared/lib/i18n'
import type {Role} from '@/features/roles/types'

export function RoleExportDialog({
    roles,
    onExport,
    onClose,
}: {
    roles: Role[]
    onExport: (roleIds: string[]) => void
    onClose: () => void
}) {
    const [pickedIds, setPickedIds] = useState<string[]>([])

    const toggle = (roleId: string) =>
        setPickedIds((current) =>
            current.includes(roleId) ? current.filter((id) => id !== roleId) : [...current, roleId],
        )

    const all = () => setPickedIds(roles.map((role) => role.id))
    const none = () => setPickedIds([])
    const send = () => onExport(pickedIds)

    const everyone = pickedIds.length === roles.length

    return (
        <DialogShell
            onClose={onClose}
            title={t('role.export.title')}
            description={t('role.export.description')}
            aside={
                <Button variant="ghost" size="sm" onClick={everyone ? none : all}>
                    {everyone ? t('role.export.clear') : t('role.export.selectAll')}
                </Button>
            }
            footer={
                <>
                    <p className="min-w-0 flex-1 text-sm text-muted-foreground">
                        {pickedIds.length === 0
                            ? t('role.export.hint')
                            : t('role.export.selected', {
                                  count: pickedIds.length,
                                  total: roles.length,
                              })}
                    </p>
                    <Button variant="ghost" size="sm" onClick={onClose}>
                        {t('role.form.cancel')}
                    </Button>
                    <Button size="sm" disabled={pickedIds.length === 0} onClick={send}>
                        {tn('role.export.confirmOne', 'role.export.confirmOther', pickedIds.length)}
                    </Button>
                </>
            }
        >
            <div className="flex flex-col gap-2 p-4">
                {roles.map((role) => (
                    <ExportRow
                        key={role.id}
                        role={role}
                        picked={pickedIds.includes(role.id)}
                        onToggle={toggle}
                    />
                ))}
            </div>
        </DialogShell>
    )
}

function ExportRow({
    role,
    picked,
    onToggle,
}: {
    role: Role
    picked: boolean
    onToggle: (roleId: string) => void
}) {
    return (
        <button
            type="button"
            aria-pressed={picked}
            onClick={() => onToggle(role.id)}
            className={cn(
                'flex w-full items-start gap-3 rounded-xl border p-3 text-left transition-colors outline-none focus-visible:ring-2 focus-visible:ring-ring',
                picked
                    ? 'border-live bg-live-tint'
                    : 'border-border bg-card hover:border-foreground/25 hover:bg-muted/40',
            )}
        >
            <span
                aria-hidden="true"
                className={cn(
                    'mt-0.5 flex size-4 shrink-0 items-center justify-center rounded-[5px] border transition-colors',
                    picked
                        ? 'border-live bg-live text-white'
                        : 'border-border-strong bg-background',
                )}
            >
                {picked && <Check className="size-3" strokeWidth={3} />}
            </span>

            <span className="flex min-w-0 flex-1 flex-col gap-1">
                <span className="flex items-center gap-2">
                    <span className="truncate text-base font-medium">{role.name}</span>
                    <EffortTag effort={role.effort} />
                </span>

                <span className="line-clamp-1 text-sm text-muted-foreground">
                    {role.description || t('role.card.noDescription')}
                </span>
            </span>
        </button>
    )
}
