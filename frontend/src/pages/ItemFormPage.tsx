import { useMemo, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'

import { EMPTY_ITEM_FORM, ItemForm, type ItemFormValues } from '@/components/items/ItemForm'
import { Alert } from '@/components/ui/Alert'
import { Spinner } from '@/components/ui/Spinner'
import { useCategoriesQuery } from '@/hooks/useCategories'
import { useCreateItem, useItemQuery, useUpdateItem } from '@/hooks/useItems'
import type { ItemPayload } from '@/types/item'

/**
 * Page 2 - create / edit item form.
 * Client side validation runs first; the API response (success or error, with
 * its per-field reasons) is then surfaced above the form.
 */
export function ItemFormPage() {
  const params = useParams<{ id?: string }>()
  const navigate = useNavigate()

  const parsedId = params.id ? Number(params.id) : undefined
  const isEdit = typeof parsedId === 'number' && Number.isFinite(parsedId) && parsedId > 0
  const itemId = isEdit ? (parsedId as number) : undefined

  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  const categoriesQuery = useCategoriesQuery()
  const itemQuery = useItemQuery(itemId)
  const createMutation = useCreateItem()
  const updateMutation = useUpdateItem(itemId ?? 0)

  const mutation = isEdit ? updateMutation : createMutation

  const initialValues = useMemo<ItemFormValues>(() => {
    if (isEdit && itemQuery.data) {
      return {
        sku: itemQuery.data.sku,
        name: itemQuery.data.name,
        category_id: String(itemQuery.data.category_id),
        unit: itemQuery.data.unit,
      }
    }
    return EMPTY_ITEM_FORM
  }, [isEdit, itemQuery.data])

  const handleSubmit = (payload: ItemPayload) => {
    setSuccessMessage(null)
    mutation.reset()

    mutation.mutate(payload, {
      onSuccess: (item) => {
        setSuccessMessage(
          isEdit
            ? `Item ${item.sku} berhasil diperbarui.`
            : `Item ${item.sku} berhasil ditambahkan.`,
        )
        // Give the user a moment to read the message, then return to the list.
        window.setTimeout(() => navigate('/items'), 900)
      },
    })
  }

  if (isEdit && itemQuery.isLoading) {
    return (
      <div className="grid place-items-center rounded-xl border border-slate-200 bg-white p-12 shadow-sm">
        <Spinner label="Memuat data item…" />
      </div>
    )
  }

  if (isEdit && itemQuery.isError) {
    return (
      <div className="flex flex-col gap-4">
        <Alert
          variant="error"
          title={itemQuery.error.message}
          code={itemQuery.error.code}
          errors={itemQuery.error.fieldErrors}
        />
        <Link to="/items" className="text-sm font-semibold text-indigo-600 hover:underline">
          &larr; Kembali ke daftar item
        </Link>
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <div>
        <Link to="/items" className="text-sm font-semibold text-indigo-600 hover:underline">
          &larr; Kembali ke daftar item
        </Link>
        <h1 className="mt-2 text-xl font-bold text-slate-900">
          {isEdit ? 'Edit Item' : 'Tambah Item Baru'}
        </h1>
        <p className="text-sm text-slate-500">
          {isEdit
            ? `Perbarui master data item #${itemId}. SKU tetap harus unik.`
            : 'Lengkapi data item. SKU, nama, kategori dan satuan wajib diisi.'}
        </p>
      </div>

      {successMessage && <Alert variant="success" title={successMessage} code={0} />}

      {mutation.isError && (
        <Alert
          variant="error"
          title={mutation.error.message}
          code={mutation.error.code}
          errors={mutation.error.fieldErrors}
          onDismiss={() => mutation.reset()}
        />
      )}

      <section className="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
        <ItemForm
          mode={isEdit ? 'edit' : 'create'}
          initialValues={initialValues}
          categories={categoriesQuery.data ?? []}
          categoriesLoading={categoriesQuery.isLoading}
          submitting={mutation.isPending}
          apiFieldErrors={mutation.error?.byField ?? {}}
          onSubmit={handleSubmit}
          onCancel={() => navigate('/items')}
        />
      </section>
    </div>
  )
}
