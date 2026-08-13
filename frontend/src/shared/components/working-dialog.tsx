import {Loader2} from 'lucide-react'

import {Dialog, DialogContent, DialogDescription, DialogTitle} from '@/shared/ui/dialog'

export function WorkingDialog({title, description}: {title: string; description: string}) {
    const stay = (event: Event) => event.preventDefault()

    return (
        <Dialog open>
            <DialogContent
                showCloseButton={false}
                className="p-5"
                onEscapeKeyDown={stay}
                onInteractOutside={stay}
            >
                <div className="flex items-center gap-3">
                    <Loader2 className="size-4 shrink-0 animate-spin text-live" />

                    <div className="flex min-w-0 flex-col gap-1">
                        <DialogTitle className="text-base">{title}</DialogTitle>
                        <DialogDescription>{description}</DialogDescription>
                    </div>
                </div>
            </DialogContent>
        </Dialog>
    )
}
