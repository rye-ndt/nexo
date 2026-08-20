/** Enough of a tree for the stand-in path picker to feel like a real one. */

import {joinPath} from '@/shared/lib/path'

export const MOCK_HOME = '/Users/rye'

const CHILDREN: Record<string, string[]> = {
    '/': ['Users', 'opt', 'tmp'],
    '/Users': ['rye'],
    '/Users/rye': ['dev', 'Documents', 'Downloads'],
    '/Users/rye/dev': ['agent-harness', 'hexago-ui', 'scratch'],
    '/Users/rye/dev/agent-harness': ['.harness', 'frontend', 'internal', 'docs'],
    '/Users/rye/dev/agent-harness/.harness': ['context', 'logs', 'wal'],
    '/Users/rye/dev/agent-harness/frontend': ['src', 'public'],
    '/Users/rye/dev/agent-harness/internal': ['implementation', 'interface', 'helpers'],
    '/Users/rye/dev/hexago-ui': ['src', '.harness'],
    '/Users/rye/dev/hexago-ui/.harness': ['context'],
    '/Users/rye/Documents': ['notes'],
    '/Users/rye/Downloads': ['agent-harness'],
    '/Users/rye/Downloads/agent-harness': ['.harness', 'frontend'],
    '/Users/rye/Downloads/agent-harness/.harness': ['context'],
}

const FILES: Record<string, string[]> = {
    '/Users/rye': ['.zshrc'],
    '/Users/rye/dev/agent-harness': ['README.md', 'config.yaml', 'go.mod'],
    '/Users/rye/dev/agent-harness/.harness/context': ['handoff.json', 'workflow.log'],
    '/Users/rye/dev/agent-harness/docs': ['architecture.md', 'diagram.png'],
    '/Users/rye/Documents/notes': ['brief.pdf', 'roadmap.md'],
    '/Users/rye/Downloads': [
        'export.csv',
        'screenshot.png',
        'roles-sound.json',
        'roles-torn.json',
        'roles-incomplete.json',
        'workflow-sound.json',
        'workflow-torn.json',
    ],
}

const ILLEGAL_NAME = /[/\\]/

export function mockChildDirectories(path: string): string[] {
    return CHILDREN[path] ?? []
}

export function mockChildFiles(path: string): string[] {
    return FILES[path] ?? []
}

const CONTENTS: Record<string, string> = {
    '/Users/rye/Downloads/roles-sound.json': JSON.stringify(
        {
            version: 2,
            exported_at: '2026-08-13T09:00:00Z',
            roles: [
                {
                    id: '0192f3a1-0900-7000-8000-000000000009',
                    name: 'Release notes writer',
                    description: 'Turns a set of merged changes into notes a user can read.',
                    effort: 'standard',
                    retryable: true,
                    pause_for_review: false,
                    inputs: {
                        since_tag: {
                            description: 'Tag to start from',
                            required: true,
                            type: 'text',
                            default: '',
                            options: [],
                        },
                    },
                    instructions: {
                        base: 'You write release notes. One line per change, in the words a user would use.',
                    },
                    output_structure: 'highlights: the three changes that matter most',
                },
            ],
        },
        null,
        2,
    ),
    '/Users/rye/Downloads/roles-torn.json': '{ this was edited by hand and is not json',
    '/Users/rye/Downloads/workflow-sound.json': JSON.stringify(
        {
            version: 2,
            exported_at: '2026-08-14T11:20:00Z',
            workflow: {
                id: '0198e3a1-0000-7000-8000-000000000101',
                name: 'Gateway credential sweep',
                createdAt: '2026-08-14T11:20:00Z',
                locked: false,
                started: false,
                cancelled: false,
                projectDir: '/Users/rye/dev/agent-harness',
                steps: [
                    {
                        id: '0198e3a1-0000-7000-8000-000000000102',
                        title: 'Read how credentials are revoked',
                        prompt: 'Follow a revoked credential from the proxy to the agent and say what still holds a copy of it.',
                        state: 'idle',
                        position: {x: 0, y: 120},
                        dependsOn: [],
                        spec: {
                            effort: 'deep',
                            instructions: [
                                'You review code. Report only defects you can point to a line for, ranked by severity.',
                            ],
                            outputStructure:
                                'verdict: ship | fix first\nfindings: what you can prove',
                            pauseForReview: true,
                        },
                        values: {},
                    },
                    {
                        id: '0198e3a1-0000-7000-8000-000000000103',
                        title: 'Cover the revoke path',
                        prompt: 'Write the tests that prove a revoked credential never reaches a running agent.',
                        state: 'idle',
                        position: {x: 340, y: 120},
                        dependsOn: ['0198e3a1-0000-7000-8000-000000000102'],
                        spec: {
                            effort: 'standard',
                            instructions: [
                                'You write tests. Cover the branches a reader would doubt, run them, and report failures verbatim.',
                            ],
                            outputStructure: '',
                            pauseForReview: false,
                        },
                        values: {},
                    },
                ],
            },
        },
        null,
        2,
    ),
    '/Users/rye/Downloads/workflow-torn.json': '{ not json',
    '/Users/rye/Downloads/roles-incomplete.json': JSON.stringify(
        {
            version: 2,
            exported_at: '2026-08-13T09:00:00Z',
            roles: [
                {
                    id: '0192f3a1-0910-7000-8000-000000000010',
                    name: 'Nameless prompt',
                    effort: 'weekly',
                    instructions: {},
                },
            ],
        },
        null,
        2,
    ),
}

export function mockReadFile(path: string): string {
    const body = CONTENTS[path]
    if (body === undefined) throw new Error(`${path} could not be read.`)

    return body
}

export function mockWriteFile(path: string, body: string) {
    const parent = path.slice(0, path.lastIndexOf('/')) || '/'
    const name = path.slice(path.lastIndexOf('/') + 1)

    if (!name) throw new Error('Give the file a name.')

    CONTENTS[path] = body

    const siblings = FILES[parent] ?? []
    if (!siblings.includes(name)) FILES[parent] = [...siblings, name]
}

/** Mirrors what the OS dialog's New Folder button does, and fails the same ways. */
export function mockCreateDirectory(parent: string, name: string): string {
    const trimmed = name.trim()

    if (!trimmed) throw new Error('Give the folder a name.')
    if (ILLEGAL_NAME.test(trimmed)) throw new Error('A folder name cannot contain a slash.')

    const siblings = CHILDREN[parent] ?? []
    if (siblings.includes(trimmed)) throw new Error(`${trimmed} already exists here.`)

    CHILDREN[parent] = [...siblings, trimmed]
    return joinPath(parent, trimmed)
}
