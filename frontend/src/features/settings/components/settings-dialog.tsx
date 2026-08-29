import {useState} from 'react'
import {Boxes, LayoutTemplate, Plug, SlidersHorizontal} from 'lucide-react'

import {AgentsPanel} from '@/features/agents/components/agents-panel'
import {MCPPanel} from '@/features/settings/components/mcp-panel'
import {PreferencesPanel} from '@/features/settings/components/preferences-panel'
import {RolesPanel} from '@/features/roles/components/roles-panel'
import {SettingsNav, type SettingsTab} from '@/features/settings/components/settings-nav'
import {t} from '@/shared/lib/i18n'
import {Dialog, DialogContent, DialogDescription, DialogTitle} from '@/shared/ui/dialog'

const SettingsTabId = {
    Preferences: 'preferences',
    Roles: 'roles',
    MCP: 'mcp',
    Agents: 'agents',
} as const

type SettingsTabId = (typeof SettingsTabId)[keyof typeof SettingsTabId]

const TABS: SettingsTab<SettingsTabId>[] = [
    {id: SettingsTabId.Preferences, label: 'settings.tab.preferences', icon: SlidersHorizontal},
    {id: SettingsTabId.Roles, label: 'settings.tab.roles', icon: LayoutTemplate},
    {id: SettingsTabId.MCP, label: 'settings.tab.mcp', icon: Plug},
    {id: SettingsTabId.Agents, label: 'settings.tab.agents', icon: Boxes},
]

export function SettingsDialog({
    open,
    onOpenChange,
}: {
    open: boolean
    onOpenChange: (open: boolean) => void
}) {
    const [activeId, setActiveId] = useState<SettingsTabId>(TABS[0].id)
    const active = TABS.find((tab) => tab.id === activeId) ?? TABS[0]

    return (
        <Dialog open={open} onOpenChange={onOpenChange}>
            <DialogContent className="flex h-[560px] max-w-[760px] gap-0 overflow-hidden rounded-[20px] bg-card p-0">
                <DialogDescription className="sr-only">
                    {t('settings.dialog.description')}
                </DialogDescription>

                <aside className="flex w-[200px] shrink-0 flex-col border-r border-border bg-sidebar">
                    <div className="flex h-14 shrink-0 items-center border-b border-border px-4">
                        <span className="micro-label">{t('settings.dialog.title')}</span>
                    </div>
                    <SettingsNav tabs={TABS} activeId={active.id} onSelect={setActiveId} />
                </aside>

                <div className="flex min-w-0 flex-1 flex-col">
                    <div className="flex h-14 shrink-0 items-center border-b border-border px-4">
                        <DialogTitle className="text-base">{t(active.label)}</DialogTitle>
                    </div>

                    <div className="min-h-0 flex-1 scroll-py-6 overflow-y-auto">
                        {active.id === SettingsTabId.Preferences && (
                            <PreferencesPanel
                                onShowAgents={() => setActiveId(SettingsTabId.Agents)}
                            />
                        )}
                        {active.id === SettingsTabId.Roles && <RolesPanel />}
                        {active.id === SettingsTabId.MCP && <MCPPanel />}
                        {active.id === SettingsTabId.Agents && <AgentsPanel />}
                    </div>
                </div>
            </DialogContent>
        </Dialog>
    )
}
