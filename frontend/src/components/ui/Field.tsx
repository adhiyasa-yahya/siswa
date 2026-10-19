import type { InputHTMLAttributes, ReactNode, SelectHTMLAttributes } from 'react'

import { useBackendOffline } from '@/hooks/useBackendStatus'

interface FieldShellProps {
  id: string
  label: string
  required?: boolean
  error?: string
  hint?: ReactNode
  children: ReactNode
}

/** Label + control + hint/error, shared by every form control. */
function FieldShell({ id, label, required, error, hint, children }: FieldShellProps) {
  return (
    <div className="flex flex-col gap-1.5">
      <label htmlFor={id} className="text-sm font-medium text-slate-700">
        {label}
        {required && <span className="ml-0.5 text-rose-600">*</span>}
      </label>
      {children}
      {error ? (
        <p id={`${id}-error`} className="text-xs font-medium text-rose-600">
          {error}
        </p>
      ) : (
        hint && <p className="text-xs text-slate-500">{hint}</p>
      )}
    </div>
  )
}

const controlClasses = (hasError: boolean) =>
  [
    'w-full rounded-lg border bg-white px-3 py-2 text-sm text-slate-900 shadow-sm',
    'placeholder:text-slate-400 focus:outline focus:outline-2 focus:outline-offset-0',
    hasError
      ? 'border-rose-400 focus:outline-rose-500'
      : 'border-slate-300 focus:border-indigo-500 focus:outline-indigo-500',
    'disabled:bg-slate-50 disabled:text-slate-500',
  ].join(' ')

interface InputProps extends Omit<InputHTMLAttributes<HTMLInputElement>, 'id'> {
  id: string
  label: string
  error?: string
  hint?: ReactNode
}

export function Input({ id, label, error, hint, required, disabled, className = '', ...rest }: InputProps) {
  // An outage freezes every form control, not just the submit buttons.
  const offline = useBackendOffline()

  return (
    <FieldShell id={id} label={label} required={required} error={error} hint={hint}>
      <input
        {...rest}
        id={id}
        required={required}
        disabled={disabled || offline}
        aria-invalid={error ? true : undefined}
        aria-describedby={error ? `${id}-error` : undefined}
        className={`${controlClasses(Boolean(error))} ${className}`}
      />
    </FieldShell>
  )
}

interface SelectProps extends Omit<SelectHTMLAttributes<HTMLSelectElement>, 'id'> {
  id: string
  label: string
  error?: string
  hint?: ReactNode
  children: ReactNode
}

export function Select({
  id,
  label,
  error,
  hint,
  required,
  disabled,
  className = '',
  children,
  ...rest
}: SelectProps) {
  const offline = useBackendOffline()

  return (
    <FieldShell id={id} label={label} required={required} error={error} hint={hint}>
      <select
        {...rest}
        id={id}
        required={required}
        disabled={disabled || offline}
        aria-invalid={error ? true : undefined}
        aria-describedby={error ? `${id}-error` : undefined}
        className={`${controlClasses(Boolean(error))} ${className}`}
      >
        {children}
      </select>
    </FieldShell>
  )
}
