/**
 * Shape of every response coming from the SISWA backend.
 *
 * Success: { success, code: 0, message, data, meta? }
 * Error:   { success: false, code, message, errors: [{ field, reason }] }
 */

export interface ApiMeta {
  page: number
  limit: number
  total: number
  total_pages: number
}

export interface ApiFieldError {
  field: string
  reason: string
}

export interface ApiEnvelope<TData> {
  success: boolean
  code: number
  message: string
  data?: TData
  meta?: ApiMeta
  errors?: ApiFieldError[]
}

/** Business codes returned by the API (0 = sukses, 99 = generic error). */
export const ApiCode = {
  SUCCESS: 0,
  ERROR: 99,
  VALIDATION: 1001,
  NOT_FOUND: 1002,
  DUPLICATE_SKU: 1003,
  NEGATIVE_STOCK: 1004,
  INVALID_CATEGORY: 1005,
  INVALID_LOCATION: 1006,
  ITEM_DELETED: 1007,
  INVALID_PAYLOAD: 1008,
  INVALID_QUANTITY: 1009,
  ROUTE_NOT_FOUND: 1010,
  DATABASE_ERROR: 2001,
  CACHE_ERROR: 2002,
  INTERNAL_ERROR: 2003,
} as const

export type ApiCodeValue = (typeof ApiCode)[keyof typeof ApiCode]

/** A paginated list plus its metadata. */
export interface Paginated<TItem> {
  items: TItem[]
  meta: ApiMeta
}
