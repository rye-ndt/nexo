/**
 * The path-chooser seam. Inside the Wails webview these are the real native
 * dialogs — ChooseDirectory wraps runtime.OpenDirectoryDialog with
 * CanCreateDirectories, so creating a folder comes from the OS, ChooseFile
 * wraps runtime.OpenFileDialog and ChooseSaveFile wraps runtime.SaveFileDialog.
 * A pattern such as "*.json" narrows what the OS offers; an empty one takes any
 * format. All three return an absolute path or "" when the user cancels. Under
 * the plain vite dev server there is no runtime and the File System Access API
 * refuses to disclose absolute paths, so a stand-in picker answers the same
 * contract.
 */

import {hasWailsRuntime} from '@/shared/api/bridge'
import {
    mockChildDirectories,
    mockChildFiles,
    mockCreateDirectory,
    MOCK_HOME,
} from '@/shared/api/mock-fs'
import {t} from '@/shared/lib/i18n'
import {ChooseDirectory, ChooseFile, ChooseSaveFile} from '@wailsjs/go/wails_api/API'

export type PathKind = 'directory' | 'file' | 'save'

export type PathRequest = {
    kind: PathKind
    title: string
    pattern: string
    defaultName: string
}

type PathChooser = (request: PathRequest) => Promise<string>

let standIn: PathChooser | null = null

export function registerPathChooser(next: PathChooser | null) {
    standIn = next
}

/** True when the OS dialogs answer these, so the stand-in never needs to mount. */
export function hasNativePathPicker() {
    return hasWailsRuntime()
}

/** Resolves to the chosen absolute path, or an empty string if the user cancels. */
async function choosePath(request: PathRequest): Promise<string> {
    if (hasNativePathPicker()) {
        if (request.kind === 'directory') return ChooseDirectory(request.title)
        if (request.kind === 'file') return ChooseFile(request.title, request.pattern)

        return ChooseSaveFile(request.title, request.defaultName, request.pattern)
    }

    if (!standIn) throw new Error(t('shared.pathPicker.unavailable'))

    return standIn(request)
}

export async function chooseDirectory(title: string): Promise<string> {
    return choosePath({kind: 'directory', title, pattern: '', defaultName: ''})
}

export async function chooseFile(title: string, pattern = ''): Promise<string> {
    return choosePath({kind: 'file', title, pattern, defaultName: ''})
}

export async function chooseSaveFile(
    title: string,
    defaultName: string,
    pattern = '',
): Promise<string> {
    const path = await choosePath({kind: 'save', title, pattern, defaultName})
    const extension = pattern.replace('*', '')

    if (!path || !extension || path.endsWith(extension)) return path

    return path + extension
}

export async function listDirectories(path: string): Promise<string[]> {
    return mockChildDirectories(path)
}

export async function listFiles(path: string, pattern = ''): Promise<string[]> {
    const files = mockChildFiles(path)
    if (!pattern) return files

    const extension = pattern.replace('*', '')

    return files.filter((name) => name.endsWith(extension))
}

export async function createDirectory(parent: string, name: string): Promise<string> {
    return mockCreateDirectory(parent, name)
}

export async function homeDirectory(): Promise<string> {
    return MOCK_HOME
}
