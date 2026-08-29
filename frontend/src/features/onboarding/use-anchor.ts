import {useEffect, useState} from 'react'

import type {Rect} from '@/features/onboarding/types'

type Tracked = {selector: string; rect: Rect | null}

const round = (value: DOMRect): Rect => ({
    top: Math.round(value.top),
    left: Math.round(value.left),
    width: Math.round(value.width),
    height: Math.round(value.height),
})

const same = (a: Rect | null, b: Rect | null) =>
    a === b ||
    (a !== null &&
        b !== null &&
        a.top === b.top &&
        a.left === b.left &&
        a.width === b.width &&
        a.height === b.height)

/**
 * The rect of whatever the selector matches right now, or null while it is not
 * on screen. Read per frame rather than on resize: the element the tour points
 * at arrives with a dialog animation, moves when the canvas pans, and resizes
 * when the rail opens. The rect is stored with the selector it came from, so a
 * stop change never shows the previous stop's position for a frame, and state
 * only changes when the rect does.
 *
 * A match with no size counts as no match: an element hidden by a breakpoint is
 * still in the DOM and still measures, and a zero rect would ring a point in the
 * corner instead of letting the card come adrift.
 */
export function useAnchorRect(selector: string | null): Rect | null {
    const [tracked, setTracked] = useState<Tracked | null>(null)

    useEffect(() => {
        if (!selector) return

        let frame = 0

        const read = () => {
            const target = document.querySelector(selector)
            const box = target?.getBoundingClientRect()
            const next = box?.width && box.height ? round(box) : null

            setTracked((current) =>
                current && current.selector === selector && same(current.rect, next)
                    ? current
                    : {selector, rect: next},
            )

            frame = requestAnimationFrame(read)
        }

        frame = requestAnimationFrame(read)
        return () => cancelAnimationFrame(frame)
    }, [selector])

    return tracked?.selector === selector ? tracked.rect : null
}
