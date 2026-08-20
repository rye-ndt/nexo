import {Plus} from 'lucide-react'

import {Button} from '@/shared/ui/button'

export function EmptyCanvas({locked, onNewStep}: {locked: boolean; onNewStep: () => void}) {
    return (
        <div className="pointer-events-none absolute inset-0 flex flex-col items-center justify-center gap-3">
            <img
                src="/empty_bg.png"
                alt=""
                aria-hidden
                draggable={false}
                className="mb-4 w-[min(60%,34rem)] max-w-none opacity-35 select-none dark:opacity-45 [mask-image:radial-gradient(ellipse_at_center,black_45%,transparent_85%)]"
            />
            <p className="text-base text-muted-foreground">
                {locked ? 'This workflow has no steps.' : 'No steps yet.'}
            </p>
            {!locked && (
                <Button
                    variant="outline"
                    size="sm"
                    className="pointer-events-auto"
                    onClick={onNewStep}
                >
                    <Plus />
                    New step
                </Button>
            )}
        </div>
    )
}
