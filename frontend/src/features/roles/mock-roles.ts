import {INPUT_TYPES, InputType, EFFORTS, Effort} from '@/shared/lib/enums'
import type {Agent} from '@/features/agents/types'
import type {DraftContext, Role, RoleArchive, RoleDraft, RoleRecord} from '@/features/roles/types'

export const MOCK_ROLES: Role[] = [
    {
        id: '0192f3a1-0001-7000-8000-000000000001',
        name: 'Code reviewer',
        description: 'Reads a diff and reports the defects it can prove.',
        effort: 'deep',
        retryable: true,
        pauseForReview: true,
        inputs: [],
        instructions: [
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
recommendation: what the next step should do about it`,
    },
    {
        id: '0192f3a1-0002-7000-8000-000000000002',
        name: 'Test writer',
        description: 'Writes the tests a change is missing and runs them.',
        effort: 'standard',
        retryable: true,
        pauseForReview: false,
        inputs: [
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
            {
                key: 'kinds',
                label: 'Kinds of test to write',
                type: 'multiselect',
                required: true,
                default: 'unit, table-driven',
                options: ['unit', 'table-driven', 'integration', 'benchmark'],
            },
        ],
        instructions: [
            {
                key: 'base',
                value: 'You write tests. Cover the branches a reader would doubt, run them, and report failures verbatim. Write {{kinds}} tests.',
            },
        ],
        outputStructure: '',
    },
    {
        id: '0192f3a1-0004-7000-8000-000000000004',
        name: 'Ticket reader',
        description: 'Reads a tracker ticket and turns it into a scoped step.',
        effort: 'quick',
        retryable: true,
        pauseForReview: false,
        inputs: [
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
        instructions: [
            {
                key: 'base',
                value: 'Fetch the ticket at https://atlassian/tickets/{{ticket_id}} and restate it as a step with a definition of done. Quote the acceptance criteria verbatim.',
            },
        ],
        outputStructure: '',
    },
    {
        id: '0192f3a1-0003-7000-8000-000000000003',
        name: 'Doc pass',
        description: 'Brings the docs back in line with the code.',
        effort: 'quick',
        retryable: false,
        pauseForReview: false,
        inputs: [
            {
                key: 'doc_path',
                label: 'File to update',
                type: 'file',
                required: true,
            },
        ],
        instructions: [
            {
                key: 'base',
                value: 'You keep documentation honest. Fix what the code contradicts and leave the voice alone.',
            },
        ],
        outputStructure: '',
    },
]

/**
 * The heavy step level is what the helper borrows, and in this app that lands on
 * Claude Code. The browser answer follows the mock roster so that installing and
 * logging the harness in actually opens the button.
 */
export function mockHelperBlocked(agents: Agent[]): string {
    const harness = agents.find((agent) => agent.id === 'claude_code')

    if (!harness?.installed) return 'Install Claude Code to use this.'
    if (!harness.loggedIn) return 'Log in to Claude Code to use this.'

    return ''
}

/**
 * Stands in for what the agent hands back. The last prompt section echoes the context
 * it was given, so a role filled in from a workflow visibly differs from one filled
 * in from Settings.
 */
export function mockRefined(name: string, role: string, context?: DraftContext): RoleDraft {
    const subject = name.trim().toLowerCase()
    const neighbours = (context?.steps ?? []).map((step) => step.title).filter(Boolean)

    return {
        name: name.trim(),
        description: role.trim(),
        effort: Effort.Deep,
        retryable: true,
        pauseForReview: true,
        inputs: [
            {
                key: 'target_path',
                label: 'What this step should work on',
                type: InputType.Text,
                required: true,
            },
            {
                key: 'depth',
                label: 'How far to go',
                type: InputType.Select,
                required: true,
                default: 'thorough',
                options: ['quick', 'thorough', 'exhaustive'],
            },
            {
                key: 'constraints',
                label: 'Anything this step must not touch',
                type: InputType.Textarea,
                required: false,
            },
        ],
        instructions: [
            {
                key: 'base',
                value: `You are the ${subject} step of a larger job. Work on {{target_path}} and nothing else.\n\nStart by reading enough to know what is already true, then do the smallest thing that meets the goal. Go {{depth}} about it.`,
            },
            {
                key: 'limits',
                value: 'Report only what you can point at. Where you had to guess, say so and say what you guessed. Leave {{constraints}} alone.',
            },
            {
                key: 'project',
                value: [
                    `Project read: ${context?.project_dir || '(none given)'}`,
                    `Steps already in the graph: ${neighbours.join(', ') || '(none)'}`,
                ].join('\n'),
            },
        ],
        outputStructure: `summary: one paragraph a non-programmer can follow
findings:
  - title: short label for this finding
    evidence: what makes this true rather than likely
next_steps: what the following step should do`,
    }
}

const ARCHIVE_VERSION = 2
const ROLE_FILE_INVALID = 'err_role_file_invalid'
const ROLE_CONFLICT = 'err_role_conflict'

function toRecord(role: Role): RoleRecord {
    return {
        id: role.id,
        name: role.name,
        description: role.description,
        effort: role.effort,
        retryable: role.retryable,
        pause_for_review: role.pauseForReview,
        inputs: Object.fromEntries(
            role.inputs.map((input) => [
                input.key,
                {
                    description: input.label,
                    required: input.required,
                    type: input.type,
                    default: input.default ?? '',
                    options: input.options ?? [],
                },
            ]),
        ),
        instructions: Object.fromEntries(
            role.instructions.map((prompt) => [prompt.key, prompt.value]),
        ),
        output_structure: role.outputStructure,
    }
}

function fromRecord(record: RoleRecord): Role {
    return {
        id: record.id,
        name: record.name.trim(),
        description: record.description ?? '',
        effort: record.effort as Effort,
        retryable: record.retryable ?? false,
        pauseForReview: record.pause_for_review ?? false,
        inputs: Object.entries(record.inputs ?? {}).map(([key, input]) => ({
            key,
            label: input.description,
            type: input.type as InputType,
            required: input.required,
            default: input.default || undefined,
            options: input.options?.length ? input.options : undefined,
        })),
        instructions: Object.entries(record.instructions ?? {}).map(([key, value]) => ({
            key,
            value,
        })),
        outputStructure: record.output_structure ?? '',
    }
}

export function mockArchive(picked: Role[]): string {
    const archive: RoleArchive = {
        version: ARCHIVE_VERSION,
        exported_at: new Date().toISOString(),
        roles: picked.map(toRecord),
    }

    return `${JSON.stringify(archive, null, 2)}\n`
}

function recordIssues(record: RoleRecord, index: number): string[] {
    const at = `roles[${index}]`
    const issues: string[] = []

    if (!record?.id) issues.push(`${at}.id is missing`)
    if (!record?.name?.trim()) issues.push(`${at}.name is missing`)
    if (!EFFORTS.some((level) => level === record?.effort))
        issues.push(`${at}.effort is not one of the values this app knows`)
    if (!Object.keys(record?.instructions ?? {}).length)
        issues.push(`${at}.instructions is missing`)

    for (const [key, input] of Object.entries(record?.inputs ?? {})) {
        if (!INPUT_TYPES.some((type) => type === input?.type))
            issues.push(`${at}.inputs[${key}].type is not one of the values this app knows`)
    }

    return issues
}

function refusal(code: string, message: string) {
    return new Error(`[Err] Type: ${code} - Message: ${message} - Critical: critical`)
}

export function mockImported(current: Role[], body: string, path: string): Role[] {
    const named = `\u201c${path.slice(path.lastIndexOf('/') + 1) || path}\u201d`

    let archive: RoleArchive

    try {
        archive = JSON.parse(body)
    } catch {
        throw refusal(ROLE_FILE_INVALID, `${named} is not a role file`)
    }

    if (archive?.version !== ARCHIVE_VERSION)
        throw refusal(
            ROLE_FILE_INVALID,
            `${named} is a version ${archive?.version} role file, and this app reads version ${ARCHIVE_VERSION}`,
        )

    const issues = (archive.roles ?? []).flatMap(recordIssues)
    if (!archive.roles?.length) issues.push('roles is missing')

    if (issues.length > 0)
        throw refusal(ROLE_FILE_INVALID, `${named} does not hold valid roles: ${issues.join(', ')}`)

    const problems: string[] = []
    const names = new Set(current.map((role) => role.name.trim().toLowerCase()))
    const ids = new Set(current.map((role) => role.id))
    const seenNames = new Set<string>()
    const seenIds = new Set<string>()

    for (const record of archive.roles) {
        const name = record.name.trim()
        const key = name.toLowerCase()

        if (seenNames.has(key)) problems.push(`the file lists \u201c${name}\u201d twice`)
        if (seenIds.has(record.id)) problems.push(`the file lists the id ${record.id} twice`)

        seenNames.add(key)
        seenIds.add(record.id)

        if (names.has(key)) {
            problems.push(`a role named \u201c${name}\u201d is already here`)
            continue
        }

        if (ids.has(record.id))
            problems.push(`\u201c${name}\u201d carries the id of a role already here`)
    }

    if (problems.length > 0)
        throw refusal(ROLE_CONFLICT, `nothing was imported: ${problems.join('; ')}`)

    return archive.roles.map(fromRecord)
}
