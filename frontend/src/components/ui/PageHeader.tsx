import type { ReactNode } from 'react'

import { ActionLink } from '@/components/ui/ActionLink'

interface PageHeaderProps {
  title: string
  subtitle?: ReactNode
  /** Primary call to action, rendered as a link button on the right. */
  action?: { to: string; label: string }
}

/** Title + subtitle + primary action, shared by the three list pages. */
export function PageHeader({ title, subtitle, action }: PageHeaderProps) {
  return (
    <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <h1 className="text-xl font-bold text-slate-900">{title}</h1>
        {subtitle && <p className="text-sm text-slate-500">{subtitle}</p>}
      </div>

      {action && (
        <ActionLink to={action.to}>
          <svg viewBox="0 0 24 24" className="size-4" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M12 5v14M5 12h14" strokeLinecap="round" />
          </svg>
          {action.label}
        </ActionLink>
      )}
    </div>
  )
}
