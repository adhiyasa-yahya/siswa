import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'

import { Alert } from '@/components/ui/Alert'
import { Button } from '@/components/ui/Button'
import { Input, Select } from '@/components/ui/Field'
import { useCreateCategory } from '@/hooks/useMasters'

/** Tambah Category — client side validation, then POST /categories. */
export function CategoryFormPage() {
  const navigate = useNavigate()
  const mutation = useCreateCategory()

  const [name, setName] = useState('')
  const [status, setStatus] = useState<'active' | 'inactive'>('active')
  const [touched, setTouched] = useState(false)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  const trimmed = name.trim()
  const clientError =
    trimmed === ''
      ? 'Nama kategori wajib diisi'
      : trimmed.length < 2
        ? 'Nama kategori minimal 2 karakter'
        : trimmed.length > 100
          ? 'Nama kategori maksimal 100 karakter'
          : undefined

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setTouched(true)
    setSuccessMessage(null)
    if (clientError) return

    mutation.reset()
    mutation.mutate(
      { name: trimmed, status },
      {
        onSuccess: (category) => {
          setSuccessMessage(`Kategori ${category.name} berhasil ditambahkan.`)
          window.setTimeout(() => navigate('/categories'), 900)
        },
      },
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <div>
        <Link to="/categories" className="text-sm font-semibold text-indigo-600 hover:underline">
          &larr; Kembali ke daftar category
        </Link>
        <h1 className="mt-2 text-xl font-bold text-slate-900">Tambah Category</h1>
        <p className="text-sm text-slate-500">
          Nama kategori harus unik. Kategori nonaktif tidak bisa dipakai saat membuat item.
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
        <form onSubmit={handleSubmit} noValidate className="flex flex-col gap-5">
          <div className="grid gap-5 sm:grid-cols-2">
            <Input
              id="category-name"
              label="Nama kategori"
              required
              value={name}
              placeholder="Material Konstruksi"
              disabled={mutation.isPending}
              error={(touched && clientError) || mutation.error?.byField.name}
              onBlur={() => setTouched(true)}
              onChange={(event) => setName(event.target.value)}
            />

            <Select
              id="category-status"
              label="Status"
              required
              value={status}
              disabled={mutation.isPending}
              hint="Hanya kategori aktif yang muncul di dropdown item."
              error={mutation.error?.byField.status}
              onChange={(event) => setStatus(event.target.value as 'active' | 'inactive')}
            >
              <option value="active">Aktif</option>
              <option value="inactive">Nonaktif</option>
            </Select>
          </div>

          <div className="flex items-center justify-end gap-2 border-t border-slate-200 pt-4">
            <Button
              type="button"
              variant="secondary"
              onClick={() => navigate('/categories')}
              disabled={mutation.isPending}
            >
              Batal
            </Button>
            <Button type="submit" loading={mutation.isPending}>
              Simpan category
            </Button>
          </div>
        </form>
      </section>
    </div>
  )
}
