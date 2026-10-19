/** Where an item is stored, as returned inside the item list. */
export interface ItemLocation {
  location_id: number
  code: string
  zone: string
  qty: number
}

export interface Item {
  id: number
  sku: string
  name: string
  category_id: number
  category_name: string
  unit: string
  total_qty: number
  locations: ItemLocation[]
  location_count: number
  created_at: string
  updated_at: string
  deleted_at?: string
}

export interface Category {
  id: number
  name: string
  status: 'active' | 'inactive'
  /** Number of live items in this category (list endpoint). */
  item_count: number
}

export interface Location {
  id: number
  code: string
  zone: string
  type: string
  /** Number of distinct live items stored at this location (list endpoint). */
  item_count: number
}

/** Payload of POST /items and PUT /items/:id. */
export interface ItemPayload {
  sku: string
  name: string
  category_id: number
  unit: string
}

/** Query of GET /items. */
export interface ItemListParams {
  search?: string
  category_id?: number | null
  page: number
  limit: number
  sort_by?: 'id' | 'sku' | 'name' | 'category_id' | 'created_at'
  sort_dir?: 'asc' | 'desc'
}

/** Units offered by the item form; free text is still allowed. */
export const UNIT_OPTIONS = [
  'pcs',
  'box',
  'roll',
  'batang',
  'sak',
  'liter',
  'lusin',
  'unit',
  'set',
  'meter',
] as const

/** Payload of POST /items when the item is placed on a rack right away. */
export interface ItemCreatePayload extends ItemPayload {
  location_id?: number
  initial_qty?: number
}

/** Payload of POST /categories. */
export interface CategoryPayload {
  name: string
  status: 'active' | 'inactive'
}

/** Payload of POST /locations. */
export interface LocationPayload {
  code: string
  zone: string
  type: LocationType
}

export const LOCATION_TYPES = ['rack', 'bin', 'pallet', 'cold_storage', 'staging'] as const

export type LocationType = (typeof LOCATION_TYPES)[number]

/** Human readable labels for the location types. */
export const LOCATION_TYPE_LABELS: Record<LocationType, string> = {
  rack: 'Rak',
  bin: 'Bin',
  pallet: 'Pallet',
  cold_storage: 'Cold storage',
  staging: 'Staging',
}
