// Eruda mobile console integration
// This service provides mobile debugging capabilities for Telegram Mini Apps

import { isTelegramEnvironment } from '../telegram/webapp'

// Eruda instance type
interface ErudaInstance {
  init: (config?: ErudaConfig) => void
  show: () => void
  hide: () => void
  destroy: () => void
  get: (name: string) => any
  add: (plugin: any) => void
  remove: (name: string) => void
  position: (config: { x: number; y: number }) => void
  scale: (scale: number) => void
}

interface ErudaConfig {
  container?: HTMLElement
  tool?: string[]
  autoScale?: boolean
  useShadowDom?: boolean
  defaults?: {
    displaySize?: number
    transparency?: number
    theme?: 'Light' | 'Dark'
  }
}

declare global {
  interface Window {
    eruda?: ErudaInstance
  }
}

// Eruda service class
class ErudaService {
  private isInitialized = false
  private isLoaded = false
  private loadPromise: Promise<void> | null = null

  // Check if Eruda should be enabled
  private shouldEnableEruda(): boolean {
    // Enable if explicitly set in environment
    if (import.meta.env.VITE_ENABLE_ERUDA === 'true') {
      return true
    }

    // Always enable in development
    if (import.meta.env.DEV) {
      return true
    }

    // Enable if debug flag is set in URL
    const urlParams = new URLSearchParams(window.location.search)
    if (urlParams.get('debug') === 'true' || urlParams.get('eruda') === 'true') {
      return true
    }

    // Enable if localStorage flag is set
    try {
      if (localStorage.getItem('ndr-debug') === 'true') {
        return true
      }
    } catch (error) {
      // Ignore localStorage errors
    }

    // Enable in Telegram environment for easier debugging
    if (isTelegramEnvironment()) {
      // Check if user is in development mode (you can customize this logic)
      const isDev = urlParams.get('env') === 'dev' || 
                   window.location.hostname === 'localhost' ||
                   window.location.hostname.includes('dev') ||
                   window.location.hostname.includes('staging')
      
      if (isDev) {
        return true
      }
    }

    return false
  }

  // Load Eruda dynamically
  private async loadEruda(): Promise<void> {
    if (this.loadPromise) {
      return this.loadPromise
    }

    this.loadPromise = new Promise((resolve, reject) => {
      try {
        // Import Eruda dynamically
        import('eruda').then((eruda) => {
          window.eruda = eruda.default
          this.isLoaded = true
          console.log('Eruda loaded successfully')
          resolve()
        }).catch((error) => {
          console.error('Failed to load Eruda:', error)
          reject(error)
        })
      } catch (error) {
        console.error('Failed to import Eruda:', error)
        reject(error)
      }
    })

    return this.loadPromise
  }

  // Initialize Eruda
  async initialize(): Promise<void> {
    if (this.isInitialized) {
      return
    }

    if (!this.shouldEnableEruda()) {
      console.log('Eruda disabled - not in debug mode')
      return
    }

    try {
      console.log('Initializing Eruda mobile console...')

      // Load Eruda
      await this.loadEruda()

      if (!window.eruda) {
        throw new Error('Eruda not available after loading')
      }

      // Configure Eruda for Telegram Mini App
      const config: ErudaConfig = {
        // Don't auto-scale to preserve Telegram's viewport handling
        autoScale: false,
        
        // Use shadow DOM to avoid conflicts with Telegram styles
        useShadowDom: true,
        
        // Enable useful tools for mobile debugging
        tool: [
          'console',    // Console logs
          'elements',   // DOM inspector
          'network',    // Network requests
          'resources',  // Resources (localStorage, etc.)
          'info',       // Device/browser info
          'sources',    // Source code viewer
          'snippets'    // Code snippets
        ],
        
        defaults: {
          // Smaller display size for mobile
          displaySize: 40,
          
          // Semi-transparent to see app behind
          transparency: 0.9,
          
          // Use dark theme to match Telegram
          theme: 'Dark'
        }
      }

      // Initialize Eruda
      window.eruda.init(config)

      // Position it in a convenient location (bottom-right)
      window.eruda.position({ x: window.innerWidth - 60, y: window.innerHeight - 60 })

      // Hide by default - user can show it by tapping the icon
      window.eruda.hide()

      this.isInitialized = true

      console.log('Eruda initialized successfully')

      // Add helpful console message
      console.log('%c🚀 Nitro Drag Royale Debug Console', 
        'color: #5288c1; font-size: 16px; font-weight: bold;')
      console.log('%cTap the Eruda icon to open the mobile console', 
        'color: #708499; font-size: 12px;')
      console.log('%cAvailable commands:', 'color: #6ab7ff; font-weight: bold;')
      console.log('  - eruda.show() - Show console')
      console.log('  - eruda.hide() - Hide console')
      console.log('  - localStorage.setItem("ndr-debug", "true") - Enable debug mode')

    } catch (error) {
      console.error('Failed to initialize Eruda:', error)
    }
  }

  // Show Eruda console
  show(): void {
    if (window.eruda && this.isInitialized) {
      window.eruda.show()
    } else {
      console.warn('Eruda not initialized')
    }
  }

  // Hide Eruda console
  hide(): void {
    if (window.eruda && this.isInitialized) {
      window.eruda.hide()
    }
  }

  // Check if Eruda is available
  isAvailable(): boolean {
    return this.isInitialized && !!window.eruda
  }

  // Enable debug mode persistently
  enableDebugMode(): void {
    try {
      localStorage.setItem('ndr-debug', 'true')
      console.log('Debug mode enabled. Reload the app to activate Eruda.')
    } catch (error) {
      console.warn('Could not enable debug mode:', error)
    }
  }

  // Disable debug mode
  disableDebugMode(): void {
    try {
      localStorage.removeItem('ndr-debug')
      console.log('Debug mode disabled. Reload the app to deactivate Eruda.')
    } catch (error) {
      console.warn('Could not disable debug mode:', error)
    }
  }

  // Log app-specific debug information
  logAppInfo(): void {
    if (!this.isAvailable()) {
      return
    }

    console.group('🎮 Nitro Drag Royale App Info')
    
    // Environment info
    console.log('Environment:', import.meta.env.MODE)
    console.log('Telegram Environment:', isTelegramEnvironment())
    
    // Telegram WebApp info
    if (window.Telegram?.WebApp) {
      const webApp = window.Telegram.WebApp
      console.log('Telegram WebApp Version:', webApp.version)
      console.log('Platform:', webApp.platform)
      console.log('Color Scheme:', webApp.colorScheme)
      console.log('Viewport Height:', webApp.viewportHeight)
      console.log('User:', webApp.initDataUnsafe.user)
    }
    
    // App version (if available in package.json)
    console.log('App Version:', import.meta.env.VITE_APP_VERSION || 'Unknown')
    
    // API Base URL
    console.log('API Base URL:', import.meta.env.VITE_API_BASE_URL || 'Default')
    
    console.groupEnd()
  }
}

// Export singleton instance
export const erudaService = new ErudaService()

// Initialize function to be called from App.tsx
export const initializeEruda = async () => {
  await erudaService.initialize()
}

// Export utility functions for global access
export const showDebugConsole = () => erudaService.show()
export const hideDebugConsole = () => erudaService.hide()
export const enableDebugMode = () => erudaService.enableDebugMode()
export const disableDebugMode = () => erudaService.disableDebugMode()
export const logAppInfo = () => erudaService.logAppInfo()

// Make debug functions globally available for easy access in production
if (typeof window !== 'undefined') {
  (window as any).NDR_DEBUG = {
    show: showDebugConsole,
    hide: hideDebugConsole,
    enable: enableDebugMode,
    disable: disableDebugMode,
    info: logAppInfo,
  }
}