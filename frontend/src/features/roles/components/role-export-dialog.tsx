import {useState} from 'react'
import {Check} from 'lucide-react'

import {Button} from '@/shared/ui/button'
import {DialogShell} from '@/shared/components/dialog-shell'
import {EffortTag} from '@/shared/components/effort-tag'
import {cn} from '@/shared/lib/utils'
import {pluralize} from '@/shared/lib/format'
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
            title="Export roles"
            description="They go into one .json file you choose the place for."
            aside={
                <Button variant="ghost" size="sm" onClick={everyone ? none : all}>
                    {everyone ? 'Clear' : 'Select all'}
                </Button>
            }
            footer={
                <>
                    <p className="min-w-0 flex-1 text-sm text-muted-foreground">
                        {pickedIds.length === 0
                            ? 'Tick the ones to take with you.'
                            : `${pickedIds.length} of ${roles.length} selected`}
                    </p>
                    <Button variant="ghost" size="sm" onClick={onClose}>
                        Cancel
                    </Button>
                    <Button size="sm" disabled={pickedIds.length === 0} onClick={send}>
                        Export {pluralize(pickedIds.length, 'role')}
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
                    {role.description || 'No description set.'}
                </span>
            </span>
        </button>
    )
}
