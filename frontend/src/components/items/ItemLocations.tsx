import { Badge } from '@/components/ui/Badge'
import type { ItemLocation } from '@/types/item'

interface ItemLocationsProps {
  locations: ItemLocation[]
  unit: string
  /** How many location codes to show before collapsing into "+N". */
  visible?: number
}

/**
 * Compact "where is it stored" cell: the first location codes as badges and a
 * "+N" chip for the rest. The full breakdown is in the title attribute, so the
 * table stays readable without an extra request per row.
 */
export function ItemLocations({ locations, unit, visible = 2 }: ItemLocationsProps) {
  if (locations.length === 0) {
    return <span className="text-xs text-slate-400">Belum ditempatkan</span>
  }

  const shown = locations.slice(0, visible)
  const hidden = locations.slice(visible)
  const fullList = locations.map((l) => `${l.code} (${l.qty} ${unit})`).join(', ')

  return (
    <span className="inline-flex flex-wrap items-center gap-1" title={fullList}>
      {shown.map((location) => (
        <Badge key={location.location_id} tone="slate" className="font-mono">
          {location.code}
          <span className="ml-1 font-sans font-normal text-slate-500">· {location.qty}</span>
        </Badge>
      ))}

      {hidden.length > 0 && (
        <Badge tone="indigo" className="font-semibold">
          +{hidden.length}
        </Badge>
      )}
    </span>
  )
}
