import { StrictMode, Suspense } from 'react'
import { createRoot } from 'react-dom/client'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ErrorBoundary } from '@/components/ErrorBoundary'
import App from './App'
import './index.css'
import { initTelegramApp } from './lib/telegram'
import { initializeTheme } from './hooks/useTheme'

// Initialize Telegram SDK
initTelegramApp()

// Initialize theme (respects localStorage preference, falls back to Telegram/system)
initializeTheme()

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60, // 1 minute
      retry: 1,
      retryDelay: 1000,
    },
  },
})

function LoadingFallback() {
  return (
    <div className="h-screen flex items-center justify-center bg-background">
      <div className="animate-spin rounded-full h-8 w-8 border-2 border-accent border-t-transparent" />
    </div>
  )
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ErrorBoundary>
      <QueryClientProvider client={queryClient}>
        <Suspense fallback={<LoadingFallback />}>
          <App />
        </Suspense>
      </QueryClientProvider>
    </ErrorBoundary>
  </StrictMode>,
)
