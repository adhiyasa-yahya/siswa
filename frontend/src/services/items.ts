import { request, requestData } from '@/services/http'
import type { ApiMeta, Paginated } from '@/types/api'
import type { Item, ItemListParams, ItemPayload } from '@/types/item'

const EMPTY_META: ApiMeta = { page: 1, limit: 10, total: 0, total_pages: 0 }

/** GET /items - server side search, category filter and pagination. */
export async function listItems(
  params: ItemListParams,
  signal?: AbortSignal,
): Promise<Paginated<Item>> {
  const envelope = await request<Item[]>('/items', {
    params: {
      search: params.search,
      category_id: params.category_id ?? undefined,
      page: params.page,
      limit: params.limit,
      sort_by: params.sort_by,
      sort_dir: params.sort_dir,
    },
    signal,
  })

  return {
    items: envelope.data ?? [],
    meta: envelope.meta ?? { ...EMPTY_META, page: params.page, limit: params.limit },
  }
}

/** GET /items/:id */
export function getItem(id: number, signal?: AbortSignal): Promise<Item> {
  return requestData<Item>(`/items/${id}`, { signal })
}

/** POST /items */
export function createItem(payload: ItemPayload): Promise<Item> {
  return requestData<Item>('/items', { method: 'POST', body: payload })
}

/** PUT /items/:id */
export function updateItem(id: number, payload: ItemPayload): Promise<Item> {
  return requestData<Item>(`/items/${id}`, { method: 'PUT', body: payload })
}

/** DELETE /items/:id (soft delete) */
export function deleteItem(id: number): Promise<{ id: number; soft_deleted: boolean }> {
  return requestData<{ id: number; soft_deleted: boolean }>(`/items/${id}`, { method: 'DELETE' })
}
