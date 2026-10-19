import { useState } from 'react'

import { ItemLocations } from '@/components/items/ItemLocations'
import { ActionLink } from '@/components/ui/ActionLink'
import { Badge } from '@/components/ui/Badge'
import { Button } from '@/components/ui/Button'
import { Spinner } from '@/components/ui/Spinner'
import type { Item } from '@/types/item'

interface ItemTableProps {
  items: Item[]
  loading: boolean
  deletingId: number | null
  onDelete: (item: Item) => void
}

const COLUMNS = ['SKU', 'Nama item', 'Kategori', 'Satuan', 'Lokasi', 'Total stok', 'Aksi'] as const

/** Warehouse item grid with an inline (two step) delete confirmation. */
export function ItemTable({ items, loading, deletingId, onDelete }: ItemTableProps) {
  const [confirmId, setConfirmId] = useState<number | null>(null)

  return (
    <div className="relative overflow-x-auto">
      {loading && (
        <div className="absolute inset-0 z-10 flex items-start justify-center bg-white/60 pt-10">
          <Spinner label="Memuat data item…" />
        </div>
      )}

      <table className="w-full min-w-[860px] border-collapse text-left text-sm">
        <thead>
          <tr className="border-b border-slate-200 bg-slate-50 text-xs uppercase tracking-wide text-slate-500">
            {COLUMNS.map((column) => (
              <th
                key={column}
                scope="col"
                className={`px-4 py-3 font-semibold ${column === 'Aksi' ? 'text-right' : ''} ${
                  column === 'Total stok' ? 'text-right' : ''
                }`}
              >
                {column}
              </th>
            ))}
          </tr>
        </thead>

        <tbody className="divide-y divide-slate-100">
          {items.length === 0 && !loading && (
            <tr>
              <td colSpan={COLUMNS.length} className="px-4 py-12 text-center">
                <p className="text-sm font-medium text-slate-700">Tidak ada item yang cocok</p>
                <p className="mt-1 text-sm text-slate-500">
                  Ubah kata kunci / filter kategori, atau tambahkan item baru.
                </p>
              </td>
            </tr>
          )}

          {items.map((item) => (
            <tr key={item.id} className="hover:bg-slate-50/80">
              <td className="whitespace-nowrap px-4 py-3">
                <span className="font-mono text-xs font-semibold text-slate-900">{item.sku}</span>
              </td>

              <td className="px-4 py-3">
                <span className="font-medium text-slate-900">{item.name}</span>
                <span className="block text-xs text-slate-400">id #{item.id}</span>
              </td>

              <td className="whitespace-nowrap px-4 py-3">
                <Badge tone="indigo">{item.category_name}</Badge>
              </td>

              <td className="whitespace-nowrap px-4 py-3 text-slate-600">{item.unit}</td>

              <td className="px-4 py-3">
                <ItemLocations locations={item.locations} unit={item.unit} />
              </td>

              <td className="whitespace-nowrap px-4 py-3 text-right">
                <span className="font-semibold text-slate-900">{item.total_qty}</span>
                <span className="ml-1 text-xs text-slate-400">{item.unit}</span>
              </td>

              <td className="whitespace-nowrap px-4 py-3 text-right">
                {confirmId === item.id ? (
                  <span className="inline-flex items-center gap-2">
                    <span className="text-xs text-slate-600">Hapus item ini?</span>
                    <Button
                      size="sm"
                      variant="danger"
                      loading={deletingId === item.id}
                      onClick={() => {
                        onDelete(item)
                        setConfirmId(null)
                      }}
                    >
                      Ya, hapus
                    </Button>
                    <Button size="sm" variant="ghost" onClick={() => setConfirmId(null)}>
                      Batal
                    </Button>
                  </span>
                ) : (
                  <span className="inline-flex items-center gap-2">
                    <ActionLink to={`/items/${item.id}/edit`} variant="outline" size="sm">
                      Edit
                    </ActionLink>
                    <Button size="sm" variant="ghost" onClick={() => setConfirmId(item.id)}>
                      Hapus
                    </Button>
                  </span>
                )}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}
