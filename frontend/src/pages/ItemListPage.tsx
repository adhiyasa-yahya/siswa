import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'

import { ItemFilters } from '@/components/items/ItemFilters'
import { ItemTable } from '@/components/items/ItemTable'
import { Alert } from '@/components/ui/Alert'
import { Pagination } from '@/components/ui/Pagination'
import { useCategoriesQuery } from '@/hooks/useCategories'
import { useDebounce } from '@/hooks/useDebounce'
import { useDeleteItem, useItemsQuery } from '@/hooks/useItems'
import type { Item, ItemListParams } from '@/types/item'

const SEARCH_DEBOUNCE_MS = 300

/**
 * Page 1 - warehouse item grid.
 * Search by name/SKU (debounced 300 ms), category filter and server side
 * pagination; every filter change resets to page 1.
 */
export function ItemListPage() {
  const [search, setSearch] = useState('')
  const [categoryId, setCategoryId] = useState<number | null>(null)
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(10)
  const [notice, setNotice] = useState<string | null>(null)

  const debouncedSearch = useDebounce(search, SEARCH_DEBOUNCE_MS)

  // A new search term or category always starts from the first page.
  useEffect(() => {
    setPage(1)
  }, [debouncedSearch, categoryId, limit])

  const params = useMemo<ItemListParams>(
    () => ({
      search: debouncedSearch.trim() || undefined,
      category_id: categoryId,
      page,
      limit,
      sort_by: 'created_at',
      sort_dir: 'desc',
    }),
    [debouncedSearch, categoryId, page, limit],
  )

  const itemsQuery = useItemsQuery(params)
  const categoriesQuery = useCategoriesQuery()
  const deleteMutation = useDeleteItem()

  const meta = itemsQuery.data?.meta ?? { page, limit, total: 0, total_pages: 0 }
  const items = itemsQuery.data?.items ?? []
  const debouncePending = search !== debouncedSearch

  const handleDelete = (item: Item) => {
    setNotice(null)
    deleteMutation.mutate(item.id, {
      onSuccess: () => setNotice(`Item ${item.sku} berhasil dihapus (soft delete).`),
    })
  }

  const handleReset = () => {
    setSearch('')
    setCategoryId(null)
    setPage(1)
  }

  return (
    <div className="flex flex-col gap-4">
      <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <h1 className="text-xl font-bold text-slate-900">Daftar Item Gudang</h1>
          <p className="text-sm text-slate-500">
            {meta.total} item terdaftar &middot; cari, filter kategori, lalu kelola datanya.
          </p>
        </div>

        <Link
          to="/items/new"
          className="inline-flex items-center justify-center gap-2 rounded-lg bg-indigo-600 px-3.5 py-2 text-sm font-semibold text-white shadow-sm transition-colors hover:bg-indigo-700"
        >
          <svg viewBox="0 0 24 24" className="size-4" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M12 5v14M5 12h14" strokeLinecap="round" />
          </svg>
          Tambah Item
        </Link>
      </div>

      {notice && (
        <Alert variant="success" title={notice} onDismiss={() => setNotice(null)} />
      )}

      {deleteMutation.isError && (
        <Alert
          variant="error"
          title={deleteMutation.error.message}
          code={deleteMutation.error.code}
          errors={deleteMutation.error.fieldErrors}
          onDismiss={() => deleteMutation.reset()}
        />
      )}

      {itemsQuery.isError && (
        <Alert
          variant="error"
          title={itemsQuery.error.message}
          code={itemsQuery.error.code}
          errors={itemsQuery.error.fieldErrors}
        >
          Gagal memuat daftar item.{' '}
          <button
            type="button"
            onClick={() => void itemsQuery.refetch()}
            className="font-semibold text-indigo-600 hover:underline"
          >
            Coba lagi
          </button>
        </Alert>
      )}

      {categoriesQuery.isError && (
        <Alert variant="warning" title="Daftar kategori gagal dimuat" code={categoriesQuery.error.code}>
          Filter kategori dinonaktifkan sementara.
        </Alert>
      )}

      <section className="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
        <ItemFilters
          search={search}
          onSearchChange={setSearch}
          categoryId={categoryId}
          onCategoryChange={setCategoryId}
          categories={categoriesQuery.data ?? []}
          categoriesLoading={categoriesQuery.isLoading}
          searching={debouncePending || itemsQuery.isFetching}
          onReset={handleReset}
        />

        <ItemTable
          items={items}
          loading={itemsQuery.isLoading}
          deletingId={deleteMutation.isPending ? deleteMutation.variables ?? null : null}
          onDelete={handleDelete}
        />

        <Pagination
          meta={meta}
          onPageChange={setPage}
          onLimitChange={setLimit}
          disabled={itemsQuery.isFetching}
        />
      </section>
    </div>
  )
}
