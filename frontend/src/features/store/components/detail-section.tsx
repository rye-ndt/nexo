import type {ReactNode} from 'react'

import {HelpTip} from '@/shared/components/help-tip'
import type {GlossaryTerm} from '@/shared/lib/glossary'
import {cn} from '@/shared/lib/utils'

export function DetailSection({
    label,
    term,
    className,
    children,
}: {
    label: string
    term?: GlossaryTerm
    className?: string
    children: ReactNode
}) {
    return (
        <section className={cn('flex flex-col gap-2', className)}>
            <span className="flex items-center gap-2">
                <span className="micro-label">{label}</span>
                {term && <HelpTip term={term} />}
            </span>
            {children}
        </section>
    )
}

export function DetailText({children}: {children: ReactNode}) {
    return (
        <p className="rounded-xl border border-border bg-muted/30 p-3 font-mono text-sm leading-[1.7] whitespace-pre-wrap">
            {children}
        </p>
    )
}
