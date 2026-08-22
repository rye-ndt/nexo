import {useState} from 'react'

import {Button} from '@/shared/ui/button'
import {Checkbox} from '@/shared/ui/checkbox'
import {
    Dialog,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
} from '@/shared/ui/dialog'
import {t} from '@/shared/lib/i18n'
import type {Workflow} from '@/features/workflows/types'

export function DuplicateWorkflowDialog({
    workflow,
    onConfirm,
    onClose,
}: {
    workflow: Workflow
    onConfirm: (copyInputs: boolean) => void
    onClose: () => void
}) {
    const [copyInputs, setCopyInputs] = useState(true)

    const close = (open: boolean) => {
        if (!open) onClose()
    }

    return (
        <Dialog open onOpenChange={close}>
            <DialogContent showCloseButton={false} className="p-5">
                <DialogHeader>
                    <DialogTitle className="break-words">
                        {t('workflow.duplicate.title', {name: workflow.name})}
                    </DialogTitle>
                    <DialogDescription>{t('workflow.duplicate.description')}</DialogDescription>
                </DialogHeader>

                <div className="mt-5 flex items-center gap-2">
                    <Checkbox
                        id="duplicate-copy-inputs"
                        checked={copyInputs}
                        onCheckedChange={(checked) => setCopyInputs(checked === true)}
                    />
                    <label
                        htmlFor="duplicate-copy-inputs"
                        className="text-base text-foreground select-none"
                    >
                        {t('workflow.duplicate.copyInputs')}
                    </label>
                </div>

                <DialogFooter className="mt-5">
                    <Button variant="outline" size="sm" onClick={onClose}>
                        {t('shared.confirm.dismiss')}
                    </Button>
                    <Button autoFocus size="sm" onClick={() => onConfirm(copyInputs)}>
                        {t('workflow.action.duplicate')}
                    </Button>
                </DialogFooter>
            </DialogContent>
        </Dialog>
    )
}
