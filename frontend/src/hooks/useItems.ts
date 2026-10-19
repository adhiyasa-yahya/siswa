import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { createItem, deleteItem, getItem, listItems, updateItem } from '@/services/items'
import type { ApiError } from '@/services/http'
import type { Paginated } from '@/types/api'
import type { Item, ItemListParams, ItemPayload } from '@/types/item'

/** Query keys of the item resource, kept in one place for cache invalidation. */
export const itemKeys = {
  all: ['items'] as const,
  list: (params: ItemListParams) => ['items', 'list', params] as const,
  detail: (id: number) => ['items', 'detail', id] as const,
}

/** GET /items with search, filter and pagination. */
export function useItemsQuery(params: ItemListParams) {
  return useQuery<Paginated<Item>, ApiError>({
    queryKey: itemKeys.list(params),
    queryFn: ({ signal }) => listItems(params, signal),
    placeholderData: (previous) => previous, // keeps the table visible while refetching
    staleTime: 10_000,
  })
}

/** GET /items/:id - skipped while `id` is undefined (create mode). */
export function useItemQuery(id?: number) {
  return useQuery<Item, ApiError>({
    queryKey: itemKeys.detail(id ?? 0),
    queryFn: ({ signal }) => getItem(id as number, signal),
    enabled: typeof id === 'number' && Number.isFinite(id) && id > 0,
    staleTime: 0,
  })
}

/** POST /items */
export function useCreateItem() {
  const queryClient = useQueryClient()

  return useMutation<Item, ApiError, ItemPayload>({
    mutationFn: createItem,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: itemKeys.all })
    },
  })
}

/** PUT /items/:id */
export function useUpdateItem(id: number) {
  const queryClient = useQueryClient()

  return useMutation<Item, ApiError, ItemPayload>({
    mutationFn: (payload) => updateItem(id, payload),
    onSuccess: (item) => {
      queryClient.setQueryData(itemKeys.detail(id), item)
      void queryClient.invalidateQueries({ queryKey: itemKeys.all })
    },
  })
}

/** DELETE /items/:id (soft delete) */
export function useDeleteItem() {
  const queryClient = useQueryClient()

  return useMutation<{ id: number; soft_deleted: boolean }, ApiError, number>({
    mutationFn: deleteItem,
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: itemKeys.all })
    },
  })
}
