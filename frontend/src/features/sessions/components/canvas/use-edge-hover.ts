import {useCallback, useEffect, useRef, useState} from 'react'

/**
 * The unlink control sits on the line but lives outside its SVG, so the pointer
 * leaves the edge on its way to the X. Holding the reveal open for a moment after
 * the pointer leaves covers that gap, and covers a wobble off a 1.75px line.
 */
const RELEASE_MS = 120

export function useEdgeHover() {
    const [hoveredEdgeId, setHoveredEdgeId] = useState<string | null>(null)
    const timer = useRef<ReturnType<typeof setTimeout> | null>(null)

    const clear = () => {
        if (timer.current) clearTimeout(timer.current)
        timer.current = null
    }

    useEffect(() => clear, [])

    const hold = useCallback((edgeId: string) => {
        clear()
        setHoveredEdgeId(edgeId)
    }, [])

    const release = useCallback(() => {
        clear()
        timer.current = setTimeout(() => setHoveredEdgeId(null), RELEASE_MS)
    }, [])

    return {hoveredEdgeId, hold, release}
}
