import {Check, Plus} from 'lucide-react'

import {Button} from '@/shared/ui/button'
import {t} from '@/shared/lib/i18n'
import type {Role} from '@/features/roles/types'

export function AddRoleButton({
    role,
    held,
    busy,
    onAdd,
    onAlreadyHeld,
}: {
    role: Role
    held: boolean
    /** The library is mid-write, so `held` is not yet an answer worth acting on. */
    busy: boolean
    onAdd: (role: Role) => void
    onAlreadyHeld: (role: Role) => void
}) {
    if (held)
        return (
            <Button
                variant="outline"
                size="sm"
                disabled={busy}
                aria-label={t('store.already.title')}
                onClick={() => onAlreadyHeld(role)}
            >
                <Check />
                {t('store.card.added')}
            </Button>
        )

    return (
        <Button
            size="sm"
            disabled={busy}
            aria-label={t('store.card.addRole', {name: role.name})}
            onClick={() => onAdd(role)}
        >
            <Plus />
            {t('store.card.add')}
        </Button>
    )
}
