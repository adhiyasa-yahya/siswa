import type { ReactNode } from 'react'

import { Button } from '@/components/ui/Button'
import { useBackendOffline } from '@/hooks/useBackendStatus'
import { Spinner } from '@/components/ui/Spinner'

interface SearchBarProps {
  id: string
  label: string
  placeholder: string
  value: string
  onChange: (value: string) => void
  /** True while the 300 ms debounce or a refetch is still running. */
  busy?: boolean
  onReset: () => void
  /** Extra controls (e.g. a category dropdown) rendered next to the input. */
  children?: ReactNode
}

/** Debounced search input + optional filters, shared by the list pages. */
export function SearchBar({
  id,
  label,
  placeholder,
  value,
  onChange,
  busy,
  onReset,
  children,
}: SearchBarProps) {
  const offline = useBackendOffline()

  return (
    <div className="flex flex-col gap-3 border-b border-slate-200 p-4 sm:flex-row sm:items-end">
      <div className="flex-1">
        <label htmlFor={id} className="mb-1.5 block text-sm font-medium text-slate-700">
          {label}
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
            id={id}
            type="search"
            value={value}
            onChange={(event) => onChange(event.target.value)}
            placeholder={placeholder}
            autoComplete="off"
            disabled={offline}
            className="w-full rounded-lg border border-slate-300 bg-white py-2 pl-9 pr-10 text-sm shadow-sm placeholder:text-slate-400 focus:border-indigo-500 focus:outline focus:outline-2 focus:outline-indigo-500"
          />

          {busy && (
            <span className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400">
              <Spinner />
            </span>
          )}
        </div>
      </div>

      {children}

      <Button variant="secondary" onClick={onReset} disabled={value.trim() === '' && !children}>
        Reset
      </Button>
    </div>
  )
}
