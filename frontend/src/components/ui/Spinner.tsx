interface SpinnerProps {
  className?: string
  label?: string
}

/** Small inline loading indicator. */
export function Spinner({ className = '', label }: SpinnerProps) {
  return (
    <span className="inline-flex items-center gap-2" role="status" aria-live="polite">
      <svg
        className={`size-4 animate-spin ${className}`}
        viewBox="0 0 24 24"
        fill="none"
        aria-hidden="true"
      >
        <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
        <path
          className="opacity-90"
          fill="currentColor"
          d="M4 12a8 8 0 0 1 8-8V0C5.37 0 0 5.37 0 12h4z"
        />
      </svg>
      {label ? <span className="text-sm">{label}</span> : <span className="sr-only">Memuat</span>}
    </span>
  )
}
