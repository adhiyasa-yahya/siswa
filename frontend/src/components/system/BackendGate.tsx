import { useEffect, useMemo, useRef, useState } from 'react'
import type { ReactNode } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'

import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import { BackendStatusContext } from '@/hooks/useBackendStatus'
import type { BackendStatusValue } from '@/hooks/useBackendStatus'
import type { ApiError } from '@/services/http'
import { getBackendStatus } from '@/services/system'
import type { BackendStatusPayload } from '@/services/system'

/** Probe cadence: fast while the backend is down, relaxed heartbeat while it is up. */
const RETRY_INTERVAL_MS = 3000
const HEARTBEAT_INTERVAL_MS = 30000

export const backendStatusKey = ['backend', 'status'] as const

interface BackendGateProps {
  children: ReactNode
}

/**
 * First gate of the app: nothing renders until the backend has answered once.
 *
 * - cold start, never reached   → blocking screen, retrying every 3 s;
 * - outage during the session   → sticky alert banner, controls disabled
 *   (via BackendStatusContext), retrying every 3 s;
 * - back online                 → banner disappears and every query is
 *   invalidated so the screen shows fresh data.
 */
export function BackendGate({ children }: BackendGateProps) {
  const queryClient = useQueryClient()

  const query = useQuery<BackendStatusPayload, ApiError>({
    queryKey: backendStatusKey,
    queryFn: ({ signal }) => getBackendStatus(signal),
    // The probe is the thing that decides "up or down": never retry it
    // internally, each attempt must be visible as one tick of the gate.
    retry: false,
    refetchOnWindowFocus: true,
    refetchOnReconnect: true,
    refetchIntervalInBackground: false,
    gcTime: 0,
    staleTime: 0,
    refetchInterval: (q) => (q.state.error ? RETRY_INTERVAL_MS : HEARTBEAT_INTERVAL_MS),
  })

  const online = query.isSuccess
  const everOnline = useRef(false)
  if (online) everOnline.current = true

  const [failures, setFailures] = useState(0)
  useEffect(() => {
    if (query.isSuccess) setFailures(0)
    else if (query.isError) setFailures((previous) => previous + 1)
    // errorUpdatedAt changes on every failed probe, dataUpdatedAt on every good one.
  }, [query.isSuccess, query.isError, query.errorUpdatedAt, query.dataUpdatedAt])

  // Coming back online: drop every cached list so the user sees current data.
  const wasOnline = useRef(true)
  useEffect(() => {
    if (online && !wasOnline.current) {
      void queryClient.invalidateQueries()
    }
    wasOnline.current = online
  }, [online, queryClient])

  const value = useMemo<BackendStatusValue>(
    () => ({
      online,
      checking: query.isFetching,
      everOnline: everOnline.current,
      failures,
      status: query.data ?? null,
      reason: query.error?.message ?? null,
      recheck: () => void query.refetch(),
    }),
    [online, query, failures],
  )

  // --- cold start: nothing has ever loaded -------------------------------
  if (!everOnline.current) {
    return (
      <BackendStatusContext.Provider value={value}>
        {query.isError ? (
          <UnreachableScreen
            reason={query.error.message}
            checking={query.isFetching}
            failures={failures}
            onRetry={() => void query.refetch()}
          />
        ) : (
          <ConnectingScreen />
        )}
      </BackendStatusContext.Provider>
    )
  }

  // --- normal operation (with an outage banner when the backend drops) ----
  return (
    <BackendStatusContext.Provider value={value}>
      {!online && (
        <OfflineBanner
          reason={query.error?.message ?? null}
          checking={query.isFetching}
          failures={failures}
          onRetry={() => void query.refetch()}
        />
      )}
      {children}
    </BackendStatusContext.Provider>
  )
}

function ConnectingScreen() {
  return (
    <div className="grid min-h-dvh place-items-center bg-slate-50 px-4">
      <div className="flex flex-col items-center gap-3 text-center">
        <span className="grid size-12 place-items-center rounded-xl bg-indigo-600 text-sm font-black text-white">
          SW
        </span>
        <Spinner label="Menghubungkan ke SISWA…" />
      </div>
    </div>
  )
}

interface OutageProps {
  reason: string | null
  checking: boolean
  failures: number
  onRetry: () => void
}

/** Blocking screen shown when the backend has never answered in this session. */
function UnreachableScreen({ reason, checking, failures, onRetry }: OutageProps) {
  return (
    <div className="grid min-h-dvh place-items-center bg-slate-50 px-4">
      <div
        role="alert"
        aria-live="assertive"
        className="w-full max-w-lg rounded-xl border border-rose-200 bg-white p-6 shadow-sm"
      >
        <div className="flex items-start gap-3">
          <span className="grid size-10 shrink-0 place-items-center rounded-lg bg-rose-50 text-rose-600">
            <svg viewBox="0 0 24 24" className="size-5" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M12 9v4m0 4h.01" strokeLinecap="round" />
              <circle cx="12" cy="12" r="9" />
            </svg>
          </span>

          <div className="flex-1">
            <h1 className="text-lg font-bold text-slate-900">Server tidak dapat dihubungi</h1>
            <p className="mt-1 text-sm text-slate-600">
              Aplikasi SISWA tidak dapat digunakan untuk sementara karena backend sedang tidak
              merespons. Data tidak dapat dimuat maupun disimpan sampai koneksi pulih.
            </p>

            <OutageDetails reason={reason} checking={checking} failures={failures} />

            <div className="mt-4 flex items-center gap-2">
              <Button onClick={onRetry} loading={checking} allowOffline>
                Coba sekarang
              </Button>
              <span className="text-xs text-slate-500">
                Sistem mencoba otomatis setiap {RETRY_INTERVAL_MS / 1000} detik.
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

/** Sticky banner shown when the backend drops while the app is already open. */
function OfflineBanner({ reason, checking, failures, onRetry }: OutageProps) {
  return (
    <div
      role="alert"
      aria-live="assertive"
      className="sticky top-0 z-50 border-b border-rose-200 bg-rose-50"
    >
      <div className="mx-auto flex max-w-6xl flex-col gap-2 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-start gap-2">
          <svg
            viewBox="0 0 24 24"
            className="mt-0.5 size-5 shrink-0 text-rose-600"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
          >
            <path d="M12 9v4m0 4h.01" strokeLinecap="round" />
            <circle cx="12" cy="12" r="9" />
          </svg>

          <div className="text-sm">
            <p className="font-semibold text-rose-800">
              Koneksi ke server terputus — aplikasi tidak dapat digunakan sementara
            </p>
            <p className="text-rose-700">
              Semua tombol dan filter dinonaktifkan sampai server kembali online.
            </p>
            <OutageDetails reason={reason} checking={checking} failures={failures} />
          </div>
        </div>

        <Button variant="secondary" size="sm" onClick={onRetry} loading={checking} allowOffline>
          Coba sekarang
        </Button>
      </div>
    </div>
  )
}

function OutageDetails({ reason, checking, failures }: Omit<OutageProps, 'onRetry'>) {
  return (
    <p className="mt-1 flex flex-wrap items-center gap-x-2 gap-y-1 text-xs text-slate-500">
      {checking ? (
        <Spinner label="Mencoba menyambung kembali…" />
      ) : (
        <span>
          Percobaan berikutnya dalam {RETRY_INTERVAL_MS / 1000} detik
          {failures > 1 ? ` · ${failures} percobaan gagal` : ''}
        </span>
      )}
      {reason && <span className="font-mono">({reason})</span>}
    </p>
  )
}
