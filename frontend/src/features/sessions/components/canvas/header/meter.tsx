import {Tooltip, TooltipContent, TooltipTrigger} from '@/shared/ui/tooltip'

/** One reading in the run cluster: what it measures, then the number. */
export function Meter({label, value, hint}: {label: string; value: string; hint: string}) {
    return (
        <Tooltip>
            <TooltipTrigger asChild>
                <span className="flex h-full items-center gap-2 px-2.5">
                    <span className="micro-label">{label}</span>
                    <span className="font-mono text-sm font-medium tabular-nums text-foreground">
                        {value}
                    </span>
                </span>
            </TooltipTrigger>
            <TooltipContent side="bottom">{hint}</TooltipContent>
        </Tooltip>
    )
}
