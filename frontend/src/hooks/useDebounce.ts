import { useEffect, useState } from 'react'

/**
 * Returns `value` after it has stayed unchanged for `delay` ms.
 * Used by the item search input (300 ms) so typing does not spam the API.
 */
export function useDebounce<T>(value: T, delay = 300): T {
  const [debounced, setDebounced] = useState<T>(value)

  useEffect(() => {
    const timer = window.setTimeout(() => setDebounced(value), delay)
    return () => window.clearTimeout(timer)
  }, [value, delay])

  return debounced
}
