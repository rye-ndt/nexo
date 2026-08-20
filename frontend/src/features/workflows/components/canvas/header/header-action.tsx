import {Button} from '@/shared/ui/button'
import {HelpTip} from '@/shared/components/help-tip'
import {Tooltip, TooltipContent, TooltipTrigger} from '@/shared/ui/tooltip'
import {ActionEmphasis, type WorkflowAction} from '@/features/workflows/workflow-actions'
import {cn} from '@/shared/lib/utils'

const VARIANTS: Record<ActionEmphasis, 'default' | 'outline' | 'ghost'> = {
    [ActionEmphasis.Primary]: 'default',
    [ActionEmphasis.Outline]: 'outline',
    [ActionEmphasis.Ghost]: 'ghost',
}

export function HeaderAction({action, onClick}: {action: WorkflowAction; onClick: () => void}) {
    const {label, icon: Icon, emphasis, term, disabledReason} = action
    const alwaysLabelled = emphasis === ActionEmphasis.Primary

    return (
        <span className="flex items-center gap-1">
            <Tooltip>
                <TooltipTrigger asChild>
                    <Button
                        variant={VARIANTS[emphasis]}
                        size="sm"
                        aria-label={label}
                        aria-disabled={disabledReason ? true : undefined}
                        className={cn(disabledReason && 'opacity-50')}
                        onClick={disabledReason ? undefined : onClick}
                    >
                        <Icon />
                        <span className={alwaysLabelled ? 'inline' : 'hidden xl:inline'}>
                            {label}
                        </span>
                    </Button>
                </TooltipTrigger>
                <TooltipContent
                    side="bottom"
                    className={cn(!disabledReason && (alwaysLabelled ? 'hidden' : 'xl:hidden'))}
                >
                    {disabledReason ?? label}
                </TooltipContent>
            </Tooltip>

            {term && <HelpTip term={term} side="bottom" />}
        </span>
    )
}
