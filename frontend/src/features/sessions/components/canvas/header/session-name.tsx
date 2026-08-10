import {useState, type ChangeEvent, type KeyboardEvent} from 'react'
import {Lock} from 'lucide-react'

import {Tooltip, TooltipContent, TooltipTrigger} from '@/shared/ui/tooltip'
import {FINALIZED_HINT} from '@/features/sessions/session-copy'
import type {Session} from '@/features/sessions/types'

const NAME_WIDTH = 'w-full max-w-96 min-w-0 lg:min-w-40'

export function SessionName({
    session,
    onRename,
}: {
    session: Session
    onRename: (name: string) => void
}) {
    if (session.finalized) return <FinalizedName name={session.name} />

    return <SessionNameInput key={session.id} name={session.name} onRename={onRename} />
}

function FinalizedName({name}: {name: string}) {
    return (
        <span className={`${NAME_WIDTH} flex items-center gap-1`}>
            <span className="truncate text-xl font-bold text-foreground">{name}</span>
            <Tooltip>
                <TooltipTrigger asChild>
                    <button
                        type="button"
                        aria-label="This session is finalized"
                        className="flex size-7 shrink-0 items-center justify-center rounded-md text-muted-foreground outline-none focus-visible:ring-2 focus-visible:ring-live"
                    >
                        <Lock className="size-3.5" />
                    </button>
                </TooltipTrigger>
                <TooltipContent side="bottom">{FINALIZED_HINT}</TooltipContent>
            </Tooltip>
        </span>
    )
}

function SessionNameInput({name, onRename}: {name: string; onRename: (name: string) => void}) {
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
            aria-label="Session name"
            onChange={change}
            onBlur={commit}
            onKeyDown={handleKeys}
            className={`${NAME_WIDTH} -ml-2 truncate rounded-md bg-transparent px-2 py-1 text-xl font-bold text-foreground outline-none transition-colors hover:bg-muted focus:bg-background focus-visible:ring-2 focus-visible:ring-live`}
        />
    )
}
