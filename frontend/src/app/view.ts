export const AppView = {
    Workflows: 'workflows',
    Store: 'store',
} as const

export type AppView = (typeof AppView)[keyof typeof AppView]
