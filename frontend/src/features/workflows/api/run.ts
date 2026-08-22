import {hasWailsRuntime} from '@/shared/api/bridge'
import type {Workflow} from '@/features/workflows/types'
import {
    answerRemoteAcceptance,
    cancelRemoteRun,
    discardRemoteRun,
    pauseRemoteRun,
    startRemoteRun,
} from '@/features/workflows/api/remote-run'
import {
    discardSimulatedRun,
    haltSimulatedRun,
    resolveSimulatedAcceptance,
    startSimulatedRun,
} from '@/features/workflows/api/simulated-run'

export type WorkflowRun = {
    start(workflow: Workflow): Promise<void>
    answerAcceptance(workflowId: string, stepId: string, accepted: boolean): Promise<void>
    pause(workflow: Workflow): Promise<void>
    cancel(workflow: Workflow): Promise<void>
    discard(workflow: Workflow): Promise<void>
}

const remoteRun: WorkflowRun = {
    start: startRemoteRun,
    answerAcceptance: answerRemoteAcceptance,
    pause: pauseRemoteRun,
    cancel: cancelRemoteRun,
    discard: discardRemoteRun,
}

const simulatedRun: WorkflowRun = {
    start: startSimulatedRun,
    answerAcceptance: resolveSimulatedAcceptance,
    pause: haltSimulatedRun,
    cancel: haltSimulatedRun,
    discard: discardSimulatedRun,
}

export const run: WorkflowRun = hasWailsRuntime() ? remoteRun : simulatedRun
