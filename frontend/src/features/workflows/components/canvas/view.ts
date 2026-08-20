import type {Point} from '@/features/workflows/types'

export const FIT_VIEW_OPTIONS = {padding: 0.3, maxZoom: 1}

export const MIN_ZOOM = 0.3

export const MAX_ZOOM = 1.5

export const ZOOM_MS = 120

export const FIT_MS = 180

export const ORIGIN: Point = {x: 0, y: 0}

export const STEP_CURSOR_OFFSET: Point = {x: 130, y: 40}

/** Double-clicking any of these is doing something else, not asking for a step. */
export const CANVAS_CHROME = '.react-flow__node, .react-flow__panel, button'
