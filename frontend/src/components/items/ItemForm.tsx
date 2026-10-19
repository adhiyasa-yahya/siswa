import { useEffect, useMemo, useState } from 'react'
import type { FormEvent } from 'react'

import { Button } from '@/components/ui/Button'
import { Input, Select } from '@/components/ui/Field'
import type { Category, ItemPayload } from '@/types/item'
import { UNIT_OPTIONS } from '@/types/item'

export interface ItemFormValues {
  sku: string
  name: string
  category_id: string
  unit: string
}

export const EMPTY_ITEM_FORM: ItemFormValues = { sku: '', name: '', category_id: '', unit: '' }

type FormErrors = Partial<Record<keyof ItemFormValues, string>>

const SKU_PATTERN = /^[A-Za-z0-9._-]+$/

/** Client side validation: SKU, nama, kategori and unit are all mandatory. */
export function validateItemForm(values: ItemFormValues): FormErrors {
  const errors: FormErrors = {}

  const sku = values.sku.trim()
  if (!sku) {
    errors.sku = 'SKU wajib diisi'
  } else if (sku.length < 3) {
    errors.sku = 'SKU minimal 3 karakter'
  } else if (sku.length > 64) {
    errors.sku = 'SKU maksimal 64 karakter'
  } else if (!SKU_PATTERN.test(sku)) {
    errors.sku = 'SKU hanya boleh huruf, angka, titik, minus dan underscore'
  }

  const name = values.name.trim()
  if (!name) {
    errors.name = 'Nama item wajib diisi'
  } else if (name.length < 2) {
    errors.name = 'Nama item minimal 2 karakter'
  } else if (name.length > 160) {
    errors.name = 'Nama item maksimal 160 karakter'
  }

  if (!values.category_id) {
    errors.category_id = 'Kategori wajib dipilih'
  }

  const unit = values.unit.trim()
  if (!unit) {
    errors.unit = 'Satuan wajib diisi'
  } else if (unit.length > 24) {
    errors.unit = 'Satuan maksimal 24 karakter'
  }

  return errors
}

interface ItemFormProps {
  mode: 'create' | 'edit'
  initialValues: ItemFormValues
  categories: Category[]
  categoriesLoading: boolean
  submitting: boolean
  /** Field errors returned by the API (e.g. sku: "Duplicate entry"). */
  apiFieldErrors: Record<string, string>
  onSubmit: (payload: ItemPayload) => void
  onCancel: () => void
}

export function ItemForm({
  mode,
  initialValues,
  categories,
  categoriesLoading,
  submitting,
  apiFieldErrors,
  onSubmit,
  onCancel,
}: ItemFormProps) {
  const [values, setValues] = useState<ItemFormValues>(initialValues)
  const [touched, setTouched] = useState<Partial<Record<keyof ItemFormValues, boolean>>>({})
  const [submitAttempted, setSubmitAttempted] = useState(false)

  // Re-seed the form once the item being edited arrives from the API.
  useEffect(() => {
    setValues(initialValues)
    setTouched({})
    setSubmitAttempted(false)
  }, [initialValues])

  const clientErrors = useMemo(() => validateItemForm(values), [values])

  const errorFor = (field: keyof ItemFormValues): string | undefined => {
    if (submitAttempted || touched[field]) {
      if (clientErrors[field]) return clientErrors[field]
    }
    return apiFieldErrors[field]
  }

  const update = (field: keyof ItemFormValues) => (value: string) =>
    setValues((previous) => ({ ...previous, [field]: value }))

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setSubmitAttempted(true)

    if (Object.keys(clientErrors).length > 0) return

    onSubmit({
      sku: values.sku.trim().toUpperCase(),
      name: values.name.trim(),
      category_id: Number(values.category_id),
      unit: values.unit.trim(),
    })
  }

  return (
    <form onSubmit={handleSubmit} noValidate className="flex flex-col gap-5">
      <div className="grid gap-5 sm:grid-cols-2">
        <Input
          id="sku"
          label="SKU"
          required
          value={values.sku}
          placeholder="SKU-GUDANG-001"
          disabled={submitting}
          hint="Kode unik item. Duplikat akan ditolak server dengan kode 1003."
          error={errorFor('sku')}
          onBlur={() => setTouched((previous) => ({ ...previous, sku: true }))}
          onChange={(event) => update('sku')(event.target.value)}
        />

        <Input
          id="name"
          label="Nama item"
          required
          value={values.name}
          placeholder="Besi Hollow 4x4 Galvanis"
          disabled={submitting}
          error={errorFor('name')}
          onBlur={() => setTouched((previous) => ({ ...previous, name: true }))}
          onChange={(event) => update('name')(event.target.value)}
        />

        <Select
          id="category_id"
          label="Kategori"
          required
          value={values.category_id}
          disabled={submitting || categoriesLoading}
          error={errorFor('category_id')}
          onBlur={() => setTouched((previous) => ({ ...previous, category_id: true }))}
          onChange={(event) => update('category_id')(event.target.value)}
        >
          <option value="">{categoriesLoading ? 'Memuat kategori…' : 'Pilih kategori'}</option>
          {categories.map((category) => (
            <option key={category.id} value={category.id}>
              {category.name}
            </option>
          ))}
        </Select>

        <div className="flex flex-col gap-1.5">
          <Input
            id="unit"
            label="Satuan"
            required
            value={values.unit}
            placeholder="pcs"
            list="unit-options"
            disabled={submitting}
            hint="Pilih dari daftar atau tulis satuan sendiri."
            error={errorFor('unit')}
            onBlur={() => setTouched((previous) => ({ ...previous, unit: true }))}
            onChange={(event) => update('unit')(event.target.value)}
          />
          <datalist id="unit-options">
            {UNIT_OPTIONS.map((unit) => (
              <option key={unit} value={unit} />
            ))}
          </datalist>
        </div>
      </div>

      <div className="flex items-center justify-end gap-2 border-t border-slate-200 pt-4">
        <Button type="button" variant="secondary" onClick={onCancel} disabled={submitting}>
          Batal
        </Button>
        <Button type="submit" loading={submitting}>
          {mode === 'create' ? 'Simpan item' : 'Perbarui item'}
        </Button>
      </div>
    </form>
  )
}
