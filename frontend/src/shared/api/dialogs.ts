/**
 * The path-chooser seam. Inside the Wails webview these are the real native
 * dialogs — ChooseDirectory wraps runtime.OpenDirectoryDialog with
 * CanCreateDirectories, so creating a folder comes from the OS, and ChooseFile
 * wraps runtime.OpenFileDialog with no filter, so any format can be picked.
 * Both return an absolute path or "" when the user cancels. Under the plain
 * vite dev server there is no runtime and the File System Access API refuses to
 * disclose absolute paths, so a stand-in picker answers the same contract.
 */

import {hasWailsRuntime} from '@/shared/api/bridge'
import {
    mockChildDirectories,
    mockChildFiles,
    mockCreateDirectory,
    MOCK_HOME,
} from '@/shared/api/mock-fs'
import {ChooseDirectory, ChooseFile} from '@wailsjs/go/wails_api/API'

export type PathKind = 'directory' | 'file'

type PathChooser = (kind: PathKind, title: string) => Promise<string>

let standIn: PathChooser | null = null

export function registerPathChooser(next: PathChooser | null) {
    standIn = next
}

/** True when the OS dialogs answer these, so the stand-in never needs to mount. */
export function hasNativePathPicker() {
    return hasWailsRuntime()
}

/** Resolves to the chosen absolute path, or an empty string if the user cancels. */
async function choosePath(kind: PathKind, title: string): Promise<string> {
    if (hasNativePathPicker()) return kind === 'file' ? ChooseFile(title) : ChooseDirectory(title)

    if (!standIn) throw new Error('No path picker is available right now.')

    return standIn(kind, title)
}

export async function chooseDirectory(title: string): Promise<string> {
    return choosePath('directory', title)
}

export async function chooseFile(title: string): Promise<string> {
    return choosePath('file', title)
}

export async function listDirectories(path: string): Promise<string[]> {
    return mockChildDirectories(path)
}

export async function listFiles(path: string): Promise<string[]> {
    return mockChildFiles(path)
}

export async function createDirectory(parent: string, name: string): Promise<string> {
    return mockCreateDirectory(parent, name)
}

export async function homeDirectory(): Promise<string> {
    return MOCK_HOME
}
