import type {ReactNode} from 'react'

import {HelpTip} from '@/shared/components/help-tip'
import type {GlossaryTerm} from '@/shared/lib/glossary'
import {cn} from '@/shared/lib/utils'

export function Field({
    htmlFor,
    label,
    hint,
    term,
    className,
    children,
}: {
    htmlFor: string
    label: string
    hint?: string
    term?: GlossaryTerm
    className?: string
    children: ReactNode
}) {
    return (
        <div className={cn('flex flex-col gap-2', className)}>
            <div className="flex items-center gap-2">
                <label htmlFor={htmlFor} className="text-base font-medium">
                    {label}
                </label>
                {term && <HelpTip term={term} />}
            </div>
            {children}
            {hint && <p className="text-sm text-muted-foreground">{hint}</p>}
        </div>
    )
}
