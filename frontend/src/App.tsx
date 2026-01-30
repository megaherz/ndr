import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { TonConnectUIProvider } from '@tonconnect/ui-react'
import { useEffect } from 'react'
import './App.css'

// Page components (will be created in subsequent tasks)
import Garage from './pages/Garage'
import Matchmaking from './pages/Matchmaking'
import Race from './pages/Race'
import Settlement from './pages/Settlement'
import GasStation from './pages/GasStation'

// Services
import { initializeTelegramWebApp } from './services/telegram/webapp'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      retry: 3,
      retryDelay: (attemptIndex) => Math.min(1000 * 2 ** attemptIndex, 30000),
      staleTime: 5 * 60 * 1000, // 5 minutes
    },
    mutations: {
      retry: 1,
    },
  },
})

// TON Connect manifest URL (will be configured properly)
const tonConnectManifestUrl = 'https://your-domain.com/tonconnect-manifest.json'

function App() {
  useEffect(() => {
    // Initialize Telegram WebApp on mount
    initializeTelegramWebApp()
  }, [])

  return (
    <TonConnectUIProvider manifestUrl={tonConnectManifestUrl}>
      <QueryClientProvider client={queryClient}>
        <Router>
          <div className="App">
            <Routes>
              {/* Default route redirects to garage */}
              <Route path="/" element={<Navigate to="/garage" replace />} />
              
              {/* Main game flow */}
              <Route path="/garage" element={<Garage />} />
              <Route path="/matchmaking" element={<Matchmaking />} />
              <Route path="/race/:matchId" element={<Race />} />
              <Route path="/settlement/:matchId" element={<Settlement />} />
              
              {/* TON wallet integration */}
              <Route path="/gas-station" element={<GasStation />} />
              
              {/* Catch-all route */}
              <Route path="*" element={<Navigate to="/garage" replace />} />
            </Routes>
          </div>
        </Router>
      </QueryClientProvider>
    </TonConnectUIProvider>
  )
}

export default App