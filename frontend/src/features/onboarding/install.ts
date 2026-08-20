import {InstallStage} from '@/shared/lib/enums'
import {clampRatio} from '@/shared/lib/format'
import type {InstallProgress} from '@/features/agents/types'

const STAGE_FLOOR: Record<InstallStage, number> = {
    [InstallStage.Queued]: 0,
    [InstallStage.Resolve]: 0.05,
    [InstallStage.Download]: 0.1,
    [InstallStage.Extract]: 0.9,
    [InstallStage.Done]: 1,
}

const DOWNLOAD_SHARE = 0.8

export function installRatio(progress: InstallProgress | null): number {
    if (!progress) return 0

    const floor = STAGE_FLOOR[progress.stage] ?? 0
    if (progress.stage !== InstallStage.Download || progress.total <= 0) return floor

    return floor + DOWNLOAD_SHARE * clampRatio(progress.downloaded / progress.total)
}
