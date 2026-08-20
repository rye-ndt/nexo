import {Panel, useReactFlow} from '@xyflow/react'
import {Maximize, Minus, Plus} from 'lucide-react'
import type {ComponentType} from 'react'

import {Button} from '@/shared/ui/button'
import {FIT_MS, FIT_VIEW_OPTIONS, ZOOM_MS} from '@/features/workflows/components/canvas/view'
import {t} from '@/shared/lib/i18n'

export function ZoomCluster() {
    const {zoomIn, zoomOut, fitView} = useReactFlow()

    return (
        <Panel position="bottom-left" style={{margin: 12}}>
            <div className="flex flex-col divide-y divide-border overflow-hidden rounded-lg bg-card shadow-[0_2px_16px_rgba(27,28,30,0.04)] ring-1 ring-border">
                <ZoomButton
                    label={t('canvas.zoom.in')}
                    icon={Plus}
                    onClick={() => zoomIn({duration: ZOOM_MS})}
                />
                <ZoomButton
                    label={t('canvas.zoom.out')}
                    icon={Minus}
                    onClick={() => zoomOut({duration: ZOOM_MS})}
                />
                <ZoomButton
                    label={t('canvas.zoom.fit')}
                    icon={Maximize}
                    onClick={() => fitView({...FIT_VIEW_OPTIONS, duration: FIT_MS})}
                />
            </div>
        </Panel>
    )
}

function ZoomButton({
    label,
    icon: Icon,
    onClick,
}: {
    label: string
    icon: ComponentType<{className?: string}>
    onClick: () => void
}) {
    return (
        <Button
            variant="ghost"
            size="icon-sm"
            aria-label={label}
            className="rounded-none"
            onClick={onClick}
        >
            <Icon />
        </Button>
    )
}
