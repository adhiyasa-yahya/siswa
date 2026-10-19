import type { ApiEnvelope, ApiFieldError } from '@/types/api'

/** Base URL of the API, e.g. "/api/v1" (proxied) or "http://localhost:3069/api/v1". */
export const API_BASE_URL: string = import.meta.env.VITE_API_BASE_URL ?? '/api/v1'

/**
 * Error thrown for any non-successful response. It keeps the business code and
 * the per-field reasons so forms can highlight the offending inputs.
 */
export class ApiError extends Error {
  readonly status: number
  readonly code: number
  readonly fieldErrors: ApiFieldError[]

  constructor(message: string, status: number, code: number, fieldErrors: ApiFieldError[] = []) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.fieldErrors = fieldErrors
  }

  /** Field errors keyed by field name, ready to merge into form state. */
  get byField(): Record<string, string> {
    return this.fieldErrors.reduce<Record<string, string>>((acc, fieldError) => {
      acc[fieldError.field] = fieldError.reason
      return acc
    }, {})
  }
}

interface RequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  body?: unknown
  params?: Record<string, string | number | boolean | null | undefined>
  signal?: AbortSignal
}

function buildUrl(path: string, params?: RequestOptions['params']): string {
  const url = `${API_BASE_URL}${path}`
  if (!params) return url

  const search = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value === null || value === undefined || value === '') continue
    search.append(key, String(value))
  }

  const query = search.toString()
  return query ? `${url}?${query}` : url
}

/**
 * Performs a request and unwraps the standard envelope.
 * Returns the whole envelope so callers can also read `meta`.
 */
export async function request<TData>(
  path: string,
  { method = 'GET', body, params, signal }: RequestOptions = {},
): Promise<ApiEnvelope<TData>> {
  let response: Response

  try {
    response = await fetch(buildUrl(path, params), {
      method,
      signal,
      headers: body ? { 'Content-Type': 'application/json' } : undefined,
      body: body ? JSON.stringify(body) : undefined,
    })
  } catch (cause) {
    if (cause instanceof DOMException && cause.name === 'AbortError') throw cause
    throw new ApiError('Tidak dapat menghubungi server. Pastikan backend berjalan di port 3069.', 0, 99)
  }

  let envelope: ApiEnvelope<TData> | null = null
  try {
    envelope = (await response.json()) as ApiEnvelope<TData>
  } catch {
    envelope = null
  }

  if (!response.ok || !envelope?.success) {
    throw new ApiError(
      envelope?.message ?? `Request gagal dengan status ${response.status}`,
      response.status,
      envelope?.code ?? 99,
      envelope?.errors ?? [],
    )
  }

  return envelope
}

/** Unwraps `data` and fails loudly when the API returned an empty payload. */
export async function requestData<TData>(
  path: string,
  options: RequestOptions = {},
): Promise<TData> {
  const envelope = await request<TData>(path, options)
  if (envelope.data === undefined) {
    throw new ApiError(envelope.message || 'Response tanpa data', 200, envelope.code)
  }
  return envelope.data
}
