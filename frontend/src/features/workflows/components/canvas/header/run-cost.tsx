import {Gauge} from 'lucide-react'

import {Button} from '@/shared/ui/button'
import {RunCostDialog} from '@/features/workflows/components/canvas/header/run-cost-dialog'
import {Tooltip, TooltipContent, TooltipTrigger} from '@/shared/ui/tooltip'
import {useToggle} from '@/shared/hooks/use-toggle'
import type {Workflow} from '@/features/workflows/types'

/** What the run has cost so far, one click away instead of crowding the bar. */
export function RunCost({workflow}: {workflow: Workflow}) {
    const dialog = useToggle()

    if (!workflow.started) return null

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

            {dialog.on && <RunCostDialog workflow={workflow} onClose={dialog.close} />}
        </>
    )
}
