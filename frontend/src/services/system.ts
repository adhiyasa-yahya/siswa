import { requestData } from '@/services/http'

export interface BackendStatusPayload {
  app: string
  version: string
  /** ok = fully usable, degraded = Redis down but the API still works. */
  status: 'ok' | 'degraded' | 'down'
  usable: boolean
  postgres: 'up' | 'down'
  redis: 'up' | 'down' | 'disabled'
  uptime_seconds: number
  server_time: string
}

/** How long a probe may take before the backend counts as unreachable. */
const PROBE_TIMEOUT_MS = 4000

/**
 * GET /status — the readiness probe behind the connection gate.
 *
 * It carries its own timeout: without one a hanging TCP connection would leave
 * the gate "checking" forever instead of reporting the backend as down.
 */
export function getBackendStatus(signal?: AbortSignal): Promise<BackendStatusPayload> {
  const timeout = AbortSignal.timeout(PROBE_TIMEOUT_MS)
  const combined =
    signal && 'any' in AbortSignal
      ? AbortSignal.any([signal, timeout])
      : timeout

  return requestData<BackendStatusPayload>('/status', { signal: combined })
}
