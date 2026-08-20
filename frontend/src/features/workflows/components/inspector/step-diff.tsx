import {FileChanges} from '@/features/workflows/components/inspector/file-changes'
import {useStepDiff} from '@/features/workflows/use-step-diff'
import {Button} from '@/shared/ui/button'
import {t} from '@/shared/lib/i18n'
import type {ReactNode} from 'react'

/** Same row the collapsed section draws when it has nothing, so the header never jumps. */
function DiffNote({text, action}: {text: string; action?: ReactNode}) {
    return (
        <div className="flex items-center gap-2 py-1">
            <span className="size-3.5 shrink-0" />
            <span className="micro-label">{t('inspector.files.label')}</span>
            <span className="min-w-0 flex-1 truncate text-sm text-muted-foreground">{text}</span>
            {action}
        </div>
    )
}

export function StepDiff({workflowId, stepId}: {workflowId: string; stepId: string}) {
    const {changes, loading, failed, retry} = useStepDiff(workflowId, stepId)

    if (loading) return <DiffNote text={t('inspector.diff.loading')} />

    if (failed)
        return (
            <DiffNote
                text={t('inspector.diff.failed')}
                action={
                    <Button variant="ghost" size="xs" onClick={retry}>
                        {t('inspector.diff.retry')}
                    </Button>
                }
            />
        )

    return <FileChanges changes={changes} />
}
