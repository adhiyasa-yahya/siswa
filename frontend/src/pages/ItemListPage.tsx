import { useEffect, useMemo, useState } from 'react'

import { ItemFilters } from '@/components/items/ItemFilters'
import { ItemTable } from '@/components/items/ItemTable'
import { Alert } from '@/components/ui/Alert'
import { PageHeader } from '@/components/ui/PageHeader'
import { Pagination } from '@/components/ui/Pagination'
import { useCategoriesQuery } from '@/hooks/useMasters'
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
      <PageHeader
        title="Daftar Item"
        subtitle={`${meta.total} item terdaftar · cari, filter kategori, lalu kelola datanya.`}
        action={{ to: '/items/new', label: 'Tambah Item' }}
      />

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
