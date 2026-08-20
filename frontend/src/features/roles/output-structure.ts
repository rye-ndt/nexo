/**
 * The grammar for a role's structured output, and the checker that runs
 * before the role can be saved. The structure is raw text a person typed,
 * so it is read by this parser and never handed to an agent to interpret.
 *
 *   summary: one paragraph a non-programmer can follow
 *   findings:
 *     - title: short label
 *       severity: high | medium | low
 *   next_steps: what to do now
 *
 * A field with a description after the colon is a leaf. A field with nothing
 * after it opens a group, and the group's fields sit two spaces deeper. A group
 * whose first line starts with "- " is a list: it describes one element, and
 * the agent repeats that element as many times as the work needs.
 */

const KEY_LINE = /^([A-Za-z_][A-Za-z0-9_]*)\s*:(.*)$/
const INDENT_STEP = 2

type StructureField = {
    key: string
    description: string
    children: StructureField[]
}

type StructureIssue = {
    line: number
    message: string
}

type Token = {
    line: number
    indent: number
    bullet: boolean
    key: string
    description: string
}

type Frame = {
    /** The column this frame's fields start at. Shifts right once a list opens. */
    childIndent: number
    /** The column a list's dashes sit at, so a second dash is caught as such. */
    bulletIndent: number | null
    owner: StructureField | null
    children: StructureField[]
}

/** The problems roleIssues shows, already carrying their line numbers. */
export function structureIssues(source: string): string[] {
    const issues: StructureIssue[] = []
    const tokens = tokenize(source, issues)

    build(tokens, issues)

    if (tokens.length === 0 && issues.length === 0)
        issues.push({line: 1, message: 'Describe at least one field.'})

    return issues
        .sort((left, right) => left.line - right.line)
        .map((issue) => `Line ${issue.line}: ${issue.message}`)
}

function tokenize(source: string, issues: StructureIssue[]): Token[] {
    const tokens: Token[] = []

    source.split('\n').forEach((text, index) => {
        const line = index + 1
        const trimmed = text.trim()

        if (trimmed === '' || trimmed.startsWith('#')) return

        if (text.includes('\t')) {
            issues.push({line, message: 'Indent with spaces, not tabs.'})
            return
        }

        const indent = text.length - text.trimStart().length
        const bullet = trimmed.startsWith('- ') || trimmed === '-'
        const content = bullet ? trimmed.replace(/^-\s*/, '') : trimmed

        if (bullet && content === '') {
            issues.push({line, message: 'A list item needs a field after the dash.'})
            return
        }

        const match = KEY_LINE.exec(content)

        if (!match) {
            issues.push({line, message: fieldShapeMessage(content)})
            return
        }

        tokens.push({
            line,
            indent,
            bullet,
            key: match[1],
            description: match[2].trim(),
        })
    })

    return tokens
}

function fieldShapeMessage(content: string): string {
    if (!content.includes(':')) return 'A field needs a colon between its name and what goes in it.'

    return 'A field name can only use letters, numbers and underscores, and cannot start with a number.'
}

function build(tokens: Token[], issues: StructureIssue[]): void {
    const root: StructureField[] = []
    const stack: Frame[] = [{childIndent: 0, bulletIndent: null, owner: null, children: root}]
    const lines = new Map<StructureField, number>()

    for (const token of tokens) {
        if (!unwindTo(token, stack, issues)) continue

        const frame = stack[stack.length - 1]

        if (token.indent !== frame.childIndent) {
            issues.push({
                line: token.line,
                message: indentMessage(token.indent, frame.childIndent),
            })
            continue
        }

        if (token.bullet && !openList(token, frame, issues)) continue

        const field: StructureField = {
            key: token.key,
            description: token.description,
            children: [],
        }

        if (frame.children.some((sibling) => sibling.key === field.key)) {
            issues.push({
                line: token.line,
                message: `Two fields at this level are both called "${field.key}".`,
            })
            continue
        }

        frame.children.push(field)
        lines.set(field, token.line)

        if (field.description === '')
            stack.push({
                // Measured from the column this field's name starts at, which is
                // past the dash on a list line rather than at token.indent.
                childIndent: frame.childIndent + INDENT_STEP,
                bulletIndent: null,
                owner: field,
                children: field.children,
            })
    }

    reportEmptyGroups(root, lines, issues)
}

/** Closes every frame the token has dedented out of, catching a second dash on the way. */
function unwindTo(token: Token, stack: Frame[], issues: StructureIssue[]): boolean {
    while (stack.length > 1 && token.indent < stack[stack.length - 1].childIndent) {
        if (token.bullet && token.indent === stack[stack.length - 1].bulletIndent) {
            issues.push({
                line: token.line,
                message:
                    'A list describes one element and the agent repeats it as often as the work needs, so it can only hold one item.',
            })
            return false
        }

        stack.pop()
    }

    return true
}

/** Opens a list under the frame's owner and shifts its fields past the dash. */
function openList(token: Token, frame: Frame, issues: StructureIssue[]): boolean {
    if (!frame.owner) {
        issues.push({
            line: token.line,
            message: 'A list needs a field above it to name what the list holds.',
        })
        return false
    }

    frame.bulletIndent = token.indent
    frame.childIndent = token.indent + INDENT_STEP

    return true
}

function indentMessage(found: number, expected: number): string {
    if (found % INDENT_STEP !== 0) return 'Indent two spaces for each level.'
    if (found > expected) return 'This field is indented too far for what it sits under.'

    return 'This field is not lined up with the fields it sits beside.'
}

function reportEmptyGroups(
    fields: StructureField[],
    lines: Map<StructureField, number>,
    issues: StructureIssue[],
): void {
    for (const field of fields) {
        if (field.description === '' && field.children.length === 0)
            issues.push({
                line: lines.get(field) ?? 1,
                message: `"${field.key}" needs either a description or fields indented under it.`,
            })

        reportEmptyGroups(field.children, lines, issues)
    }
}
