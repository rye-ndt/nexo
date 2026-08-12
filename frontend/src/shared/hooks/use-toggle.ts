import {useCallback, useState} from 'react'

type Toggle = {
    on: boolean
    open: () => void
    close: () => void
    toggle: () => void
    /** For controlled components that report the state they want, such as Radix `onOpenChange`. */
    set: (on: boolean) => void
}

export function useToggle(initial = false): Toggle {
    const [on, setOn] = useState(initial)

    return {
        on,
        open: useCallback(() => setOn(true), []),
        close: useCallback(() => setOn(false), []),
        toggle: useCallback(() => setOn((current) => !current), []),
        set: setOn,
    }
}
