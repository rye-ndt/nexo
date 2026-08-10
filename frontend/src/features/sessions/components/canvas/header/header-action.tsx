import {Button} from '@/shared/ui/button'
import {Tooltip, TooltipContent, TooltipTrigger} from '@/shared/ui/tooltip'
import {ActionEmphasis, type SessionAction} from '@/features/sessions/session-actions'

const VARIANTS: Record<ActionEmphasis, 'default' | 'outline' | 'ghost'> = {
    [ActionEmphasis.Primary]: 'default',
    [ActionEmphasis.Outline]: 'outline',
    [ActionEmphasis.Ghost]: 'ghost',
}

export function HeaderAction({action, onClick}: {action: SessionAction; onClick: () => void}) {
    const {label, icon: Icon, emphasis} = action
    const alwaysLabelled = emphasis === ActionEmphasis.Primary

    return (
        <Tooltip>
            <TooltipTrigger asChild>
                <Button variant={VARIANTS[emphasis]} size="sm" aria-label={label} onClick={onClick}>
                    <Icon />
                    <span className={alwaysLabelled ? 'inline' : 'hidden xl:inline'}>{label}</span>
                </Button>
            </TooltipTrigger>
            <TooltipContent side="bottom" className={alwaysLabelled ? 'hidden' : 'xl:hidden'}>
                {label}
            </TooltipContent>
        </Tooltip>
    )
}
