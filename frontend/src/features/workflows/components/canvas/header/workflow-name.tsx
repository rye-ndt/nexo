import {useState, type ChangeEvent, type KeyboardEvent} from 'react'
import {Lock} from 'lucide-react'

import {Tooltip, TooltipContent, TooltipTrigger} from '@/shared/ui/tooltip'
import {lockedHint} from '@/features/workflows/workflow-copy'
import {isLocked} from '@/features/workflows/graph'
import {t} from '@/shared/lib/i18n'
import type {Workflow} from '@/features/workflows/types'

const NAME_WIDTH = 'w-full max-w-96 min-w-0 lg:min-w-40'

export function WorkflowName({
    workflow,
    onRename,
}: {
    workflow: Workflow
    onRename: (name: string) => void
}) {
    if (isLocked(workflow)) return <LockedName name={workflow.name} />

    return <WorkflowNameInput key={workflow.id} name={workflow.name} onRename={onRename} />
}

function LockedName({name}: {name: string}) {
    return (
        <span className={`${NAME_WIDTH} flex items-center gap-1`}>
            <span className="truncate text-xl font-bold text-foreground">{name}</span>
            <Tooltip>
                <TooltipTrigger asChild>
                    <button
                        type="button"
                        aria-label={t('canvas.name.locked')}
                        className="flex size-7 shrink-0 items-center justify-center rounded-md text-muted-foreground outline-none focus-visible:ring-2 focus-visible:ring-live"
                    >
                        <Lock className="size-3.5" />
                    </button>
                </TooltipTrigger>
                <TooltipContent side="bottom">{lockedHint()}</TooltipContent>
            </Tooltip>
        </span>
    )
}

function WorkflowNameInput({name, onRename}: {name: string; onRename: (name: string) => void}) {
    const [draft, setDraft] = useState(name)

    const change = (event: ChangeEvent<HTMLInputElement>) => setDraft(event.target.value)

    const commit = () => {
        const next = draft.trim()
        if (!next) {
            setDraft(name)
            return
        }

        if (next !== name) onRename(next)
    }

    const handleKeys = (event: KeyboardEvent<HTMLInputElement>) => {
        if (event.key === 'Enter') event.currentTarget.blur()
        if (event.key === 'Escape') {
            setDraft(name)
            event.currentTarget.blur()
        }
    }

    return (
        <input
            value={draft}
            aria-label={t('canvas.name.label')}
            autoCapitalize="off"
            autoCorrect="off"
            onChange={change}
            onBlur={commit}
            onKeyDown={handleKeys}
            className={`${NAME_WIDTH} -ml-2 truncate rounded-md bg-transparent px-2 py-1 text-xl font-bold text-foreground outline-none transition-colors hover:bg-muted focus:bg-background focus-visible:ring-2 focus-visible:ring-live`}
        />
    )
}
