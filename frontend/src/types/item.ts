export interface Item {
  id: number
  sku: string
  name: string
  category_id: number
  category_name: string
  unit: string
  total_qty: number
  created_at: string
  updated_at: string
  deleted_at?: string
}

export interface Category {
  id: number
  name: string
  status: 'active' | 'inactive'
}

export interface Location {
  id: number
  code: string
  zone: string
  type: string
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
