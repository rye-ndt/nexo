import {Plus} from 'lucide-react'

import {Button} from '@/shared/ui/button'

export function EmptyCanvas({locked, onNewNode}: {locked: boolean; onNewNode: () => void}) {
    return (
        <div className="pointer-events-none absolute inset-0 flex flex-col items-center justify-center gap-3">
            <p className="text-base text-muted-foreground">
                {locked ? 'This session has no nodes.' : 'No nodes yet.'}
            </p>
            {!locked && (
                <Button
                    variant="outline"
                    size="sm"
                    className="pointer-events-auto"
                    onClick={onNewNode}
                >
                    <Plus />
                    New node
                </Button>
            )}
        </div>
    )
}
