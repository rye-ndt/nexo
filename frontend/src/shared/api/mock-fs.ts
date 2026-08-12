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
    '/Users/rye/dev/agent-harness/.harness/context': ['handover.json', 'session.log'],
    '/Users/rye/dev/agent-harness/docs': ['architecture.md', 'diagram.png'],
    '/Users/rye/Documents/notes': ['brief.pdf', 'roadmap.md'],
    '/Users/rye/Downloads': ['export.csv', 'screenshot.png'],
}

const ILLEGAL_NAME = /[/\\]/

export function mockChildDirectories(path: string): string[] {
    return CHILDREN[path] ?? []
}

export function mockChildFiles(path: string): string[] {
    return FILES[path] ?? []
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
