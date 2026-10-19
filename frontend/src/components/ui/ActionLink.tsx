import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'

import { useBackendOffline } from '@/hooks/useBackendStatus'

type Variant = 'primary' | 'outline'
type Size = 'sm' | 'md'

const VARIANTS: Record<Variant, string> = {
  primary: 'bg-indigo-600 text-white shadow-sm hover:bg-indigo-700',
  outline: 'text-indigo-600 ring-1 ring-inset ring-indigo-200 hover:bg-indigo-50',
}

const DISABLED: Record<Variant, string> = {
  primary: 'bg-indigo-300 text-white',
  outline: 'text-slate-400 ring-1 ring-inset ring-slate-200',
}

const SIZES: Record<Size, string> = {
  sm: 'px-2.5 py-1.5 text-xs',
  md: 'px-3.5 py-2 text-sm',
}

interface ActionLinkProps {
  to: string
  children: ReactNode
  variant?: Variant
  size?: Size
  className?: string
}

/**
 * A navigation link that performs an action (create / edit).
 *
 * While the backend is unreachable it renders as an inert, disabled-looking
 * span: the target page could not load or submit anything anyway, so sending
 * the user there would only produce a dead form.
 */
export function ActionLink({
  to,
  children,
  variant = 'primary',
  size = 'md',
  className = '',
}: ActionLinkProps) {
  const offline = useBackendOffline()

  const base = [
    'inline-flex items-center justify-center gap-2 rounded-lg font-semibold transition-colors',
    SIZES[size],
    className,
  ].join(' ')

  if (offline) {
    return (
      <span
        aria-disabled="true"
        title="Tidak tersedia: server sedang tidak dapat dihubungi"
        className={`${base} ${DISABLED[variant]} cursor-not-allowed`}
      >
        {children}
      </span>
    )
  }

  return (
    <Link to={to} className={`${base} ${VARIANTS[variant]}`}>
      {children}
    </Link>
  )
}
