import { Button } from '@/components/ui/Button'
import { useBackendOffline } from '@/hooks/useBackendStatus'
import { Spinner } from '@/components/ui/Spinner'
import type { Category } from '@/types/item'

interface ItemFiltersProps {
  search: string
  onSearchChange: (value: string) => void
  categoryId: number | null
  onCategoryChange: (value: number | null) => void
  categories: Category[]
  categoriesLoading: boolean
  /** True while a debounced search is still pending or a refetch is running. */
  searching: boolean
  onReset: () => void
}

/**
 * Search + category filter bar. The input is uncontrolled by the query: the
 * page debounces the value (300 ms) before it reaches the API.
 */
export function ItemFilters({
  search,
  onSearchChange,
  categoryId,
  onCategoryChange,
  categories,
  categoriesLoading,
  searching,
  onReset,
}: ItemFiltersProps) {
  const offline = useBackendOffline()
  const hasFilter = search.trim() !== '' || categoryId !== null

  return (
    <div className="flex flex-col gap-3 border-b border-slate-200 p-4 sm:flex-row sm:items-end">
      <div className="flex-1">
        <label htmlFor="item-search" className="mb-1.5 block text-sm font-medium text-slate-700">
          Cari item
        </label>
        <div className="relative">
          <svg
            className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-slate-400"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
            aria-hidden="true"
          >
            <circle cx="11" cy="11" r="7" />
            <path d="M20 20l-3.5-3.5" strokeLinecap="round" />
          </svg>

          <input
            id="item-search"
            type="search"
            value={search}
            onChange={(event) => onSearchChange(event.target.value)}
            placeholder="Nama atau SKU, misal: hollow / SKU-00042"
            autoComplete="off"
            disabled={offline}
            className="w-full rounded-lg border border-slate-300 bg-white py-2 pl-9 pr-10 text-sm shadow-sm placeholder:text-slate-400 focus:border-indigo-500 focus:outline focus:outline-2 focus:outline-indigo-500"
          />

          {searching && (
            <span className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400">
              <Spinner />
            </span>
          )}
        </div>
      </div>

      <div className="sm:w-56">
        <label htmlFor="item-category" className="mb-1.5 block text-sm font-medium text-slate-700">
          Kategori
        </label>
        <select
          id="item-category"
          value={categoryId ?? ''}
          disabled={categoriesLoading || offline}
          onChange={(event) =>
            onCategoryChange(event.target.value === '' ? null : Number(event.target.value))
          }
          className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm shadow-sm focus:border-indigo-500 focus:outline focus:outline-2 focus:outline-indigo-500 disabled:bg-slate-50"
        >
          <option value="">{categoriesLoading ? 'Memuat kategori…' : 'Semua kategori'}</option>
          {categories.map((category) => (
            <option key={category.id} value={category.id}>
              {category.name}
            </option>
          ))}
        </select>
      </div>

      <Button variant="secondary" onClick={onReset} disabled={!hasFilter}>
        Reset
      </Button>
    </div>
  )
}
