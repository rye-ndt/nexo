/**
 * The only place the generated Wails bindings for onboarding are touched.
 * Onboarding is a once-per-machine event, so the flag lives in the Go user
 * config: uninstalling a harness later must not drag the welcome modal back.
 * Under the plain vite dev server there is no Go side, so it lives in this
 * module's memory — start with ?onboarded=1 to reach the settled app directly.
 */

import {bridge, hasWailsRuntime} from '@/shared/api/bridge'
import {CompleteOnboarding, Onboarded} from '@wailsjs/go/wails_api/API'

let onboarded = new URLSearchParams(window.location.search).get('onboarded') === '1'

type OnboardingBackend = {
    read(): Promise<boolean>
    complete(): Promise<void>
}

const storedFlag: OnboardingBackend = {
    read: () => bridge(Onboarded),
    complete: async () => {
        await bridge(CompleteOnboarding)
    },
}

const memoryFlag: OnboardingBackend = {
    read: async () => onboarded,
    complete: async () => {
        onboarded = true
    },
}

const backend: OnboardingBackend = hasWailsRuntime() ? storedFlag : memoryFlag

export async function hasOnboarded(): Promise<boolean> {
    return backend.read()
}

export async function completeOnboarding(): Promise<void> {
    return backend.complete()
}
