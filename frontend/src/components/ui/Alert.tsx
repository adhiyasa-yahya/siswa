import type { ReactNode } from 'react'

import type { ApiFieldError } from '@/types/api'

type Variant = 'success' | 'error' | 'info' | 'warning'

const VARIANTS: Record<Variant, { box: string; title: string; icon: string }> = {
  success: {
    box: 'border-emerald-200 bg-emerald-50',
    title: 'text-emerald-800',
    icon: 'M5 13l4 4L19 7',
  },
  error: {
    box: 'border-rose-200 bg-rose-50',
    title: 'text-rose-800',
    icon: 'M12 9v4m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z',
  },
  info: {
    box: 'border-indigo-200 bg-indigo-50',
    title: 'text-indigo-800',
    icon: 'M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z',
  },
  warning: {
    box: 'border-amber-200 bg-amber-50',
    title: 'text-amber-800',
    icon: 'M12 9v4m0 4h.01M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z',
  },
}

interface AlertProps {
  variant: Variant
  title: string
  /** Business code from the API envelope, shown for traceability. */
  code?: number
  /** Per-field reasons returned by the API. */
  errors?: ApiFieldError[]
  children?: ReactNode
  onDismiss?: () => void
}

/** Renders an API success/error message, including its field level reasons. */
export function Alert({ variant, title, code, errors, children, onDismiss }: AlertProps) {
  const styles = VARIANTS[variant]

  return (
    <div className={`flex gap-3 rounded-lg border p-3 ${styles.box}`} role="alert">
      <svg
        className={`mt-0.5 size-5 shrink-0 ${styles.title}`}
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        aria-hidden="true"
      >
        <path d={styles.icon} />
      </svg>

      <div className="flex-1 text-sm">
        <p className={`font-semibold ${styles.title}`}>
          {title}
          {typeof code === 'number' && (
            <span className="ml-2 font-mono text-xs font-normal opacity-70">code: {code}</span>
          )}
        </p>

        {children && <div className="mt-1 text-slate-700">{children}</div>}

        {errors && errors.length > 0 && (
          <ul className="mt-1.5 list-inside list-disc space-y-0.5 text-slate-700">
            {errors.map((fieldError) => (
              <li key={`${fieldError.field}-${fieldError.reason}`}>
                <span className="font-mono text-xs font-semibold">{fieldError.field}</span>
                {': '}
                {fieldError.reason}
              </li>
            ))}
          </ul>
        )}
      </div>

      {onDismiss && (
        <button
          type="button"
          onClick={onDismiss}
          className="shrink-0 rounded p-1 text-slate-500 hover:bg-white/60 hover:text-slate-700"
          aria-label="Tutup notifikasi"
        >
          <svg viewBox="0 0 24 24" className="size-4" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M6 6l12 12M18 6L6 18" strokeLinecap="round" />
          </svg>
        </button>
      )}
    </div>
  )
}
