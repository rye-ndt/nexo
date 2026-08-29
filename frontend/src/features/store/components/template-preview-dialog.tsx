import {useState} from 'react'
import {ChevronRight} from 'lucide-react'

import {Button} from '@/shared/ui/button'
import {DialogShell} from '@/shared/components/dialog-shell'
import {EffortTag} from '@/shared/components/effort-tag'
import {StatusChip} from '@/shared/components/status-chip'
import {StepPreviewDialog} from '@/features/store/components/step-preview-dialog'
import {TemplateGraph} from '@/features/store/components/template-graph'
import {WorkflowLocationsFields} from '@/features/workflows/components/workflow-locations'
import {t} from '@/shared/lib/i18n'
import type {Role} from '@/features/roles/types'
import type {StoreStep, StoreTemplate} from '@/features/store/types'
import type {WorkflowLocations} from '@/features/workflows/types'

export function TemplatePreviewDialog({
    template,
    roles,
    held,
    busy,
    onAdd,
    onPreviewRole,
    onClose,
}: {
    template: StoreTemplate
    roles: Role[]
    held: (roleId: string) => boolean
    busy: boolean
    onAdd: (locations: WorkflowLocations) => void
    onPreviewRole: (role: Role) => void
    onClose: () => void
}) {
    const [projectDir, setProjectDir] = useState('')
    const [reading, setReading] = useState<StoreStep | null>(null)

    const locations = {projectDir: projectDir.trim()}
    const located = locations.projectDir.length > 0

    const roleOf = (step: StoreStep) => roles.find((role) => role.id === step.roleId)
    const titleOf = (stepId: string) =>
        template.steps.find((step) => step.id === stepId)?.title ?? stepId

    const readingRole = reading && roleOf(reading)

    const confirm = () => {
        if (located && !busy) onAdd(locations)
    }

    return (
        <DialogShell
            onClose={onClose}
            title={template.name}
            description={t('store.add.description')}
            term="workflow"
            footer={
                <>
                    <span className="flex-1" />
                    <Button variant="outline" size="sm" onClick={onClose}>
                        {t('store.add.cancel')}
                    </Button>
                    <Button size="sm" disabled={!located || busy} onClick={confirm}>
                        {busy ? t('store.card.adding') : t('store.add.confirm')}
                    </Button>
                </>
            }
        >
            <div className="flex h-20 shrink-0 items-center justify-center border-b border-border bg-accent px-4">
                <TemplateGraph steps={template.steps} />
            </div>

            <section className="flex flex-col gap-3 border-b border-border px-4 py-4">
                <span className="micro-label">{t('store.add.steps')}</span>

                <p className="text-sm text-muted-foreground">{template.description}</p>

                <div className="flex flex-col gap-1">
                    {template.steps.map((step) => (
                        <StepRow
                            key={step.id}
                            step={step}
                            role={roleOf(step)}
                            onRead={setReading}
                        />
                    ))}
                </div>
            </section>

            <section className="flex flex-col gap-3 border-b border-border px-4 py-4">
                <span className="micro-label">{t('store.card.roles')}</span>

                <div className="flex flex-wrap gap-1.5">
                    {roles.map((role) => (
                        <button
                            key={role.id}
                            type="button"
                            aria-label={t('store.card.openRole', {name: role.name})}
                            onClick={() => onPreviewRole(role)}
                            className="rounded-sm outline-none focus-visible:ring-2 focus-visible:ring-ring/50"
                        >
                            <StatusChip
                                tone={held(role.id) ? 'muted' : 'outline'}
                                className="transition-colors duration-[120ms] hover:text-foreground"
                            >
                                {held(role.id)
                                    ? t('store.card.roleOwned', {name: role.name})
                                    : role.name}
                            </StatusChip>
                        </button>
                    ))}
                </div>

                <p className="text-sm text-muted-foreground">{t('store.add.rolesToo')}</p>
            </section>

            <WorkflowLocationsFields projectDir={projectDir} onProjectDirChange={setProjectDir} />

            {reading && readingRole && (
                <StepPreviewDialog
                    key={reading.id}
                    step={reading}
                    role={readingRole}
                    waitsFor={reading.dependsOn.map(titleOf)}
                    onClose={() => setReading(null)}
                />
            )}
        </DialogShell>
    )
}

function StepRow({
    step,
    role,
    onRead,
}: {
    step: StoreStep
    role?: Role
    onRead: (step: StoreStep) => void
}) {
    return (
        <button
            type="button"
            aria-label={t('store.card.openStep', {name: step.title})}
            onClick={() => onRead(step)}
            className="flex items-center gap-3 rounded-lg border border-border bg-card px-3 py-2 text-left outline-none transition-colors duration-[120ms] hover:border-border-strong hover:bg-muted/40 focus-visible:ring-2 focus-visible:ring-ring/50"
        >
            <span className="min-w-0 flex-1 truncate text-base">{step.title}</span>
            {role && <span className="shrink-0 text-sm text-muted-foreground">{role.name}</span>}
            {role && <EffortTag effort={role.effort} />}
            <ChevronRight aria-hidden="true" className="size-3.5 shrink-0 text-muted-foreground" />
        </button>
    )
}
