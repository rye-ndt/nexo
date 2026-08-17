import {Gauge} from 'lucide-react'

import {Button} from '@/shared/ui/button'
import {RunCostDialog} from '@/features/sessions/components/canvas/header/run-cost-dialog'
import {Tooltip, TooltipContent, TooltipTrigger} from '@/shared/ui/tooltip'
import {useToggle} from '@/shared/hooks/use-toggle'
import type {Session} from '@/features/sessions/types'

/** What the run has cost so far, one click away instead of crowding the bar. */
export function RunCost({session}: {session: Session}) {
    const dialog = useToggle()

    if (!session.started) return null

    return (
        <>
            <Tooltip>
                <TooltipTrigger asChild>
                    <Button variant="ghost" size="icon" aria-label="Run cost" onClick={dialog.open}>
                        <Gauge />
                    </Button>
                </TooltipTrigger>
                <TooltipContent side="bottom">Run cost</TooltipContent>
            </Tooltip>

            {dialog.on && <RunCostDialog session={session} onClose={dialog.close} />}
        </>
    )
}
