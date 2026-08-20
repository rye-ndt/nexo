import type {ReactNode} from 'react'

import {HelpTip} from '@/shared/components/help-tip'
import type {GlossaryTerm} from '@/shared/lib/glossary'
import {cn} from '@/shared/lib/utils'
import {Dialog, DialogContent, DialogDescription, DialogTitle} from '@/shared/ui/dialog'

const SIZES = {
    default: 'h-[600px] max-w-[560px]',
    wide: 'h-[min(90vh,880px)] max-w-[min(94vw,1060px)]',
}

/** Always mounted open — callers render it conditionally — so there is no `open` prop. */
export function DialogShell({
    title,
    description,
    term,
    aside,
    footer,
    size = 'default',
    onClose,
    children,
}: {
    title: string
    description?: string
    term?: GlossaryTerm
    aside?: ReactNode
    footer: ReactNode
    size?: keyof typeof SIZES
    onClose: () => void
    children: ReactNode
}) {
    const change = (open: boolean) => {
        if (!open) onClose()
    }

    return (
        <Dialog open onOpenChange={change}>
            <DialogContent
                {...(description ? {} : {'aria-describedby': undefined})}
                className={cn('flex flex-col gap-0 overflow-hidden p-0', SIZES[size])}
            >
                <div className="flex h-16 shrink-0 items-center gap-3 border-b border-border px-4 pr-12">
                    <div className="flex min-w-0 flex-1 flex-col gap-1">
                        <div className="flex min-w-0 items-center gap-2">
                            <DialogTitle className="truncate">{title}</DialogTitle>
                            {term && <HelpTip term={term} side="bottom" />}
                        </div>
                        {description && (
                            <DialogDescription className="truncate">
                                {description}
                            </DialogDescription>
                        )}
                    </div>
                    {aside}
                </div>

                <div className="min-h-0 flex-1 scroll-py-6 overflow-y-auto">{children}</div>

                <div className="flex h-14 shrink-0 items-center gap-2 border-t border-border px-4">
                    {footer}
                </div>
            </DialogContent>
        </Dialog>
    )
}
