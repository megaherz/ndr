import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { TonConnectUIProvider } from '@tonconnect/ui-react'
import { useEffect, useState } from 'react'
import './App.css'

// Page components (will be created in subsequent tasks)
import Garage from './pages/Garage'
import Matchmaking from './pages/Matchmaking'
import Race from './pages/Race'
import Settlement from './pages/Settlement'
import GasStation from './pages/GasStation'

// Services
import { initializeTelegramWebApp } from './services/telegram/webapp'
import { extractTelegramInitData } from './services/auth/telegram'
import { loginWithTelegram } from './services/api/auth'
import { initializeEruda } from './services/debug/eruda'
import './services/debug' // Initialize debug utilities

// Stores
import { useAuthStore } from './stores/authStore'

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
  const [isInitializing, setIsInitializing] = useState(true)
  const { 
    isAuthenticated, 
    isTokenExpired, 
    setUser, 
    setTokens, 
    setLoading, 
    setError,
    logout 
  } = useAuthStore()

  // Automatic authentication on app launch
  useEffect(() => {
    const initializeApp = async () => {
      try {
        setIsInitializing(true)
        setLoading(true)
        setError(null)

        // Initialize Telegram WebApp first
        initializeTelegramWebApp()

        // Initialize Eruda debug console (only in dev/debug mode)
        await initializeEruda()

        // Wait a bit for Telegram WebApp to fully initialize
        await new Promise(resolve => setTimeout(resolve, 100))

        // Check if we have valid tokens
        if (isAuthenticated && !isTokenExpired()) {
          console.log('User already authenticated with valid tokens')
          setIsInitializing(false)
          setLoading(false)
          return
        }

        // Clear expired tokens
        if (isAuthenticated && isTokenExpired()) {
          console.log('Tokens expired, clearing auth state')
          logout()
        }

        // Extract Telegram initData for authentication
        const authPayload = extractTelegramInitData()
        
        if (!authPayload) {
          console.warn('No Telegram initData available - user needs to open app from Telegram')
          setError('Please open this app from Telegram')
          setIsInitializing(false)
          setLoading(false)
          return
        }

        // Attempt authentication with backend
        console.log('Attempting automatic authentication...')
        const { user, tokens } = await loginWithTelegram(authPayload)

        // Store authentication data
        setUser(user)
        setTokens(tokens)

        console.log('Automatic authentication successful')

      } catch (error) {
        console.error('Automatic authentication failed:', error)
        
        // Set user-friendly error message
        const errorMessage = error instanceof Error 
          ? error.message 
          : 'Authentication failed. Please try again.'
        
        setError(errorMessage)
        
        // Clear any stale auth data
        logout()

      } finally {
        setIsInitializing(false)
        setLoading(false)
      }
    }

    initializeApp()
  }, [isAuthenticated, isTokenExpired, logout, setError, setLoading, setTokens, setUser]) // Include all dependencies

  // Show loading screen during initialization
  if (isInitializing) {
    return (
      <div className="App">
        <div className="loading-screen">
          <div className="loading-spinner"></div>
          <p>Initializing Nitro Drag Royale...</p>
        </div>
      </div>
    )
  }

  // Show error screen if authentication failed
  if (!isAuthenticated && useAuthStore.getState().error) {
    return (
      <div className="App">
        <div className="error-screen">
          <h2>Authentication Error</h2>
          <p>{useAuthStore.getState().error}</p>
          <button 
            onClick={() => window.location.reload()}
            className="retry-button"
          >
            Retry
          </button>
        </div>
      </div>
    )
  }

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