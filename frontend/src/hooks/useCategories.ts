import { useQuery } from '@tanstack/react-query'

import type { ApiError } from '@/services/http'
import { listCategories, listLocations } from '@/services/masters'
import type { Category, Location } from '@/types/item'

export const masterKeys = {
  categories: ['categories'] as const,
  locations: ['locations'] as const,
}

/** Category master data for the filter dropdown and the item form. */
export function useCategoriesQuery() {
  return useQuery<Category[], ApiError>({
    queryKey: masterKeys.categories,
    queryFn: ({ signal }) => listCategories(signal),
    staleTime: 5 * 60_000, // master data barely changes
  })
}

/** Warehouse locations (used when receiving stock). */
export function useLocationsQuery(enabled = true) {
  return useQuery<Location[], ApiError>({
    queryKey: masterKeys.locations,
    queryFn: ({ signal }) => listLocations(signal),
    enabled,
    staleTime: 5 * 60_000,
  })
}
