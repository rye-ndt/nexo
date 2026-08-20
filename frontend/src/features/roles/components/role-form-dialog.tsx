import {useState} from 'react'

import {ConfirmDialog} from '@/shared/components/confirm-dialog'
import {DialogShell} from '@/shared/components/dialog-shell'
import {RefineButton} from '@/features/roles/components/refine-button'
import {RoleForm} from '@/features/roles/components/role-form'
import {TourBanner} from '@/features/onboarding/components/tour-banner'
import {TourStopId} from '@/features/onboarding/tour'
import {Button} from '@/shared/ui/button'
import {t} from '@/shared/lib/i18n'
import {useRoleHelper} from '@/features/roles/use-role-helper'
import {useRoles} from '@/features/roles/use-roles'
import {useTourStop} from '@/features/onboarding/tour-context'
import {emptyRole, roleIssues} from '@/features/roles/role'
import type {DraftContext, Role, RoleDraft} from '@/features/roles/types'

export function RoleFormDialog({
    role,
    context,
    onClose,
}: {
    role: Role | null
    context?: DraftContext
    onClose: () => void
}) {
    const {saveRole, saving} = useRoles()
    const guided = useTourStop(TourStopId.Role)
    const [draft, setDraft] = useState<RoleDraft>(() =>
        role ? structuredClone(role) : emptyRole(),
    )
    const [opened] = useState(() => JSON.stringify(draft))
    const [asking, setAsking] = useState<'save' | 'discard' | null>(null)
    /** What the form held before the agent filled it in, so the user can get it back. */
    const [beforeFill, setBeforeFill] = useState<RoleDraft | null>(null)

    const helper = useRoleHelper((filled, sent) => {
        setBeforeFill(sent)
        setDraft(filled)
    }, context)

    const issues = roleIssues(draft)

    const dismiss = () => setAsking(null)

    const undoFill = () => {
        if (beforeFill) setDraft(beforeFill)
        setBeforeFill(null)
    }

    const commit = () => {
        setAsking(null)
        saveRole(draft, {onSuccess: onClose})
    }

    const save = () => {
        if (issues.length > 0) return
        if (role) setAsking('save')
        else commit()
    }

    const requestClose = () => {
        if (guided) return
        if (JSON.stringify(draft) !== opened) setAsking('discard')
        else onClose()
    }

    const saveLabel = role ? t('role.form.save') : t('role.form.create')

    /** There is a name and a role, so the agent has something to work from. */
    const fillable = Boolean(draft.name.trim() && draft.description.trim())

    // A blocked helper takes the hint only once the user could plausibly reach for
    // it, since that is the moment they find the button dead and want to know why.
    const hint = helper.filling
        ? t('role.form.filling')
        : (fillable ? helper.blocked : '') || (issues[0] ?? '')

    return (
        <DialogShell
            onClose={requestClose}
            size="wide"
            title={role ? t('role.form.titleEdit') : t('role.form.titleNew')}
            banner={guided ? <TourBanner stop={TourStopId.Role} /> : undefined}
            footer={
                <>
                    <p className="min-w-0 flex-1 truncate text-sm text-muted-foreground">{hint}</p>
                    <RefineButton
                        disabled={!fillable || Boolean(helper.blocked)}
                        reason={helper.blocked}
                        working={helper.filling}
                        onFill={() => helper.fillIn(draft)}
                    />
                    {beforeFill && !helper.filling && (
                        <Button variant="ghost" size="sm" onClick={undoFill}>
                            {t('role.form.undo')}
                        </Button>
                    )}
                    <Button variant="ghost" size="sm" onClick={requestClose}>
                        {t('role.form.cancel')}
                    </Button>
                    <Button
                        size="sm"
                        disabled={saving || helper.filling || issues.length > 0}
                        onClick={save}
                    >
                        {saving ? t('role.form.saving') : saveLabel}
                    </Button>
                </>
            }
        >
            <RoleForm draft={draft} onChange={setDraft} />

            {asking === 'save' && role && (
                <ConfirmDialog
                    title={t('role.form.saveTitle', {name: role.name})}
                    description={t('role.form.saveBody')}
                    confirmLabel={t('role.form.save')}
                    busy={saving}
                    onConfirm={commit}
                    onClose={dismiss}
                />
            )}

            {asking === 'discard' && (
                <ConfirmDialog
                    destructive
                    title={
                        role
                            ? t('role.form.discardEditTitle', {name: role.name})
                            : t('role.form.discardNewTitle')
                    }
                    description={
                        role ? t('role.form.discardEditBody') : t('role.form.discardNewBody')
                    }
                    confirmLabel={t('role.form.discard')}
                    dismissLabel={t('role.form.keepEditing')}
                    onConfirm={onClose}
                    onClose={dismiss}
                />
            )}
        </DialogShell>
    )
}
