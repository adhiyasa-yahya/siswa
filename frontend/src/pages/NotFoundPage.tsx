import { Link } from 'react-router-dom'

export function NotFoundPage() {
  return (
    <div className="grid place-items-center rounded-xl border border-slate-200 bg-white p-12 text-center shadow-sm">
      <p className="text-5xl font-black text-slate-200">404</p>
      <h1 className="mt-2 text-lg font-bold text-slate-900">Halaman tidak ditemukan</h1>
      <p className="mt-1 text-sm text-slate-500">
        Alamat yang kamu buka tidak tersedia di aplikasi SISWA.
      </p>
      <Link
        to="/items"
        className="mt-5 inline-flex items-center rounded-lg bg-indigo-600 px-3.5 py-2 text-sm font-semibold text-white hover:bg-indigo-700"
      >
        Ke daftar item
      </Link>
    </div>
  )
}
