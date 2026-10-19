import { Alert } from '@/components/ui/Alert'
import { Badge } from '@/components/ui/Badge'
import { PageHeader } from '@/components/ui/PageHeader'
import { Pagination } from '@/components/ui/Pagination'
import { SearchBar } from '@/components/ui/SearchBar'
import { Spinner } from '@/components/ui/Spinner'
import { useClientTable } from '@/hooks/useClientTable'
import { useCategoriesQuery } from '@/hooks/useMasters'
import type { Category } from '@/types/item'

const matchesCategory = (category: Category, term: string) =>
  category.name.toLowerCase().includes(term) || String(category.id) === term

/** Daftar Category — list view + "Tambah Category" button. */
export function CategoryListPage() {
  const categoriesQuery = useCategoriesQuery('all')
  const table = useClientTable<Category>(categoriesQuery.data ?? [], matchesCategory)

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        title="Daftar Category"
        subtitle={`${table.meta.total} kategori terdaftar · dipakai sebagai master data item.`}
        action={{ to: '/categories/new', label: 'Tambah Category' }}
      />

      {categoriesQuery.isError && (
        <Alert
          variant="error"
          title={categoriesQuery.error.message}
          code={categoriesQuery.error.code}
          errors={categoriesQuery.error.fieldErrors}
        >
          Gagal memuat kategori.{' '}
          <button
            type="button"
            onClick={() => void categoriesQuery.refetch()}
            className="font-semibold text-indigo-600 hover:underline"
          >
            Coba lagi
          </button>
        </Alert>
      )}

      <section className="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
        <SearchBar
          id="category-search"
          label="Cari kategori"
          placeholder="Nama kategori, misal: Sparepart"
          value={table.search}
          onChange={table.setSearch}
          busy={table.searchPending || categoriesQuery.isFetching}
          onReset={table.reset}
        />

        <div className="relative overflow-x-auto">
          {categoriesQuery.isLoading && (
            <div className="absolute inset-0 z-10 flex items-start justify-center bg-white/60 pt-10">
              <Spinner label="Memuat kategori…" />
            </div>
          )}

          <table className="w-full min-w-[480px] border-collapse text-left text-sm">
            <thead>
              <tr className="border-b border-slate-200 bg-slate-50 text-xs uppercase tracking-wide text-slate-500">
                <th scope="col" className="px-4 py-3 font-semibold">ID</th>
                <th scope="col" className="px-4 py-3 font-semibold">Nama kategori</th>
                <th scope="col" className="px-4 py-3 text-right font-semibold">Jumlah item</th>
                <th scope="col" className="px-4 py-3 text-right font-semibold">Status</th>
              </tr>
            </thead>

            <tbody className="divide-y divide-slate-100">
              {table.rows.length === 0 && !categoriesQuery.isLoading && (
                <tr>
                  <td colSpan={4} className="px-4 py-12 text-center">
                    <p className="text-sm font-medium text-slate-700">Tidak ada kategori yang cocok</p>
                    <p className="mt-1 text-sm text-slate-500">
                      Ubah kata kunci, atau tambahkan kategori baru.
                    </p>
                  </td>
                </tr>
              )}

              {table.rows.map((category) => (
                <tr key={category.id} className="hover:bg-slate-50/80">
                  <td className="whitespace-nowrap px-4 py-3 font-mono text-xs text-slate-500">
                    #{category.id}
                  </td>
                  <td className="px-4 py-3 font-medium text-slate-900">{category.name}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-right">
                    <span className="font-semibold text-slate-900">{category.item_count}</span>
                    <span className="ml-1 text-xs text-slate-400">item</span>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-right">
                    <Badge tone={category.status === 'active' ? 'emerald' : 'slate'}>
                      {category.status === 'active' ? 'Aktif' : 'Nonaktif'}
                    </Badge>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <Pagination meta={table.meta} onPageChange={table.setPage} onLimitChange={table.setLimit} />
      </section>
    </div>
  )
}
