import { Alert } from '@/components/ui/Alert'
import { Badge } from '@/components/ui/Badge'
import { PageHeader } from '@/components/ui/PageHeader'
import { Pagination } from '@/components/ui/Pagination'
import { SearchBar } from '@/components/ui/SearchBar'
import { Spinner } from '@/components/ui/Spinner'
import { useClientTable } from '@/hooks/useClientTable'
import { useLocationsQuery } from '@/hooks/useMasters'
import type { Location, LocationType } from '@/types/item'
import { LOCATION_TYPE_LABELS } from '@/types/item'

const matchesLocation = (location: Location, term: string) =>
  location.code.toLowerCase().includes(term) ||
  location.zone.toLowerCase().includes(term) ||
  location.type.toLowerCase().includes(term)

/** Daftar Warehouse — list view of rack/zone locations + "Tambah Location". */
export function WarehouseListPage() {
  const locationsQuery = useLocationsQuery()
  const table = useClientTable<Location>(locationsQuery.data ?? [], matchesLocation)

  return (
    <div className="flex flex-col gap-4">
      <PageHeader
        title="Daftar Warehouse"
        subtitle={`${table.meta.total} lokasi rak/zona terdaftar · tujuan penempatan dan penerimaan stok.`}
        action={{ to: '/warehouses/new', label: 'Tambah Location' }}
      />

      {locationsQuery.isError && (
        <Alert
          variant="error"
          title={locationsQuery.error.message}
          code={locationsQuery.error.code}
          errors={locationsQuery.error.fieldErrors}
        >
          Gagal memuat lokasi gudang.{' '}
          <button
            type="button"
            onClick={() => void locationsQuery.refetch()}
            className="font-semibold text-indigo-600 hover:underline"
          >
            Coba lagi
          </button>
        </Alert>
      )}

      <section className="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm">
        <SearchBar
          id="location-search"
          label="Cari lokasi"
          placeholder="Kode, zona atau tipe, misal: A-01 / rack"
          value={table.search}
          onChange={table.setSearch}
          busy={table.searchPending || locationsQuery.isFetching}
          onReset={table.reset}
        />

        <div className="relative overflow-x-auto">
          {locationsQuery.isLoading && (
            <div className="absolute inset-0 z-10 flex items-start justify-center bg-white/60 pt-10">
              <Spinner label="Memuat lokasi…" />
            </div>
          )}

          <table className="w-full min-w-[520px] border-collapse text-left text-sm">
            <thead>
              <tr className="border-b border-slate-200 bg-slate-50 text-xs uppercase tracking-wide text-slate-500">
                <th scope="col" className="px-4 py-3 font-semibold">ID</th>
                <th scope="col" className="px-4 py-3 font-semibold">Kode lokasi</th>
                <th scope="col" className="px-4 py-3 font-semibold">Zona</th>
                <th scope="col" className="px-4 py-3 text-right font-semibold">Jumlah item</th>
                <th scope="col" className="px-4 py-3 text-right font-semibold">Tipe</th>
              </tr>
            </thead>

            <tbody className="divide-y divide-slate-100">
              {table.rows.length === 0 && !locationsQuery.isLoading && (
                <tr>
                  <td colSpan={5} className="px-4 py-12 text-center">
                    <p className="text-sm font-medium text-slate-700">Tidak ada lokasi yang cocok</p>
                    <p className="mt-1 text-sm text-slate-500">
                      Ubah kata kunci, atau tambahkan lokasi baru.
                    </p>
                  </td>
                </tr>
              )}

              {table.rows.map((location) => (
                <tr key={location.id} className="hover:bg-slate-50/80">
                  <td className="whitespace-nowrap px-4 py-3 font-mono text-xs text-slate-500">
                    #{location.id}
                  </td>
                  <td className="whitespace-nowrap px-4 py-3">
                    <span className="font-mono text-xs font-semibold text-slate-900">
                      {location.code}
                    </span>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-slate-600">Zona {location.zone}</td>
                  <td className="whitespace-nowrap px-4 py-3 text-right">
                    <span className="font-semibold text-slate-900">{location.item_count}</span>
                    <span className="ml-1 text-xs text-slate-400">item</span>
                  </td>
                  <td className="whitespace-nowrap px-4 py-3 text-right">
                    <Badge tone="indigo">
                      {LOCATION_TYPE_LABELS[location.type as LocationType] ?? location.type}
                    </Badge>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <Pagination meta={table.meta} onPageChange={table.setPage} onLimitChange={table.setLimit} />
      </section>
    </div>
  )
}
