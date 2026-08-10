import {useState} from 'react'

export type Decision = {
    answer: (accepted: boolean) => void
    labelOf: (accepted: boolean, resting: string, working: string) => string
}

/** Remembers which way the answer went so only that button shows its working label. */
export function useDecision(busy: boolean, onAnswer: (accepted: boolean) => void): Decision {
    const [sent, setSent] = useState<boolean | null>(null)

    return {
        answer: (accepted) => {
            if (busy) return

            setSent(accepted)
            onAnswer(accepted)
        },
        labelOf: (accepted, resting, working) => (busy && sent === accepted ? working : resting),
    }
}
