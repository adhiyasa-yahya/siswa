import { useEffect, useMemo, useState } from 'react'

import { useDebounce } from '@/hooks/useDebounce'
import type { ApiMeta } from '@/types/api'

interface ClientTable<T> {
  /** Raw search input value (bind this to the input). */
  search: string
  setSearch: (value: string) => void
  /** True while the debounce window has not elapsed yet. */
  searchPending: boolean
  rows: T[]
  meta: ApiMeta
  setPage: (page: number) => void
  setLimit: (limit: number) => void
  reset: () => void
}

/**
 * Client side search + pagination for the small master data sets
 * (categories and warehouse locations are returned in full by the API).
 * Items use the server side equivalent, since there are 1 000+ of them.
 */
export function useClientTable<T>(
  data: T[],
  matches: (row: T, term: string) => boolean,
  initialLimit = 10,
): ClientTable<T> {
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [limit, setLimit] = useState(initialLimit)

  const debouncedSearch = useDebounce(search, 300)

  useEffect(() => {
    setPage(1)
  }, [debouncedSearch, limit])

  const filtered = useMemo(() => {
    const term = debouncedSearch.trim().toLowerCase()
    if (!term) return data
    return data.filter((row) => matches(row, term))
    // `matches` is a stable predicate defined at module level by the callers.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [data, debouncedSearch])

  const total = filtered.length
  const totalPages = Math.max(Math.ceil(total / limit), 1)
  const safePage = Math.min(page, totalPages)
  const rows = filtered.slice((safePage - 1) * limit, safePage * limit)

  return {
    search,
    setSearch,
    searchPending: search !== debouncedSearch,
    rows,
    meta: { page: safePage, limit, total, total_pages: total === 0 ? 0 : totalPages },
    setPage,
    setLimit,
    reset: () => {
      setSearch('')
      setPage(1)
    },
  }
}
