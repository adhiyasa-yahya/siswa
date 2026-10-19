import { QueryClient } from '@tanstack/react-query'

import { ApiError } from '@/services/http'

/**
 * Shared TanStack Query client.
 * Client side faults (4xx) are never retried: they will not succeed on a retry.
 */
export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: (failureCount, error) => {
        if (error instanceof ApiError && error.status >= 400 && error.status < 500) return false
        return failureCount < 2
      },
      refetchOnWindowFocus: false,
      staleTime: 30_000,
    },
    mutations: {
      retry: false,
    },
  },
})
