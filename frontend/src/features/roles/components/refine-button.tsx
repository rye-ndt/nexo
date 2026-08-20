import {Loader2, Sparkle} from 'lucide-react'

import {Button} from '@/shared/ui/button'
import {t} from '@/shared/lib/i18n'
import {cn} from '@/shared/lib/utils'

/**
 * Hands the name and role to an agent and takes back a finished role. It keeps
 * the footer's button geometry and spends its colour on the one thing that is
 * different about it: something else is doing the work.
 */
export function RefineButton({
    disabled,
    reason,
    working,
    onFill,
}: {
    disabled: boolean
    /** Why it cannot run, shown in place of the hint when it cannot. */
    reason: string
    working: boolean
    onFill: () => void
}) {
    return (
        <Button
            size="sm"
            variant="outline"
            aria-label={reason || undefined}
            data-working={working}
            disabled={disabled || working}
            onClick={onFill}
            className={cn(
                'shrink-0',
                !disabled && 'sheen border-live/40 text-foreground hover:border-live/60',
            )}
        >
            {working ? (
                <Loader2 className="animate-spin text-live" />
            ) : (
                <Sparkle className={cn(!disabled && 'text-live')} />
            )}
            <span className="relative">
                {working ? t('role.helper.filling') : t('role.helper.fill')}
            </span>
        </Button>
    )
}
