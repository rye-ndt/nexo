import {Button} from '@/shared/ui/button'
import {DetailSection, DetailText} from '@/features/store/components/detail-section'
import {DialogShell} from '@/shared/components/dialog-shell'
import {EffortTag} from '@/shared/components/effort-tag'
import {HelpTip} from '@/shared/components/help-tip'
import {RoleInputList} from '@/features/store/components/role-detail'
import {t} from '@/shared/lib/i18n'
import type {Role} from '@/features/roles/types'
import type {StoreStep} from '@/features/store/types'

export function StepPreviewDialog({
    step,
    role,
    waitsFor,
    onClose,
}: {
    step: StoreStep
    role: Role
    /** The titles of the steps this one starts from, empty when it starts the graph. */
    waitsFor: string[]
    onClose: () => void
}) {
    return (
        <DialogShell
            onClose={onClose}
            title={step.title}
            term="step"
            footer={
                <>
                    <span className="flex-1" />
                    <Button variant="outline" size="sm" onClick={onClose}>
                        {t('store.preview.close')}
                    </Button>
                </>
            }
        >
            <div className="flex flex-col gap-2 border-b border-border px-4 py-3">
                <div className="flex items-center justify-between gap-3">
                    <span className="flex items-center gap-2">
                        <span className="micro-label">{t('store.preview.role')}</span>
                        <HelpTip term="role" />
                    </span>
                    <span className="flex min-w-0 items-center gap-2">
                        <span className="truncate font-mono text-base">{role.name}</span>
                        <EffortTag effort={role.effort} />
                    </span>
                </div>
                <p className="text-sm text-muted-foreground">
                    {waitsFor.length === 0
                        ? t('store.preview.first')
                        : t('store.preview.waitsFor', {names: waitsFor.join(', ')})}
                </p>
            </div>

            <div className="flex flex-col gap-6 p-4">
                <DetailSection label={t('store.preview.prompt')} term="prompt">
                    <DetailText>{step.prompt}</DetailText>
                </DetailSection>

                <DetailSection label={t('store.preview.inputs')} term="input">
                    <RoleInputList inputs={role.inputs} />
                </DetailSection>
            </div>
        </DialogShell>
    )
}
