import {useState} from 'react'
import {Pencil, Plus, Trash2} from 'lucide-react'

import {Button} from '@/shared/ui/button'
import {ConfirmDialog} from '@/shared/components/confirm-dialog'
import {StatusChip} from '@/shared/components/status-chip'
import {EffortTag} from '@/shared/components/effort-tag'
import {t} from '@/shared/lib/i18n'
import type {Role} from '@/features/roles/types'

export function RoleList({
    roles,
    loading,
    onPick,
    onEdit,
    onRemove,
    onCreate,
}: {
    roles: Role[]
    loading: boolean
    onPick: (role: Role) => void
    onEdit: (role: Role) => void
    onRemove: (roleId: string) => void
    onCreate: () => void
}) {
    if (loading)
        return <p className="px-4 py-3 text-base text-muted-foreground">{t('role.list.loading')}</p>

    if (roles.length === 0)
        return (
            <div className="flex flex-col items-center gap-3 px-6 py-12 text-center">
                <p className="text-base text-muted-foreground">{t('role.list.empty')}</p>
                <Button variant="outline" size="sm" onClick={onCreate}>
                    <Plus />
                    {t('role.action.new')}
                </Button>
            </div>
        )

    return (
        <div className="flex flex-col gap-2 p-4">
            {roles.map((role) => (
                <RoleCard
                    key={role.id}
                    role={role}
                    onPick={onPick}
                    onEdit={onEdit}
                    onRemove={onRemove}
                />
            ))}
        </div>
    )
}

function RoleCard({
    role,
    onPick,
    onEdit,
    onRemove,
}: {
    role: Role
    onPick: (role: Role) => void
    onEdit: (role: Role) => void
    onRemove: (roleId: string) => void
}) {
    const [confirming, setConfirming] = useState(false)

    const pick = () => onPick(role)
    const edit = () => onEdit(role)
    const askRemove = () => setConfirming(true)
    const cancelRemove = () => setConfirming(false)
    const remove = () => onRemove(role.id)

    return (
        <div className="group relative rounded-xl border border-border bg-card transition-colors hover:border-foreground/25 hover:bg-muted/40">
            <button
                type="button"
                onClick={pick}
                className="flex w-full flex-col gap-2 rounded-xl p-3 text-left outline-none focus-visible:ring-2 focus-visible:ring-ring"
            >
                <span className="flex items-center gap-2 pr-16">
                    <span className="truncate text-base font-medium">{role.name}</span>
                    <EffortTag effort={role.effort} />
                    {role.outputStructure.trim() && (
                        <StatusChip tone="info">{t('role.card.structured')}</StatusChip>
                    )}
                    {!role.retryable && (
                        <span className="shrink-0 text-xs text-muted-foreground">
                            {t('role.card.noRetry')}
                        </span>
                    )}
                </span>

                <span className="line-clamp-2 text-sm text-muted-foreground">
                    {role.description || t('role.card.noDescription')}
                </span>

                {role.inputs.length === 0 ? (
                    <span className="text-sm text-muted-foreground">{t('role.card.noInputs')}</span>
                ) : (
                    <span className="flex flex-wrap gap-1">
                        {role.inputs.map((input) => (
                            <span
                                key={input.key}
                                className="rounded-sm bg-muted px-1.5 py-0.5 font-mono text-xs break-all text-muted-foreground"
                            >
                                {input.key}
                                {input.required && <span className="text-destructive">*</span>}
                            </span>
                        ))}
                    </span>
                )}
            </button>

            <span className="absolute top-2 right-2 flex items-center gap-1 opacity-0 transition-opacity group-focus-within:opacity-100 group-hover:opacity-100">
                <Button
                    variant="ghost"
                    size="icon-sm"
                    aria-label={t('role.card.edit', {name: role.name})}
                    onClick={edit}
                >
                    <Pencil />
                </Button>
                <Button
                    variant="ghost"
                    size="icon-sm"
                    aria-label={t('role.card.remove', {name: role.name})}
                    className="text-muted-foreground hover:bg-destructive/10 hover:text-destructive"
                    onClick={askRemove}
                >
                    <Trash2 />
                </Button>
            </span>

            {confirming && (
                <ConfirmDialog
                    title={t('role.card.removeTitle', {name: role.name})}
                    description={t('role.card.removeBody')}
                    confirmLabel={t('role.card.removeConfirm')}
                    destructive
                    onConfirm={remove}
                    onClose={cancelRemove}
                />
            )}
        </div>
    )
}
