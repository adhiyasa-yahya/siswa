import { Navigate, Route, Routes } from 'react-router-dom'

import { AppShell } from '@/components/layout/AppShell'
import { CategoryFormPage } from '@/pages/CategoryFormPage'
import { CategoryListPage } from '@/pages/CategoryListPage'
import { ItemFormPage } from '@/pages/ItemFormPage'
import { ItemListPage } from '@/pages/ItemListPage'
import { NotFoundPage } from '@/pages/NotFoundPage'
import { WarehouseFormPage } from '@/pages/WarehouseFormPage'
import { WarehouseListPage } from '@/pages/WarehouseListPage'

/**
 * Route map (React Router v6) — one list view + one create view per menu entry:
 *   /items            -> Daftar Item (grid)
 *   /items/new        -> Tambah Item (create form, with warehouse placement)
 *   /items/:id/edit   -> Edit Item
 *   /categories       -> Daftar Category (list)
 *   /categories/new   -> Tambah Category
 *   /warehouses       -> Daftar Warehouse (locations list)
 *   /warehouses/new   -> Tambah Location
 */
export default function App() {
  return (
    <Routes>
      <Route element={<AppShell />}>
        <Route index element={<Navigate to="/items" replace />} />

        <Route path="/items" element={<ItemListPage />} />
        <Route path="/items/new" element={<ItemFormPage />} />
        <Route path="/items/:id/edit" element={<ItemFormPage />} />

        <Route path="/categories" element={<CategoryListPage />} />
        <Route path="/categories/new" element={<CategoryFormPage />} />

        <Route path="/warehouses" element={<WarehouseListPage />} />
        <Route path="/warehouses/new" element={<WarehouseFormPage />} />

        <Route path="*" element={<NotFoundPage />} />
      </Route>
    </Routes>
  )
}
