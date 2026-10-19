import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'

import { Alert } from '@/components/ui/Alert'
import { Button } from '@/components/ui/Button'
import { Input, Select } from '@/components/ui/Field'
import { useCreateLocation } from '@/hooks/useMasters'
import type { LocationType } from '@/types/item'
import { LOCATION_TYPES, LOCATION_TYPE_LABELS } from '@/types/item'

interface FormValues {
  code: string
  zone: string
  type: LocationType
}

type FormErrors = Partial<Record<keyof FormValues, string>>

const CODE_PATTERN = /^[A-Za-z0-9._-]+$/

function validate(values: FormValues): FormErrors {
  const errors: FormErrors = {}

  const code = values.code.trim()
  if (!code) {
    errors.code = 'Kode lokasi wajib diisi'
  } else if (code.length < 2) {
    errors.code = 'Kode lokasi minimal 2 karakter'
  } else if (code.length > 32) {
    errors.code = 'Kode lokasi maksimal 32 karakter'
  } else if (!CODE_PATTERN.test(code)) {
    errors.code = 'Kode hanya boleh huruf, angka, titik, minus dan underscore'
  }

  const zone = values.zone.trim()
  if (!zone) {
    errors.zone = 'Zona wajib diisi'
  } else if (zone.length > 32) {
    errors.zone = 'Zona maksimal 32 karakter'
  }

  if (!values.type) errors.type = 'Tipe lokasi wajib dipilih'

  return errors
}

/** Tambah Location — client side validation, then POST /locations. */
export function WarehouseFormPage() {
  const navigate = useNavigate()
  const mutation = useCreateLocation()

  const [values, setValues] = useState<FormValues>({ code: '', zone: '', type: 'rack' })
  const [touched, setTouched] = useState<Partial<Record<keyof FormValues, boolean>>>({})
  const [submitAttempted, setSubmitAttempted] = useState(false)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  const clientErrors = validate(values)

  const errorFor = (field: keyof FormValues) => {
    if ((submitAttempted || touched[field]) && clientErrors[field]) return clientErrors[field]
    return mutation.error?.byField[field]
  }

  const update = (field: keyof FormValues, value: string) =>
    setValues((previous) => ({ ...previous, [field]: value }))

  const handleSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setSubmitAttempted(true)
    setSuccessMessage(null)
    if (Object.keys(clientErrors).length > 0) return

    mutation.reset()
    mutation.mutate(
      {
        code: values.code.trim().toUpperCase(),
        zone: values.zone.trim().toUpperCase(),
        type: values.type,
      },
      {
        onSuccess: (location) => {
          setSuccessMessage(`Lokasi ${location.code} berhasil ditambahkan.`)
          window.setTimeout(() => navigate('/warehouses'), 900)
        },
      },
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <div>
        <Link to="/warehouses" className="text-sm font-semibold text-indigo-600 hover:underline">
          &larr; Kembali ke daftar warehouse
        </Link>
        <h1 className="mt-2 text-xl font-bold text-slate-900">Tambah Location</h1>
        <p className="text-sm text-slate-500">
          Lokasi rak/zona gudang. Kode lokasi harus unik, misal <code>A-01-R03</code>.
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
              id="location-code"
              label="Kode lokasi"
              required
              value={values.code}
              placeholder="A-01-R03"
              disabled={mutation.isPending}
              hint="Format bebas, disimpan dalam huruf kapital."
              error={errorFor('code')}
              onBlur={() => setTouched((previous) => ({ ...previous, code: true }))}
              onChange={(event) => update('code', event.target.value)}
            />

            <Input
              id="location-zone"
              label="Zona"
              required
              value={values.zone}
              placeholder="A"
              disabled={mutation.isPending}
              error={errorFor('zone')}
              onBlur={() => setTouched((previous) => ({ ...previous, zone: true }))}
              onChange={(event) => update('zone', event.target.value)}
            />

            <Select
              id="location-type"
              label="Tipe lokasi"
              required
              value={values.type}
              disabled={mutation.isPending}
              error={errorFor('type')}
              onChange={(event) => update('type', event.target.value)}
            >
              {LOCATION_TYPES.map((type) => (
                <option key={type} value={type}>
                  {LOCATION_TYPE_LABELS[type]}
                </option>
              ))}
            </Select>
          </div>

          <div className="flex items-center justify-end gap-2 border-t border-slate-200 pt-4">
            <Button
              type="button"
              variant="secondary"
              onClick={() => navigate('/warehouses')}
              disabled={mutation.isPending}
            >
              Batal
            </Button>
            <Button type="submit" loading={mutation.isPending}>
              Simpan location
            </Button>
          </div>
        </form>
      </section>
    </div>
  )
}
