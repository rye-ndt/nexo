import {useState} from 'react'
import {Download, Plus, Upload} from 'lucide-react'

import {Button} from '@/shared/ui/button'
import {NoticeDialog} from '@/shared/components/notice-dialog'
import {TemplateExportDialog} from '@/features/templates/components/template-export-dialog'
import {TemplateFormDialog} from '@/features/templates/components/template-form-dialog'
import {TemplateList} from '@/features/templates/components/template-list'
import {WorkingDialog} from '@/shared/components/working-dialog'
import {chooseFile, chooseSaveFile} from '@/shared/api/dialogs'
import {pluralize} from '@/shared/lib/format'
import {reportError} from '@/shared/lib/error-bus'
import {useTemplates, useTemplateTransfer} from '@/features/templates/use-templates'
import {useToggle} from '@/shared/hooks/use-toggle'
import type {DraftContext, Template} from '@/features/templates/types'

const JSON_FILES = '*.json'

type TemplateEdit = {template: Template | null}

type Notice = {title: string; description: string; detail: string}

function exportFileName() {
    return `nexo-templates-${new Date().toISOString().slice(0, 10)}.json`
}

export function TemplatesPanel({
    context,
    onPick,
}: {
    context?: DraftContext
    onPick?: (template: Template) => void
}) {
    const {templates, loading, removeTemplate} = useTemplates()
    const transfer = useTemplateTransfer()

    const [editing, setEditing] = useState<TemplateEdit | null>(null)
    const [notice, setNotice] = useState<Notice | null>(null)
    const exporting = useToggle()

    const newTemplate = () => setEditing({template: null})
    const editTemplate = (template: Template) => setEditing({template})
    const closeForm = () => setEditing(null)
    const dismissNotice = () => setNotice(null)

    const pickPath = async (pick: () => Promise<string>) => {
        try {
            return await pick()
        } catch (cause) {
            reportError(cause, 'Could not open the file picker')
            return ''
        }
    }

    const runExport = async (templateIds: string[]) => {
        const path = await pickPath(() =>
            chooseSaveFile('Export templates', exportFileName(), JSON_FILES),
        )
        if (!path) return

        const count = await transfer.exportTemplates({templateIds, path}).catch(() => null)
        if (count === null) return

        exporting.close()
        setNotice({
            title: `Exported ${pluralize(count, 'template')}`,
            description: 'The file is yours to keep, move, or hand to someone else.',
            detail: path,
        })
    }

    const runImport = async () => {
        const path = await pickPath(() => chooseFile('Import templates', JSON_FILES))
        if (!path) return

        const count = await transfer.importTemplates(path).catch(() => null)
        if (count === null) return

        setNotice({
            title: `Imported ${pluralize(count, 'template')}`,
            description: 'They are ready to build nodes from.',
            detail: path,
        })
    }

    return (
        <section className="flex flex-col">
            <div className="flex items-start gap-3 px-4 pt-4 pb-3">
                <div className="flex min-w-0 flex-1 flex-col gap-1">
                    <h3 className="text-lg font-medium">
                        {onPick ? 'Pick a template' : 'Templates your nodes start from'}
                    </h3>
                    <p className="text-sm text-muted-foreground">
                        One kind of work: the agent's role, its inputs, and how hard to try.
                    </p>
                </div>

                {templates.length > 0 && (
                    <Button variant="outline" size="sm" className="shrink-0" onClick={newTemplate}>
                        <Plus />
                        New template
                    </Button>
                )}
            </div>

            <div className="border-t border-border">
                <TemplateList
                    templates={templates}
                    loading={loading}
                    onPick={onPick ?? editTemplate}
                    onEdit={editTemplate}
                    onRemove={removeTemplate}
                    onCreate={newTemplate}
                />
            </div>

            {!onPick && (
                <div className="flex items-center gap-1 border-t border-border px-4 py-3">
                    <p className="min-w-0 flex-1 truncate text-sm text-muted-foreground">
                        Templates travel as a .json file.
                    </p>

                    <Button variant="ghost" size="sm" onClick={runImport}>
                        <Upload />
                        Import
                    </Button>

                    {templates.length > 0 && (
                        <Button variant="ghost" size="sm" onClick={exporting.open}>
                            <Download />
                            Export
                        </Button>
                    )}
                </div>
            )}

            {editing && (
                <TemplateFormDialog
                    template={editing.template}
                    context={context}
                    onClose={closeForm}
                />
            )}

            {exporting.on && (
                <TemplateExportDialog
                    templates={templates}
                    onExport={runExport}
                    onClose={exporting.close}
                />
            )}

            {transfer.sending && (
                <WorkingDialog
                    title="Exporting templates"
                    description="Writing the file. Hold on."
                />
            )}

            {transfer.receiving && (
                <WorkingDialog
                    title="Importing templates"
                    description="Reading the file. Hold on."
                />
            )}

            {notice && (
                <NoticeDialog
                    title={notice.title}
                    description={notice.description}
                    detail={notice.detail}
                    onClose={dismissNotice}
                />
            )}
        </section>
    )
}
