import {DetailSection, DetailText} from '@/features/store/components/detail-section'
import {EffortTag} from '@/shared/components/effort-tag'
import {StatusChip} from '@/shared/components/status-chip'
import {INPUT_TYPE_LABELS} from '@/shared/lib/enums'
import {t} from '@/shared/lib/i18n'
import type {Role, RoleInput} from '@/features/roles/types'

export function RoleInputList({inputs}: {inputs: RoleInput[]}) {
    if (inputs.length === 0)
        return <p className="text-sm text-muted-foreground">{t('store.preview.noInputs')}</p>

    return (
        <ul className="flex flex-col gap-1.5">
            {inputs.map((input) => (
                <li
                    key={input.key}
                    className="flex flex-col gap-1 rounded-lg border border-border bg-card px-3 py-2"
                >
                    <span className="flex items-center gap-2">
                        <span className="min-w-0 truncate font-mono text-sm">
                            {input.key}
                            {input.required && <span className="text-destructive">*</span>}
                        </span>
                        <span className="flex-1" />
                        <StatusChip tone="outline">{t(INPUT_TYPE_LABELS[input.type])}</StatusChip>
                    </span>
                    <span className="text-sm text-muted-foreground">{input.label}</span>
                    {input.default && (
                        <span className="font-mono text-xs text-muted-foreground">
                            {t('store.preview.default', {value: input.default})}
                        </span>
                    )}
                </li>
            ))}
        </ul>
    )
}

export function RoleDetail({role}: {role: Role}) {
    const structure = role.outputStructure.trim()

    return (
        <div className="grid grid-cols-1 gap-x-6 gap-y-6 p-4 lg:grid-cols-[minmax(0,1fr)_minmax(0,1.15fr)]">
            <div className="flex flex-col gap-6">
                <p className="text-base leading-[1.7] text-muted-foreground">{role.description}</p>

                <div className="flex flex-wrap items-center gap-1.5">
                    <EffortTag effort={role.effort} />
                    {role.pauseForReview && (
                        <StatusChip tone="attention">{t('store.preview.pauses')}</StatusChip>
                    )}
                    {!role.retryable && (
                        <StatusChip tone="muted">{t('store.preview.noRetry')}</StatusChip>
                    )}
                </div>

                <DetailSection label={t('store.preview.inputs')} term="input">
                    <RoleInputList inputs={role.inputs} />
                </DetailSection>

                <DetailSection label={t('store.preview.report')} term="reportFormat">
                    {structure ? (
                        <DetailText>{structure}</DetailText>
                    ) : (
                        <p className="text-sm text-muted-foreground">{t('store.card.freeform')}</p>
                    )}
                </DetailSection>
            </div>

            <div className="flex flex-col gap-6">
                <DetailSection label={t('store.preview.instructions')} term="instructions">
                    {role.instructions.map((instruction) => (
                        <DetailText key={instruction.key}>{instruction.value}</DetailText>
                    ))}
                </DetailSection>
            </div>
        </div>
    )
}
