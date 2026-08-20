import {Button} from '@/shared/ui/button'
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogTitle,
} from '@/shared/ui/dialog'
import {t} from '@/shared/lib/i18n'

export function NoticeDialog({
    title,
    description,
    detail,
    onClose,
}: {
    title: string
    description: string
    detail: string
    onClose: () => void
}) {
    const close = (open: boolean) => {
        if (!open) onClose()
    }

    return (
        <Dialog open onOpenChange={close}>
            <DialogContent showCloseButton={false} className="gap-0 overflow-hidden p-0">
                <span aria-hidden="true" className="block h-0.5 bg-state-done" />

                <div className="flex flex-col gap-2 px-5 pt-4 pb-5">
                    <DialogTitle className="break-words">{title}</DialogTitle>

                    <DialogDescription className="text-base break-words text-foreground">
                        {description}
                    </DialogDescription>

                    <p className="truncate font-mono text-sm text-muted-foreground">{detail}</p>
                </div>

                <DialogFooter className="h-14 border-t border-border px-5">
                    <Button autoFocus size="sm" onClick={onClose}>
                        {t('shared.notice.done')}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    )
}
