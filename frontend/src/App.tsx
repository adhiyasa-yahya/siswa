import { Navigate, Route, Routes } from 'react-router-dom'

import { AppShell } from '@/components/layout/AppShell'
import { ItemFormPage } from '@/pages/ItemFormPage'
import { ItemListPage } from '@/pages/ItemListPage'
import { NotFoundPage } from '@/pages/NotFoundPage'

/**
 * Route map (React Router v6):
 *   /items            -> page 1: item grid
 *   /items/new        -> page 2: create form
 *   /items/:id/edit   -> page 2: edit form
 */
export default function App() {
  return (
    <Routes>
      <Route element={<AppShell />}>
        <Route index element={<Navigate to="/items" replace />} />
        <Route path="/items" element={<ItemListPage />} />
        <Route path="/items/new" element={<ItemFormPage />} />
        <Route path="/items/:id/edit" element={<ItemFormPage />} />
        <Route path="*" element={<NotFoundPage />} />
      </Route>
    </Routes>
  )
}
