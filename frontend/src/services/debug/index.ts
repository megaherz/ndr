// Debug utilities index
// Central export for all debugging utilities

export {
  erudaService,
  initializeEruda,
  showDebugConsole,
  hideDebugConsole,
  enableDebugMode,
  disableDebugMode,
  logAppInfo,
} from './eruda'

// Debug utilities for development
export const debugUtils = {
  // Quick console access
  console: () => {
    if ((window as any).NDR_DEBUG) {
      (window as any).NDR_DEBUG.show()
    } else {
      console.log('Debug console not available. Try adding ?debug=true to URL')
    }
  },

  // Log current auth state
  auth: () => {
    try {
      // Try to import the auth store dynamically
      import('../../stores/authStore').then(({ useAuthStore }) => {
        const authStore = useAuthStore.getState()
        console.group('🔐 Auth State')
        console.log('Authenticated:', authStore.isAuthenticated)
        console.log('User:', authStore.user)
        console.log('Tokens:', authStore.tokens ? 'Present' : 'None')
        console.log('Token Expired:', authStore.isTokenExpired?.())
        console.log('Loading:', authStore.isLoading)
        console.log('Error:', authStore.error)
        console.groupEnd()
      }).catch(() => {
        console.log('Auth store import failed')
      })
    } catch (error) {
      console.log('Auth store not available:', error)
    }
  },

  // Log Telegram WebApp state
  telegram: () => {
    const webApp = (window as any).Telegram?.WebApp
    if (webApp) {
      console.group('📱 Telegram WebApp')
      console.log('Version:', webApp.version)
      console.log('Platform:', webApp.platform)
      console.log('Color Scheme:', webApp.colorScheme)
      console.log('Viewport Height:', webApp.viewportHeight)
      console.log('User:', webApp.initDataUnsafe?.user)
      console.log('Init Data Length:', webApp.initData?.length || 0)
      console.groupEnd()
    } else {
      console.log('Telegram WebApp not available')
    }
  },

  // Log environment info
  env: () => {
    console.group('🌍 Environment')
    console.log('Mode:', import.meta.env.MODE)
    console.log('Dev:', import.meta.env.DEV)
    console.log('API URL:', import.meta.env.VITE_API_URL)
    console.log('Centrifugo URL:', import.meta.env.VITE_CENTRIFUGO_URL)
    console.log('Enable Eruda:', import.meta.env.VITE_ENABLE_ERUDA)
    console.log('App Version:', import.meta.env.VITE_APP_VERSION)
    console.groupEnd()
  },

  // Test API connectivity
  api: async () => {
    const baseUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'
    
    console.group('🌐 API Connectivity')
    console.log('API Base URL:', baseUrl)
    
    // Test different endpoints
    const endpoints = [
      { name: 'Health', url: `${baseUrl.replace('/api/v1', '')}/health` },
      { name: 'Auth (POST)', url: `${baseUrl}/auth/telegram`, method: 'POST' },
    ]
    
    for (const endpoint of endpoints) {
      try {
        const options: RequestInit = {
          method: endpoint.method || 'GET',
          headers: {
            'Content-Type': 'application/json',
          },
        }
        
        // For POST requests, add minimal body to avoid 400 errors
        if (endpoint.method === 'POST') {
          options.body = JSON.stringify({ test: true })
        }
        
        const response = await fetch(endpoint.url, options)
        console.log(`${endpoint.name}:`, response.ok ? '✅ OK' : '❌ Failed')
        console.log(`  Status: ${response.status}`)
        console.log(`  URL: ${endpoint.url}`)
        
        if (!response.ok) {
          const text = await response.text()
          console.log(`  Error: ${text}`)
        }
      } catch (error) {
        console.log(`${endpoint.name}: ❌ Network Error`)
        console.error(`  Error:`, error)
      }
    }
    
    console.groupEnd()
  },

  // Clear all storage
  clearStorage: () => {
    try {
      localStorage.clear()
      sessionStorage.clear()
      console.log('✅ Storage cleared')
    } catch (error) {
      console.error('❌ Failed to clear storage:', error)
    }
  },

  // Show all available debug commands
  help: () => {
    console.group('🛠 Debug Commands')
    console.log('NDR_DEBUG.show() - Show Eruda console')
    console.log('NDR_DEBUG.hide() - Hide Eruda console')
    console.log('NDR_DEBUG.enable() - Enable debug mode')
    console.log('NDR_DEBUG.disable() - Disable debug mode')
    console.log('NDR_DEBUG.info() - Log app info')
    console.log('')
    console.log('debugUtils.console() - Show debug console')
    console.log('debugUtils.auth() - Log auth state')
    console.log('debugUtils.telegram() - Log Telegram state')
    console.log('debugUtils.env() - Log environment')
    console.log('debugUtils.api() - Test API connectivity')
    console.log('debugUtils.clearStorage() - Clear all storage')
    console.log('debugUtils.help() - Show this help')
    console.groupEnd()
  },
}

// Make debug utils globally available in development
if (import.meta.env.DEV && typeof window !== 'undefined') {
  (window as any).debugUtils = debugUtils
  
  // Log helpful message
  console.log(
    '%c🛠 Debug Utils Available',
    'color: #5288c1; font-size: 14px; font-weight: bold;'
  )
  console.log(
    '%cType debugUtils.help() for available commands',
    'color: #708499; font-size: 12px;'
  )
}