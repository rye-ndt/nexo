import {useRef, useState} from 'react'
import type {CSSProperties, KeyboardEvent, PointerEvent} from 'react'

const DRAG_SLOP = 4

type Grab = {
    workflowId: string
    from: number
    startY: number
    rows: DOMRect[]
    slot: number
    dragging: boolean
}

type Lift = {workflowId: string; from: number; slot: number; offset: number}

function slotAt(rows: DOMRect[], y: number) {
    return rows.filter((row) => row.top + row.height / 2 < y).length
}

function targetIndex(lift: {from: number; slot: number}) {
    return lift.slot > lift.from ? lift.slot - 1 : lift.slot
}

function swallowNextClick() {
    const swallow = (event: MouseEvent) => {
        event.stopPropagation()
        event.preventDefault()
    }

    window.addEventListener('click', swallow, {capture: true, once: true})
    window.setTimeout(() => window.removeEventListener('click', swallow, {capture: true}), 0)
}

export function useRailReorder(
    count: number,
    onReorder: (workflowId: string, toIndex: number) => void,
) {
    const grab = useRef<Grab | null>(null)
    const [lift, setLift] = useState<Lift | null>(null)

    const rowProps = (workflowId: string, index: number) => ({
        onPointerDown: (event: PointerEvent<HTMLElement>) => {
            if (event.button !== 0) return
            if ((event.target as HTMLElement).closest('[aria-haspopup]')) return

            const siblings = Array.from(event.currentTarget.parentElement?.children ?? [])

            grab.current = {
                workflowId,
                from: index,
                startY: event.clientY,
                rows: siblings.map((sibling) => sibling.getBoundingClientRect()),
                slot: index,
                dragging: false,
            }

            event.currentTarget.setPointerCapture(event.pointerId)
        },

        onPointerMove: (event: PointerEvent<HTMLElement>) => {
            const held = grab.current
            if (!held) return

            const offset = event.clientY - held.startY
            if (!held.dragging && Math.abs(offset) < DRAG_SLOP) return

            held.dragging = true
            held.slot = slotAt(held.rows, event.clientY)

            setLift({workflowId, from: held.from, slot: held.slot, offset})
        },

        onPointerUp: () => {
            const held = grab.current
            grab.current = null
            setLift(null)

            if (!held?.dragging) return

            swallowNextClick()

            const to = targetIndex(held)
            if (to !== held.from) onReorder(held.workflowId, to)
        },

        onPointerCancel: () => {
            grab.current = null
            setLift(null)
        },

        onKeyDown: (event: KeyboardEvent<HTMLElement>) => {
            if (!event.altKey) return

            const step = event.key === 'ArrowUp' ? -1 : event.key === 'ArrowDown' ? 1 : 0
            const to = index + step
            if (!step || to < 0 || to >= count) return

            event.preventDefault()
            onReorder(workflowId, to)
        },
    })

    const dragging = (workflowId: string) => lift?.workflowId === workflowId

    const liftStyle = (workflowId: string): CSSProperties | undefined =>
        dragging(workflowId) ? {transform: `translateY(${lift?.offset}px)`} : undefined

    const settled = lift && (lift.slot === lift.from || lift.slot === lift.from + 1)

    return {rowProps, dragging, liftStyle, dropSlot: settled ? null : (lift?.slot ?? null)}
}
