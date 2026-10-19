import { requestData } from '@/services/http'
import type { Category, CategoryPayload, Location, LocationPayload } from '@/types/item'

/** GET /categories — `all` also returns inactive categories. */
export function listCategories(status: 'active' | 'all' = 'active', signal?: AbortSignal): Promise<Category[]> {
  return requestData<Category[]>('/categories', { params: { status }, signal })
}

/** POST /categories */
export function createCategory(payload: CategoryPayload): Promise<Category> {
  return requestData<Category>('/categories', { method: 'POST', body: payload })
}

/** GET /locations — warehouse racks / zones. */
export function listLocations(signal?: AbortSignal): Promise<Location[]> {
  return requestData<Location[]>('/locations', { signal })
}

/** POST /locations */
export function createLocation(payload: LocationPayload): Promise<Location> {
  return requestData<Location>('/locations', { method: 'POST', body: payload })
}
