import {useState} from 'react'
import {Search, X} from 'lucide-react'

import {NoticeDialog} from '@/shared/components/notice-dialog'
import {RailToggle} from '@/features/workflows/components/canvas/header/rail-toggle'
import {RoleCard} from '@/features/store/components/role-card'
import {RolePreviewDialog} from '@/features/store/components/role-preview-dialog'
import {StoreSection} from '@/features/store/types'
import {TemplateCard} from '@/features/store/components/template-card'
import {TemplatePreviewDialog} from '@/features/store/components/template-preview-dialog'
import {TourBanner} from '@/features/onboarding/components/tour-banner'
import {TourStopId} from '@/features/onboarding/tour'
import {useTour} from '@/features/onboarding/tour-context'
import {t} from '@/shared/lib/i18n'
import type {Role} from '@/features/roles/types'
import type {Store} from '@/features/store/use-store'
import type {StoreTemplate} from '@/features/store/types'
import type {WorkflowLocations} from '@/features/workflows/types'

function matches(query: string, ...fields: string[]) {
    const needle = query.trim().toLowerCase()
    if (!needle) return true

    return fields.some((field) => field.toLowerCase().includes(needle))
}

export function StoreView({
    store,
    section,
    railOpen,
    onToggleRail,
}: {
    store: Store
    section: StoreSection
    railOpen: boolean
    onToggleRail: () => void
}) {
    const tour = useTour()

    const [query, setQuery] = useState('')
    const [openTemplate, setOpenTemplate] = useState<StoreTemplate | null>(null)
    const [openRole, setOpenRole] = useState<Role | null>(null)
    const [alreadyHeld, setAlreadyHeld] = useState<Role | null>(null)

    const templates = store.templates.filter((template) =>
        matches(query, template.name, template.description),
    )
    const roles = store.roles.filter((role) => matches(query, role.name, role.description))

    const onWorkflows = section === StoreSection.Workflows
    const shown = onWorkflows ? templates.length : roles.length

    const addTemplate = async (locations: WorkflowLocations) => {
        if (!openTemplate) return
        if (!(await store.addTemplate(openTemplate, locations))) return

        setOpenTemplate(null)
        tour.leaveStore()
    }

    return (
        <>
            <main className="surface-card flex min-w-0 flex-1 flex-col overflow-hidden ring-1 ring-border-strong">
                <header className="flex h-14 shrink-0 items-center gap-3 border-b border-border-strong bg-card px-4">
                    <RailToggle open={railOpen} onToggle={onToggleRail} />
                    <span className="text-lg font-medium">{t('store.rail.title')}</span>
                    <span className="flex-1" />
                    <SearchField query={query} onChange={setQuery} />
                </header>

                <div className="min-h-0 flex-1 overflow-y-auto bg-background">
                    <div className="flex flex-col gap-1 px-4 pt-6 pb-4">
                        <h2 className="text-xl font-medium">
                            {t(onWorkflows ? 'store.workflows.title' : 'store.roles.title')}
                        </h2>
                        <p className="text-base text-muted-foreground">
                            {t(onWorkflows ? 'store.workflows.subtitle' : 'store.roles.subtitle')}
                        </p>
                    </div>

                    {shown === 0 ? (
                        <NothingFound query={query} />
                    ) : (
                        <div className="grid grid-cols-[repeat(auto-fill,minmax(260px,1fr))] gap-3 px-4 pb-6">
                            {onWorkflows
                                ? templates.map((template) => (
                                      <TemplateCard
                                          key={template.id}
                                          template={template}
                                          roleCount={store.rolesOf(template).length}
                                          onOpen={setOpenTemplate}
                                      />
                                  ))
                                : roles.map((role) => (
                                      <RoleCard
                                          key={role.id}
                                          role={role}
                                          held={store.held(role.id)}
                                          busy={store.busy}
                                          onOpen={setOpenRole}
                                          onAdd={store.addRole}
                                          onAlreadyHeld={setAlreadyHeld}
                                      />
                                  ))}
                        </div>
                    )}
                </div>

                <TourBanner stop={TourStopId.Template} />
            </main>

            {openTemplate && (
                <TemplatePreviewDialog
                    key={openTemplate.id}
                    template={openTemplate}
                    roles={store.rolesOf(openTemplate)}
                    held={store.held}
                    busy={store.addingTemplate}
                    onAdd={addTemplate}
                    onPreviewRole={setOpenRole}
                    onClose={() => setOpenTemplate(null)}
                />
            )}

            {openRole && (
                <RolePreviewDialog
                    key={openRole.id}
                    role={openRole}
                    held={store.held(openRole.id)}
                    busy={store.busy}
                    onAdd={store.addRole}
                    onAlreadyHeld={setAlreadyHeld}
                    onClose={() => setOpenRole(null)}
                />
            )}

            {alreadyHeld && (
                <NoticeDialog
                    title={t('store.already.title')}
                    description={t('store.already.body', {name: alreadyHeld.name})}
                    detail={t('store.already.detail')}
                    onClose={() => setAlreadyHeld(null)}
                />
            )}
        </>
    )
}

function SearchField({query, onChange}: {query: string; onChange: (query: string) => void}) {
    return (
        <div className="relative w-[240px]">
            <Search className="pointer-events-none absolute top-1/2 left-2.5 size-3.5 -translate-y-1/2 text-muted-foreground" />
            <input
                type="search"
                value={query}
                placeholder={t('store.header.search')}
                aria-label={t('store.header.search')}
                onChange={(event) => onChange(event.target.value)}
                className="h-8 w-full rounded-lg border border-border bg-background pr-8 pl-8 text-base outline-none placeholder:text-muted-foreground focus-visible:ring-2 focus-visible:ring-ring/50 [&::-webkit-search-cancel-button]:hidden"
            />
            {query && (
                <button
                    type="button"
                    aria-label={t('store.header.clearSearch')}
                    onClick={() => onChange('')}
                    className="absolute top-1/2 right-2 flex size-5 -translate-y-1/2 items-center justify-center rounded-md text-muted-foreground outline-none hover:bg-muted hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring/50"
                >
                    <X className="size-3.5" />
                </button>
            )}
        </div>
    )
}

function NothingFound({query}: {query: string}) {
    return (
        <div className="flex flex-col items-center gap-2 px-4 py-16 text-center">
            <p className="text-base font-medium">{t('store.empty.title', {query: query.trim()})}</p>
            <p className="text-sm text-muted-foreground">{t('store.empty.body')}</p>
        </div>
    )
}
