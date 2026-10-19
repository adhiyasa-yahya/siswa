import { createContext, useContext } from 'react'

import type { BackendStatusPayload } from '@/services/system'

export interface BackendStatusValue {
  /** False whenever the last probe failed: the UI must stop accepting input. */
  online: boolean
  /** True while a probe is in flight. */
  checking: boolean
  /** True once a probe has succeeded at least once in this session. */
  everOnline: boolean
  /** Consecutive failed probes. */
  failures: number
  status: BackendStatusPayload | null
  /** Human readable reason for the outage, when there is one. */
  reason: string | null
  /** Forces an immediate probe ("Coba sekarang"). */
  recheck: () => void
}

const FALLBACK: BackendStatusValue = {
  online: true,
  checking: false,
  everOnline: true,
  failures: 0,
  status: null,
  reason: null,
  recheck: () => {},
}

/**
 * Connection state of the backend, provided by BackendGate.
 *
 * The default assumes "online" so a component rendered outside the gate (tests,
 * storybook) behaves normally instead of locking itself.
 */
export const BackendStatusContext = createContext<BackendStatusValue>(FALLBACK)

/** Reads the backend connection state; used to disable inputs during an outage. */
export function useBackendStatus(): BackendStatusValue {
  return useContext(BackendStatusContext)
}

/** Convenience helper for the common `disabled={!online}` case. */
export function useBackendOffline(): boolean {
  return !useContext(BackendStatusContext).online
}
