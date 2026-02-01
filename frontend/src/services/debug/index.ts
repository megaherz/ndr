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
    const ndrDebug = (window as unknown as { NDR_DEBUG?: { show: () => void } }).NDR_DEBUG
    if (ndrDebug) {
      ndrDebug.show()
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
    const telegram = (window as unknown as { Telegram?: { WebApp: unknown } }).Telegram
    const webApp = telegram?.WebApp as {
      version?: string
      platform?: string
      colorScheme?: string
      viewportHeight?: number
      initDataUnsafe?: { user?: unknown }
      initData?: string
      isExpanded?: boolean
      MainButton?: unknown
      BackButton?: unknown
    } | undefined
    
    if (webApp) {
      console.group('📱 Telegram WebApp')
      console.log('Version:', webApp.version)
      console.log('Platform:', webApp.platform)
      console.log('Color Scheme:', webApp.colorScheme)
      console.log('Viewport Height:', webApp.viewportHeight)
      console.log('Is Expanded:', webApp.isExpanded)
      console.log('User:', webApp.initDataUnsafe?.user)
      console.log('Init Data Length:', webApp.initData?.length || 0)
      
      if (webApp.initData) {
        console.log('Init Data Preview:', webApp.initData.substring(0, 200) + '...')
        
        // Try to parse the init data to see its structure
        try {
          const params = new URLSearchParams(webApp.initData)
          console.log('Init Data Parsed:')
          for (const [key, value] of params.entries()) {
            if (key === 'user' || key === 'chat') {
              try {
                console.log(`  ${key}:`, JSON.parse(value))
              } catch {
                console.log(`  ${key}:`, value)
              }
            } else {
              console.log(`  ${key}:`, value)
            }
          }
        } catch (error) {
          console.log('Failed to parse init data:', error)
        }
      }
      
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
      { name: 'Auth (POST)', url: `${baseUrl}/auth/telegram`, method: 'POST', requiresAuth: true },
    ]
    
    for (const endpoint of endpoints) {
      try {
        const options: RequestInit = {
          method: endpoint.method || 'GET',
          headers: {
            'Content-Type': 'application/json',
          },
        }
        
        // For auth endpoints, we need proper Telegram init data
        if (endpoint.requiresAuth) {
          const telegram = (window as unknown as { Telegram?: { WebApp: { initData?: string } } }).Telegram
          const initData = telegram?.WebApp?.initData
          
          if (initData) {
            console.log(`${endpoint.name}: 📤 Sending Telegram init data (length: ${initData.length})`)
            console.log(`  Init data preview: ${initData.substring(0, 100)}...`)
            options.body = JSON.stringify({ init_data: initData })
          } else {
            console.log(`${endpoint.name}: ⚠️ Skipped (no Telegram init data available)`)
            console.log(`  Note: This endpoint requires valid Telegram WebApp init data`)
            continue
          }
        } else if (endpoint.method === 'POST') {
          // For non-auth POST requests, add minimal body
          options.body = JSON.stringify({ test: true })
        }
        
        const response = await fetch(endpoint.url, options)
        console.log(`${endpoint.name}:`, response.ok ? '✅ OK' : '❌ Failed')
        console.log(`  Status: ${response.status}`)
        console.log(`  URL: ${endpoint.url}`)
        
        if (!response.ok) {
          const text = await response.text()
          console.log(`  Error: ${text}`)
          
          // Provide helpful hints for common issues
          if (response.status === 503 && endpoint.name === 'Health') {
            console.log(`  💡 Hint: Backend service is unhealthy. Check database connection.`)
          } else if (response.status === 400 && endpoint.name.includes('Auth')) {
            console.log(`  💡 Hint: Auth endpoint requires valid 'init_data' from Telegram WebApp.`)
          }
        }
      } catch (error) {
        console.log(`${endpoint.name}: ❌ Network Error`)
        console.error(`  Error:`, error)
        
        if (error instanceof TypeError && error.message.includes('fetch')) {
          console.log(`  💡 Hint: Check if backend server is running and accessible.`)
        }
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

  // Detailed backend diagnostics
  backend: async () => {
    const baseUrl = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'
    const healthUrl = baseUrl.replace('/api/v1', '/health')
    
    console.group('🔧 Backend Diagnostics')
    console.log('Health URL:', healthUrl)
    
    try {
      const response = await fetch(healthUrl)
      const data = await response.json()
      
      console.log('Status Code:', response.status)
      console.log('Response:', data)
      
      if (data.status === 'unhealthy') {
        console.log('')
        console.log('🚨 Backend Issues Detected:')
        
        if (data.error?.includes('database')) {
          console.log('  • Database connection failed')
          console.log('  • Check if database server is running')
          console.log('  • Verify database connection string')
          console.log('  • Check database credentials')
        }
        
        if (data.error?.includes('sql: database is closed')) {
          console.log('  • Database connection was closed unexpectedly')
          console.log('  • This might be a connection pool issue')
          console.log('  • Try restarting the backend service')
        }
        
        console.log('')
        console.log('💡 Troubleshooting Steps:')
        console.log('  1. Check backend logs for detailed error messages')
        console.log('  2. Verify database service is running')
        console.log('  3. Test database connection manually')
        console.log('  4. Restart backend service')
        console.log('  5. Check environment variables')
      }
    } catch (error) {
      console.log('❌ Cannot reach backend service')
      console.error('Error:', error)
      console.log('')
      console.log('💡 Possible causes:')
      console.log('  • Backend server is not running')
      console.log('  • Network connectivity issues')
      console.log('  • Incorrect API URL configuration')
      console.log('  • CORS issues (check browser network tab)')
    }
    
    console.groupEnd()
  },

  // Manual hash validation test
  testTelegramHash: async () => {
    const telegram = (window as unknown as { Telegram?: { WebApp: { initData?: string } } }).Telegram
    const initData = telegram?.WebApp?.initData
    
    if (!initData) {
      console.log('❌ No Telegram init data available')
      return
    }
    
    console.group('🔐 Telegram Hash Validation Test')
    console.log('Init Data:', initData)
    
    // Parse the init data
    const params = new URLSearchParams(initData)
    const hash = params.get('hash')
    
    // Create data check string (same logic as backend)
    const pairs = initData.split('&')
    const dataPairs = pairs.filter(pair => !pair.startsWith('hash='))
    dataPairs.sort()
    const dataCheckString = dataPairs.join('\n')
    
    console.log('Hash from Telegram:', hash)
    console.log('Data check string:', dataCheckString)
    console.log('Data pairs:', dataPairs)
    
    // Note: We can't calculate the expected hash here because we don't have the bot token
    // But we can at least verify the data structure
    console.groupEnd()
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
    console.log('debugUtils.backend() - Detailed backend diagnostics')
    console.log('debugUtils.testTelegramHash() - Test Telegram hash validation')
    console.log('debugUtils.clearStorage() - Clear all storage')
    console.log('debugUtils.help() - Show this help')
    console.groupEnd()
  },
}

// Make debug utils globally available in development
if (import.meta.env.DEV && typeof window !== 'undefined') {
  (window as unknown as { debugUtils: typeof debugUtils }).debugUtils = debugUtils
  
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