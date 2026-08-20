import {Button} from '@/shared/ui/button'
import {HelpTip} from '@/shared/components/help-tip'
import {Tooltip, TooltipContent, TooltipTrigger} from '@/shared/ui/tooltip'
import {ActionEmphasis, type WorkflowAction} from '@/features/workflows/workflow-actions'

const VARIANTS: Record<ActionEmphasis, 'default' | 'outline' | 'ghost'> = {
    [ActionEmphasis.Primary]: 'default',
    [ActionEmphasis.Outline]: 'outline',
    [ActionEmphasis.Ghost]: 'ghost',
}

export function HeaderAction({action, onClick}: {action: WorkflowAction; onClick: () => void}) {
    const {label, icon: Icon, emphasis, term} = action
    const alwaysLabelled = emphasis === ActionEmphasis.Primary

    return (
        <span className="flex items-center gap-1">
            <Tooltip>
                <TooltipTrigger asChild>
                    <Button
                        variant={VARIANTS[emphasis]}
                        size="sm"
                        aria-label={label}
                        onClick={onClick}
                    >
                        <Icon />
                        <span className={alwaysLabelled ? 'inline' : 'hidden xl:inline'}>
                            {label}
                        </span>
                    </Button>
                </TooltipTrigger>
                <TooltipContent side="bottom" className={alwaysLabelled ? 'hidden' : 'xl:hidden'}>
                    {label}
                </TooltipContent>
            </Tooltip>

            {term && <HelpTip term={term} side="bottom" />}
        </span>
    )
}
