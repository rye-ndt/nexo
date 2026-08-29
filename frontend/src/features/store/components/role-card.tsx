import {AddRoleButton} from '@/features/store/components/add-role-button'
import {EffortTag} from '@/shared/components/effort-tag'
import {t, tn} from '@/shared/lib/i18n'
import type {Role} from '@/features/roles/types'

const TOP_LEVEL_KEY = /^([A-Za-z_][A-Za-z0-9_]*)\s*:/
const KEYS_SHOWN = 3

/** The names a role's report comes back under, which is the most telling thing about it. */
function reportKeys(structure: string): string[] {
    return structure
        .split('\n')
        .filter((line) => line.length > 0 && line === line.trimStart())
        .map((line) => TOP_LEVEL_KEY.exec(line)?.[1])
        .filter((key): key is string => Boolean(key))
        .slice(0, KEYS_SHOWN)
}

export function RoleCard({
    role,
    held,
    busy,
    onOpen,
    onAdd,
    onAlreadyHeld,
}: {
    role: Role
    held: boolean
    /** The library is mid-write, so `held` is not yet an answer worth acting on. */
    busy: boolean
    onOpen: (role: Role) => void
    onAdd: (role: Role) => void
    onAlreadyHeld: (role: Role) => void
}) {
    const keys = reportKeys(role.outputStructure)

    return (
        <article className="group surface-card flex flex-col overflow-hidden ring-1 ring-border transition-shadow duration-[120ms] hover:ring-border-strong">
            <button
                type="button"
                aria-label={t('store.card.openRole', {name: role.name})}
                onClick={() => onOpen(role)}
                className="flex flex-1 flex-col text-left outline-none focus-visible:ring-2 focus-visible:ring-ring/50 focus-visible:ring-inset"
            >
                <span className="flex h-16 shrink-0 flex-col justify-center gap-0.5 overflow-hidden border-b border-border bg-accent px-4">
                    {keys.length > 0 ? (
                        keys.map((key) => (
                            <span
                                key={key}
                                className="truncate font-mono text-xs text-muted-foreground transition-colors duration-[120ms] group-hover:text-foreground"
                            >
                                {key}:
                            </span>
                        ))
                    ) : (
                        <span className="font-mono text-xs text-muted-foreground">
                            {t('store.card.freeform')}
                        </span>
                    )}
                </span>

                <span className="flex flex-1 flex-col gap-1.5 p-3">
                    <span className="flex items-start gap-2">
                        <span className="min-w-0 flex-1 text-lg font-medium">{role.name}</span>
                        <EffortTag effort={role.effort} />
                    </span>
                    <span className="line-clamp-3 text-sm text-muted-foreground">
                        {role.description}
                    </span>
                </span>
            </button>

            <div className="flex h-12 shrink-0 items-center gap-2 border-t border-border px-3">
                <span className="min-w-0 flex-1 truncate font-mono text-sm tabular-nums text-muted-foreground">
                    {role.inputs.length === 0
                        ? t('store.card.noInputs')
                        : tn(
                              'store.card.inputs.one',
                              'store.card.inputs.other',
                              role.inputs.length,
                          )}
                </span>

                <AddRoleButton
                    role={role}
                    held={held}
                    busy={busy}
                    onAdd={onAdd}
                    onAlreadyHeld={onAlreadyHeld}
                />
            </div>
        </article>
    )
}
