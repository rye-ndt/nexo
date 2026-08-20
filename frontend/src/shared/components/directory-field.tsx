import {Folder} from 'lucide-react'

import {chooseDirectory} from '@/shared/api/dialogs'
import {Button} from '@/shared/ui/button'
import {HelpTip} from '@/shared/components/help-tip'
import type {GlossaryTerm} from '@/shared/lib/glossary'
import {reportError} from '@/shared/lib/error-bus'
import {t} from '@/shared/lib/i18n'

export function DirectoryField({
    label,
    hint,
    term,
    title,
    value,
    onChange,
}: {
    label: string
    hint: string
    term?: GlossaryTerm
    title: string
    value: string
    onChange: (path: string) => void
}) {
    const choose = async () => {
        try {
            const path = await chooseDirectory(title)
            if (path) onChange(path)
        } catch (cause) {
            reportError(cause, t('shared.directory.pickerFailed'))
        }
    }

    return (
        <div className="flex flex-col gap-2">
            <span className="flex items-center gap-2">
                <span className="text-base font-medium">{label}</span>
                {term && <HelpTip term={term} />}
            </span>

            <div className="flex items-center gap-2">
                <span
                    className={
                        value
                            ? 'min-w-0 flex-1 truncate rounded-lg border border-input px-3 py-2 font-mono text-base'
                            : 'min-w-0 flex-1 truncate rounded-lg border border-dashed border-input px-3 py-2 text-base text-muted-foreground'
                    }
                >
                    {value || t('shared.directory.empty')}
                </span>

                <Button variant="outline" size="sm" className="shrink-0" onClick={choose}>
                    <Folder />
                    {value ? t('shared.directory.change') : t('shared.directory.choose')}
                </Button>
            </div>

            <p className="text-sm text-muted-foreground">{hint}</p>
        </div>
    )
}
