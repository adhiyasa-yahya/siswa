import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import type { ApiError } from '@/services/http'
import {
  createCategory,
  createLocation,
  listCategories,
  listLocations,
} from '@/services/masters'
import type { Category, CategoryPayload, Location, LocationPayload } from '@/types/item'

export const masterKeys = {
  categories: (status: 'active' | 'all') => ['categories', status] as const,
  allCategories: ['categories'] as const,
  locations: ['locations'] as const,
}

/**
 * Category master data.
 * `active` feeds the filter/form dropdowns, `all` feeds the category list page.
 */
export function useCategoriesQuery(status: 'active' | 'all' = 'active') {
  return useQuery<Category[], ApiError>({
    queryKey: masterKeys.categories(status),
    queryFn: ({ signal }) => listCategories(status, signal),
    staleTime: 5 * 60_000, // master data barely changes
  })
}

/** POST /categories */
export function useCreateCategory() {
  const queryClient = useQueryClient()

  return useMutation<Category, ApiError, CategoryPayload>({
    mutationFn: createCategory,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: masterKeys.allCategories })
    },
  })
}

/** Warehouse locations (racks / zones). */
export function useLocationsQuery(enabled = true) {
  return useQuery<Location[], ApiError>({
    queryKey: masterKeys.locations,
    queryFn: ({ signal }) => listLocations(signal),
    enabled,
    staleTime: 5 * 60_000,
  })
}

/** POST /locations */
export function useCreateLocation() {
  const queryClient = useQueryClient()

  return useMutation<Location, ApiError, LocationPayload>({
    mutationFn: createLocation,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: masterKeys.locations })
    },
  })
}
