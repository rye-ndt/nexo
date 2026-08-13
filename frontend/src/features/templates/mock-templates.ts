import {PARAM_TYPES, TASK_LEVELS, type ParamType, type TaskLevel} from '@/shared/lib/enums'
import type {Template, TemplateArchive, TemplateRecord} from '@/features/templates/types'

export const MOCK_TEMPLATES: Template[] = [
    {
        id: '0192f3a1-0001-7000-8000-000000000001',
        name: 'Code reviewer',
        role: 'Reads a diff and reports the defects it can prove.',
        taskLevel: 'heavy_task',
        retryable: true,
        manualAcceptRequired: true,
        params: [],
        systemPrompts: [
            {
                key: 'base',
                value: 'You review code. Report only defects you can point to a line for, ranked by severity. Say nothing about style.',
            },
        ],
        outputStructure: `verdict: ship | fix first | needs a rewrite
defects:
  - file: path and line the defect sits on
    severity: high | medium | low
    proof: what makes this wrong rather than unusual
recommendation: what the next node should do about it`,
    },
    {
        id: '0192f3a1-0002-7000-8000-000000000002',
        name: 'Test writer',
        role: 'Writes the tests a change is missing and runs them.',
        taskLevel: 'daily_task',
        retryable: true,
        manualAcceptRequired: false,
        params: [
            {
                key: 'package_path',
                label: 'Package to cover',
                type: 'text',
                required: true,
            },
            {
                key: 'max_cases',
                label: 'Cases to write at most',
                type: 'number',
                required: false,
                default: '8',
            },
            {
                key: 'focus',
                label: 'What the tests should prove',
                type: 'textarea',
                required: false,
            },
        ],
        systemPrompts: [
            {
                key: 'base',
                value: 'You write tests. Cover the branches a reader would doubt, run them, and report failures verbatim.',
            },
        ],
        outputStructure: '',
    },
    {
        id: '0192f3a1-0004-7000-8000-000000000004',
        name: 'Ticket reader',
        role: 'Reads a tracker ticket and turns it into a scoped task.',
        taskLevel: 'lightweight_task',
        retryable: true,
        manualAcceptRequired: false,
        params: [
            {
                key: 'ticket_id',
                label: 'Ticket to read',
                type: 'text',
                required: true,
            },
            {
                key: 'include_comments',
                label: 'Read the comment thread too',
                type: 'boolean',
                required: false,
                default: 'true',
            },
        ],
        systemPrompts: [
            {
                key: 'base',
                value: 'Fetch the ticket at https://atlassian/tickets/{{ticket_id}} and restate it as a task with a definition of done. Quote the acceptance criteria verbatim.',
            },
        ],
        outputStructure: '',
    },
    {
        id: '0192f3a1-0003-7000-8000-000000000003',
        name: 'Doc pass',
        role: 'Brings the docs back in line with the code.',
        taskLevel: 'lightweight_task',
        retryable: false,
        manualAcceptRequired: false,
        params: [
            {
                key: 'doc_path',
                label: 'File to update',
                type: 'file',
                required: true,
            },
        ],
        systemPrompts: [
            {
                key: 'base',
                value: 'You keep documentation honest. Fix what the code contradicts and leave the voice alone.',
            },
        ],
        outputStructure: '',
    },
]

const ARCHIVE_VERSION = 1
const TEMPLATE_FILE_INVALID = 'err_template_file_invalid'
const TEMPLATE_CONFLICT = 'err_template_conflict'

function toRecord(template: Template): TemplateRecord {
    return {
        id: template.id,
        name: template.name,
        role: template.role,
        task_level: template.taskLevel,
        retryable: template.retryable,
        manual_accept_required: template.manualAcceptRequired,
        params: Object.fromEntries(
            template.params.map((param) => [
                param.key,
                {
                    description: param.label,
                    required: param.required,
                    type: param.type,
                    default: param.default ?? '',
                    options: param.options ?? [],
                },
            ]),
        ),
        system_prompts: Object.fromEntries(
            template.systemPrompts.map((prompt) => [prompt.key, prompt.value]),
        ),
        output_structure: template.outputStructure,
    }
}

function fromRecord(record: TemplateRecord): Template {
    return {
        id: record.id,
        name: record.name.trim(),
        role: record.role ?? '',
        taskLevel: record.task_level as TaskLevel,
        retryable: record.retryable ?? false,
        manualAcceptRequired: record.manual_accept_required ?? false,
        params: Object.entries(record.params ?? {}).map(([key, param]) => ({
            key,
            label: param.description,
            type: param.type as ParamType,
            required: param.required,
            default: param.default || undefined,
            options: param.options?.length ? param.options : undefined,
        })),
        systemPrompts: Object.entries(record.system_prompts ?? {}).map(([key, value]) => ({
            key,
            value,
        })),
        outputStructure: record.output_structure ?? '',
    }
}

export function mockArchive(picked: Template[]): string {
    const archive: TemplateArchive = {
        version: ARCHIVE_VERSION,
        exported_at: new Date().toISOString(),
        templates: picked.map(toRecord),
    }

    return `${JSON.stringify(archive, null, 2)}\n`
}

function recordIssues(record: TemplateRecord, index: number): string[] {
    const at = `templates[${index}]`
    const issues: string[] = []

    if (!record?.id) issues.push(`${at}.id is missing`)
    if (!record?.name?.trim()) issues.push(`${at}.name is missing`)
    if (!TASK_LEVELS.some((level) => level === record?.task_level))
        issues.push(`${at}.task_level is not one of the values this app knows`)
    if (!Object.keys(record?.system_prompts ?? {}).length)
        issues.push(`${at}.system_prompts is missing`)

    for (const [key, param] of Object.entries(record?.params ?? {})) {
        if (!PARAM_TYPES.some((type) => type === param?.type))
            issues.push(`${at}.params[${key}].type is not one of the values this app knows`)
    }

    return issues
}

function refusal(code: string, message: string) {
    return new Error(`[Err] Type: ${code} - Message: ${message} - Critical: critical`)
}

export function mockImported(current: Template[], body: string, path: string): Template[] {
    const named = `\u201c${path.slice(path.lastIndexOf('/') + 1) || path}\u201d`

    let archive: TemplateArchive

    try {
        archive = JSON.parse(body)
    } catch {
        throw refusal(TEMPLATE_FILE_INVALID, `${named} is not a template file`)
    }

    if (archive?.version !== ARCHIVE_VERSION)
        throw refusal(
            TEMPLATE_FILE_INVALID,
            `${named} is a version ${archive?.version} template file, and this app reads version ${ARCHIVE_VERSION}`,
        )

    const issues = (archive.templates ?? []).flatMap(recordIssues)
    if (!archive.templates?.length) issues.push('templates is missing')

    if (issues.length > 0)
        throw refusal(
            TEMPLATE_FILE_INVALID,
            `${named} does not hold valid templates: ${issues.join(', ')}`,
        )

    const problems: string[] = []
    const names = new Set(current.map((template) => template.name.trim().toLowerCase()))
    const ids = new Set(current.map((template) => template.id))
    const seenNames = new Set<string>()
    const seenIds = new Set<string>()

    for (const record of archive.templates) {
        const name = record.name.trim()
        const key = name.toLowerCase()

        if (seenNames.has(key)) problems.push(`the file lists \u201c${name}\u201d twice`)
        if (seenIds.has(record.id)) problems.push(`the file lists the id ${record.id} twice`)

        seenNames.add(key)
        seenIds.add(record.id)

        if (names.has(key)) {
            problems.push(`a template named \u201c${name}\u201d is already here`)
            continue
        }

        if (ids.has(record.id))
            problems.push(`\u201c${name}\u201d carries the id of a template already here`)
    }

    if (problems.length > 0)
        throw refusal(TEMPLATE_CONFLICT, `nothing was imported: ${problems.join('; ')}`)

    return archive.templates.map(fromRecord)
}
