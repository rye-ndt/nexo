import {PanelLeftClose, PanelLeftOpen} from 'lucide-react'

import {Button} from '@/shared/ui/button'
import {Tooltip, TooltipContent, TooltipTrigger} from '@/shared/ui/tooltip'
import {t} from '@/shared/lib/i18n'

export function RailToggle({open, onToggle}: {open: boolean; onToggle: () => void}) {
    const label = open ? t('canvas.rail.hide') : t('canvas.rail.show')

    return (
        <Tooltip>
            <TooltipTrigger asChild>
                <Button
                    variant="ghost"
                    size="icon"
                    className="-ml-1 shrink-0"
                    aria-label={label}
                    onClick={onToggle}
                >
                    {open ? <PanelLeftClose /> : <PanelLeftOpen />}
                </Button>
            </TooltipTrigger>
            <TooltipContent side="bottom">{label}</TooltipContent>
        </Tooltip>
    )
}
