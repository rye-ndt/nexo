import type {InstallStage} from '@/shared/lib/enums'

export type Agent = {
    id: string
    name: string
    version: string
    installed: boolean
    loggedIn: boolean
    instanceCount: number
}

export type InstallProgress = {
    stage: InstallStage
    downloaded: number
    total: number
}
