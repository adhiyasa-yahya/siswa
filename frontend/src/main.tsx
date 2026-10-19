import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClientProvider } from '@tanstack/react-query'
import { BrowserRouter } from 'react-router-dom'

import App from '@/App'
import { BackendGate } from '@/components/system/BackendGate'
import { queryClient } from '@/lib/queryClient'
import '@/index.css'

const container = document.getElementById('root')
if (!container) throw new Error('Root element #root not found')

createRoot(container).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      {/* Nothing renders until the backend has answered at least once. */}
      <BackendGate>
        <BrowserRouter>
          <App />
        </BrowserRouter>
      </BackendGate>
    </QueryClientProvider>
  </StrictMode>,
)
