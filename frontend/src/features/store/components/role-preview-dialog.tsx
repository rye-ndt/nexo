import {AddRoleButton} from '@/features/store/components/add-role-button'
import {Button} from '@/shared/ui/button'
import {DialogShell} from '@/shared/components/dialog-shell'
import {RoleDetail} from '@/features/store/components/role-detail'
import {t} from '@/shared/lib/i18n'
import type {Role} from '@/features/roles/types'

/**
 * Everything a role carries, laid out the way the role form lays it out, with the
 * form's controls taken away. Adding leaves it open: the button flips to Added in
 * place, so reading on is never interrupted by a dialog closing under you.
 */
export function RolePreviewDialog({
    role,
    held,
    busy,
    onAdd,
    onAlreadyHeld,
    onClose,
}: {
    role: Role
    held: boolean
    busy: boolean
    onAdd: (role: Role) => void
    onAlreadyHeld: (role: Role) => void
    onClose: () => void
}) {
    return (
        <DialogShell
            onClose={onClose}
            size="wide"
            title={role.name}
            term="role"
            footer={
                <>
                    <span className="flex-1" />
                    <Button variant="outline" size="sm" onClick={onClose}>
                        {t('store.preview.close')}
                    </Button>
                    <AddRoleButton
                        role={role}
                        held={held}
                        busy={busy}
                        onAdd={onAdd}
                        onAlreadyHeld={onAlreadyHeld}
                    />
                </>
            }
        >
            <RoleDetail role={role} />
        </DialogShell>
    )
}
