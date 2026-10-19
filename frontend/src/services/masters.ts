import { requestData } from '@/services/http'
import type { Category, Location } from '@/types/item'

/** GET /categories - used by the category dropdown. */
export function listCategories(signal?: AbortSignal): Promise<Category[]> {
  return requestData<Category[]>('/categories', { signal })
}

/** GET /locations - warehouse racks / zones. */
export function listLocations(signal?: AbortSignal): Promise<Location[]> {
  return requestData<Location[]>('/locations', { signal })
}
