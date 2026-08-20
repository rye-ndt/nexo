import React from 'react'
import {createRoot} from 'react-dom/client'
import './index.css'
import App from './App'
import {AppProviders} from '@/app/providers'
import {loadLanguage} from '@/features/settings/api/preferences'

const container = document.getElementById('root')

const root = createRoot(container!)

/**
 * The stored language is read before the first paint so a Vietnamese user never
 * watches the app correct itself from English. A bridge call that never settles
 * must not cost them the window, so the wait is bounded — `applyLanguage`
 * repaints through `useLanguage` whenever the answer does arrive.
 */
const SETTLE_MS = 1500

const settled = new Promise((resolve) => setTimeout(resolve, SETTLE_MS))

void Promise.race([loadLanguage(), settled]).finally(() =>
    root.render(
        <React.StrictMode>
            <AppProviders>
                <App />
            </AppProviders>
        </React.StrictMode>,
    ),
)
