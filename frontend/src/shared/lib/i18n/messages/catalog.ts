/** Both languages side by side, so neither can drift from the other. */
export type Message = {en: string; vi: string}

export type Catalog = Record<string, Message>
