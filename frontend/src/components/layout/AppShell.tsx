import { NavLink, Outlet } from 'react-router-dom'

const NAV = [
  { to: '/items', label: 'Daftar Item' },
  { to: '/categories', label: 'Daftar Category' },
  { to: '/warehouses', label: 'Daftar Warehouse' },
]

/** Page frame: header with navigation, content outlet and footer. */
export function AppShell() {
  return (
    <div className="flex min-h-dvh flex-col bg-slate-50 text-slate-900">
      <header className="border-b border-slate-200 bg-white">
        <div className="mx-auto flex max-w-6xl flex-col gap-3 px-4 py-4 sm:flex-row sm:items-center sm:justify-between">
          <div className="flex items-center gap-3">
            <span className="grid size-10 place-items-center rounded-xl bg-indigo-600 text-sm font-black text-white">
              SW
            </span>
            <div>
              <p className="text-base font-bold leading-tight">SISWA</p>
              <p className="text-xs text-slate-500">Sistem Item Stok Warehouse &amp; Aset</p>
            </div>
          </div>

          <nav className="flex items-center gap-1" aria-label="Navigasi utama">
            {NAV.map((entry) => (
              <NavLink
                key={entry.to}
                to={entry.to}
                className={({ isActive }) =>
                  [
                    'rounded-lg px-3 py-2 text-sm font-semibold transition-colors',
                    isActive ? 'bg-indigo-50 text-indigo-700' : 'text-slate-600 hover:bg-slate-100',
                  ].join(' ')
                }
              >
                {entry.label}
              </NavLink>
            ))}
          </nav>
        </div>
      </header>

      <main className="mx-auto w-full max-w-6xl flex-1 px-4 py-6">
        <Outlet />
      </main>

      <footer className="border-t border-slate-200 bg-white">
        <div className="mx-auto flex max-w-6xl flex-col gap-1 px-4 py-4 text-xs text-slate-500 sm:flex-row sm:items-center sm:justify-between">
          <span>SISWA &middot; React 19 + TanStack Query &middot; Go + PostgreSQL + Redis</span>
          <a
            href="http://localhost:3069/swagger/index.html"
            target="_blank"
            rel="noreferrer"
            className="font-semibold text-indigo-600 hover:underline"
          >
            Dokumentasi API (Swagger)
          </a>
        </div>
      </footer>
    </div>
  )
}
