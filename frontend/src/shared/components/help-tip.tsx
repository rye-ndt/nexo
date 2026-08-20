import {CircleHelp} from 'lucide-react'

import {GLOSSARY, type GlossaryTerm} from '@/shared/lib/glossary'
import {t} from '@/shared/lib/i18n'
import {cn} from '@/shared/lib/utils'
import {Tooltip, TooltipContent, TooltipTrigger} from '@/shared/ui/tooltip'

export function HelpTip({
    term,
    side = 'top',
    className,
}: {
    term: GlossaryTerm
    side?: 'top' | 'right' | 'bottom' | 'left'
    className?: string
}) {
    const entry = GLOSSARY[term]

    return (
        <Tooltip>
            <TooltipTrigger asChild>
                <button
                    type="button"
                    aria-label={t('shared.helpTip.aria', {term: t(entry.title).toLowerCase()})}
                    onClick={(event) => event.preventDefault()}
                    className={cn(
                        '-m-1 inline-flex shrink-0 rounded-full p-1 text-muted-foreground/60 outline-none transition-colors hover:text-live focus-visible:text-live focus-visible:ring-2 focus-visible:ring-ring/50',
                        className,
                    )}
                >
                    <CircleHelp className="size-3.5" />
                </button>
            </TooltipTrigger>

            <TooltipContent side={side} className="max-w-[19rem] flex-col items-start gap-1">
                <span className="font-medium">{t(entry.what)}</span>
                <span className="text-background/70">{t(entry.why)}</span>
            </TooltipContent>
        </Tooltip>
    )
}
