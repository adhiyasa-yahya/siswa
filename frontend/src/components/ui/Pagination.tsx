import { Button } from '@/components/ui/Button'
import type { ApiMeta } from '@/types/api'

interface PaginationProps {
  meta: ApiMeta
  onPageChange: (page: number) => void
  onLimitChange: (limit: number) => void
  disabled?: boolean
}

const LIMIT_OPTIONS = [10, 25, 50, 100]

/** Builds a compact page list: 1 … 4 5 [6] 7 8 … 20 */
function pageWindow(current: number, totalPages: number): (number | 'gap')[] {
  if (totalPages <= 7) {
    return Array.from({ length: totalPages }, (_, index) => index + 1)
  }

  const pages = new Set<number>([1, totalPages, current])
  for (const offset of [-2, -1, 1, 2]) {
    const page = current + offset
    if (page > 1 && page < totalPages) pages.add(page)
  }

  const sorted = [...pages].sort((a, b) => a - b)
  const withGaps: (number | 'gap')[] = []
  sorted.forEach((page, index) => {
    if (index > 0 && page - (sorted[index - 1] as number) > 1) withGaps.push('gap')
    withGaps.push(page)
  })
  return withGaps
}

export function Pagination({ meta, onPageChange, onLimitChange, disabled }: PaginationProps) {
  const totalPages = Math.max(meta.total_pages, 1)
  const from = meta.total === 0 ? 0 : (meta.page - 1) * meta.limit + 1
  const to = Math.min(meta.page * meta.limit, meta.total)

  return (
    <div className="flex flex-col gap-3 border-t border-slate-200 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
      <div className="flex items-center gap-3 text-sm text-slate-600">
        <span>
          Menampilkan <span className="font-semibold text-slate-900">{from}</span>–
          <span className="font-semibold text-slate-900">{to}</span> dari{' '}
          <span className="font-semibold text-slate-900">{meta.total}</span> item
        </span>

        <label className="flex items-center gap-1.5">
          <span className="sr-only">Jumlah baris per halaman</span>
          <select
            value={meta.limit}
            disabled={disabled}
            onChange={(event) => onLimitChange(Number(event.target.value))}
            className="rounded-md border border-slate-300 bg-white py-1 pl-2 pr-7 text-xs text-slate-700 focus:border-indigo-500 focus:outline focus:outline-2 focus:outline-indigo-500"
          >
            {LIMIT_OPTIONS.map((option) => (
              <option key={option} value={option}>
                {option} / halaman
              </option>
            ))}
          </select>
        </label>
      </div>

      <nav className="flex items-center gap-1" aria-label="Navigasi halaman">
        <Button
          variant="secondary"
          size="sm"
          disabled={disabled || meta.page <= 1}
          onClick={() => onPageChange(meta.page - 1)}
        >
          Sebelumnya
        </Button>

        <ul className="hidden items-center gap-1 sm:flex">
          {pageWindow(meta.page, totalPages).map((page, index) =>
            page === 'gap' ? (
              <li key={`gap-${index}`} className="px-1.5 text-slate-400">
                &hellip;
              </li>
            ) : (
              <li key={page}>
                <button
                  type="button"
                  disabled={disabled}
                  onClick={() => onPageChange(page)}
                  aria-current={page === meta.page ? 'page' : undefined}
                  className={[
                    'min-w-8 rounded-md px-2 py-1.5 text-xs font-semibold',
                    page === meta.page
                      ? 'bg-indigo-600 text-white'
                      : 'text-slate-600 hover:bg-slate-100',
                  ].join(' ')}
                >
                  {page}
                </button>
              </li>
            ),
          )}
        </ul>

        <span className="px-2 text-xs text-slate-500 sm:hidden">
          {meta.page} / {totalPages}
        </span>

        <Button
          variant="secondary"
          size="sm"
          disabled={disabled || meta.page >= totalPages}
          onClick={() => onPageChange(meta.page + 1)}
        >
          Berikutnya
        </Button>
      </nav>
    </div>
  )
}
