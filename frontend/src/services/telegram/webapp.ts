// Telegram WebApp API integration
// This service handles initialization and interaction with Telegram Mini App APIs

// Telegram WebApp types (basic definitions)
interface TelegramWebApp {
  initData: string
  initDataUnsafe: {
    query_id?: string
    user?: {
      id: number
      is_bot?: boolean
      first_name: string
      last_name?: string
      username?: string
      language_code?: string
      is_premium?: boolean
      allows_write_to_pm?: boolean
    }
    auth_date?: number
    hash?: string
  }
  version: string
  platform: string
  colorScheme: 'light' | 'dark'
  themeParams: {
    bg_color?: string
    text_color?: string
    hint_color?: string
    link_color?: string
    button_color?: string
    button_text_color?: string
    secondary_bg_color?: string
  }
  isExpanded: boolean
  viewportHeight: number
  viewportStableHeight: number
  headerColor: string
  backgroundColor: string
  isClosingConfirmationEnabled: boolean
  
  // Methods
  ready(): void
  expand(): void
  close(): void
  enableClosingConfirmation(): void
  disableClosingConfirmation(): void
  onEvent(eventType: string, eventHandler: () => void): void
  offEvent(eventType: string, eventHandler: () => void): void
  sendData(data: string): void
  switchInlineQuery(query: string, choose_chat_types?: string[]): void
  openLink(url: string, options?: { try_instant_view?: boolean }): void
  openTelegramLink(url: string): void
  openInvoice(url: string, callback?: (status: string) => void): void
  showPopup(params: {
    title?: string
    message: string
    buttons?: Array<{
      id?: string
      type?: 'default' | 'ok' | 'close' | 'cancel' | 'destructive'
      text: string
    }>
  }, callback?: (buttonId: string) => void): void
  showAlert(message: string, callback?: () => void): void
  showConfirm(message: string, callback?: (confirmed: boolean) => void): void
  showScanQrPopup(params: {
    text?: string
  }, callback?: (text: string) => boolean): void
  closeScanQrPopup(): void
  readTextFromClipboard(callback?: (text: string) => void): void
  requestWriteAccess(callback?: (granted: boolean) => void): void
  requestContact(callback?: (granted: boolean) => void): void
  
  // Haptic feedback
  HapticFeedback: {
    impactOccurred(style: 'light' | 'medium' | 'heavy' | 'rigid' | 'soft'): void
    notificationOccurred(type: 'error' | 'success' | 'warning'): void
    selectionChanged(): void
  }
  
  // Main button
  MainButton: {
    text: string
    color: string
    textColor: string
    isVisible: boolean
    isActive: boolean
    isProgressVisible: boolean
    setText(text: string): void
    onClick(callback: () => void): void
    offClick(callback: () => void): void
    show(): void
    hide(): void
    enable(): void
    disable(): void
    showProgress(leaveActive?: boolean): void
    hideProgress(): void
    setParams(params: {
      text?: string
      color?: string
      text_color?: string
      is_active?: boolean
      is_visible?: boolean
    }): void
  }
  
  // Back button
  BackButton: {
    isVisible: boolean
    onClick(callback: () => void): void
    offClick(callback: () => void): void
    show(): void
    hide(): void
  }
}

// Global Telegram object
declare global {
  interface Window {
    Telegram: {
      WebApp: TelegramWebApp
    } | undefined
  }
}

// Telegram WebApp service class
class TelegramWebAppService {
  private webApp: TelegramWebApp | null = null
  private isInitialized = false

  // Initialize Telegram WebApp
  initialize(): void {
    if (this.isInitialized) {
      return
    }

    // Check if running in Telegram environment
    if (typeof window !== 'undefined' && window.Telegram?.WebApp) {
      this.webApp = window.Telegram.WebApp
      
      // Initialize the WebApp
      this.webApp.ready()
      
      // Expand to full height
      this.webApp.expand()
      
      // Disable closing confirmation by default
      this.webApp.disableClosingConfirmation()
      
      // Set theme
      this.setupTheme()
      
      this.isInitialized = true
      
      console.log('Telegram WebApp initialized:', {
        version: this.webApp.version,
        platform: this.webApp.platform,
        colorScheme: this.webApp.colorScheme,
        user: this.webApp.initDataUnsafe.user,
      })
    } else {
      console.warn('Not running in Telegram WebApp environment')
      
      // In development, create mock WebApp for testing
      if (import.meta.env.DEV) {
        this.createMockWebApp()
      }
    }
  }

  // Create mock WebApp for development
  private createMockWebApp(): void {
    const mockUser = {
      id: 123456789,
      first_name: 'Test',
      last_name: 'User',
      username: 'testuser',
      language_code: 'en',
    }

    const mockInitDataUnsafe = {
      user: mockUser,
      auth_date: Math.floor(Date.now() / 1000),
      hash: 'mock_hash',
    }

    this.webApp = {
      initData: 'mock_init_data',
      initDataUnsafe: mockInitDataUnsafe,
      version: '6.7',
      platform: 'web',
      colorScheme: 'dark',
      themeParams: {
        bg_color: '#17212b',
        text_color: '#ffffff',
        hint_color: '#708499',
        link_color: '#6ab7ff',
        button_color: '#5288c1',
        button_text_color: '#ffffff',
        secondary_bg_color: '#232e3c',
      },
      isExpanded: true,
      viewportHeight: 600,
      viewportStableHeight: 600,
      headerColor: '#17212b',
      backgroundColor: '#17212b',
      isClosingConfirmationEnabled: false,
      
      // Mock methods
      ready: () => console.log('Mock WebApp ready'),
      expand: () => console.log('Mock WebApp expand'),
      close: () => console.log('Mock WebApp close'),
      enableClosingConfirmation: () => {},
      disableClosingConfirmation: () => {},
      onEvent: () => {},
      offEvent: () => {},
      sendData: (data: string) => console.log('Mock sendData:', data),
      switchInlineQuery: () => {},
      openLink: (url: string) => window.open(url, '_blank'),
      openTelegramLink: () => {},
      openInvoice: () => {},
      showPopup: (params: { message: string }, callback?: (result: string) => void) => {
        alert(params.message)
        callback?.('ok')
      },
      showAlert: (message: string, callback?: () => void) => {
        alert(message)
        callback?.()
      },
      showConfirm: (message: string, callback?: (result: boolean) => void) => {
        const result = confirm(message)
        callback?.(result)
      },
      showScanQrPopup: () => {},
      closeScanQrPopup: () => {},
      readTextFromClipboard: () => {},
      requestWriteAccess: (callback?: (granted: boolean) => void) => callback?.(true),
      requestContact: (callback?: (granted: boolean) => void) => callback?.(true),
      
      HapticFeedback: {
        impactOccurred: () => {},
        notificationOccurred: () => {},
        selectionChanged: () => {},
      },
      
      MainButton: {
        text: '',
        color: '#5288c1',
        textColor: '#ffffff',
        isVisible: false,
        isActive: true,
        isProgressVisible: false,
        setText: () => {},
        onClick: () => {},
        offClick: () => {},
        show: () => {},
        hide: () => {},
        enable: () => {},
        disable: () => {},
        showProgress: () => {},
        hideProgress: () => {},
        setParams: () => {},
      },
      
      BackButton: {
        isVisible: false,
        onClick: () => {},
        offClick: () => {},
        show: () => {},
        hide: () => {},
      },
    }

    this.isInitialized = true
    console.log('Mock Telegram WebApp created for development')
  }

  // Setup theme based on Telegram's color scheme
  private setupTheme(): void {
    if (!this.webApp) return

    const { themeParams, colorScheme } = this.webApp
    
    // Apply CSS custom properties for theming
    const root = document.documentElement
    
    if (themeParams.bg_color) {
      root.style.setProperty('--tg-bg-color', themeParams.bg_color)
    }
    
    if (themeParams.text_color) {
      root.style.setProperty('--tg-text-color', themeParams.text_color)
    }
    
    if (themeParams.hint_color) {
      root.style.setProperty('--tg-hint-color', themeParams.hint_color)
    }
    
    if (themeParams.link_color) {
      root.style.setProperty('--tg-link-color', themeParams.link_color)
    }
    
    if (themeParams.button_color) {
      root.style.setProperty('--tg-button-color', themeParams.button_color)
    }
    
    if (themeParams.button_text_color) {
      root.style.setProperty('--tg-button-text-color', themeParams.button_text_color)
    }
    
    // Set color scheme class
    document.body.classList.add(`tg-theme-${colorScheme}`)
  }

  // Get Telegram user data
  getUser() {
    return this.webApp?.initDataUnsafe.user || null
  }

  // Get init data for authentication
  getInitData(): string {
    return this.webApp?.initData || ''
  }

  // Get init data as object
  getInitDataUnsafe() {
    return this.webApp?.initDataUnsafe || null
  }

  // Check if running in Telegram
  isTelegramEnvironment(): boolean {
    return this.webApp !== null
  }

  // Show main button
  showMainButton(text: string, onClick: () => void): void {
    if (!this.webApp) return

    this.webApp.MainButton.setText(text)
    this.webApp.MainButton.onClick(onClick)
    this.webApp.MainButton.show()
  }

  // Hide main button
  hideMainButton(): void {
    if (!this.webApp) return
    this.webApp.MainButton.hide()
  }

  // Show back button
  showBackButton(onClick: () => void): void {
    if (!this.webApp) return

    this.webApp.BackButton.onClick(onClick)
    this.webApp.BackButton.show()
  }

  // Hide back button
  hideBackButton(): void {
    if (!this.webApp) return
    this.webApp.BackButton.hide()
  }

  // Haptic feedback
  hapticFeedback(type: 'impact' | 'notification' | 'selection', style?: string): void {
    if (!this.webApp) return

    switch (type) {
      case 'impact':
        this.webApp.HapticFeedback.impactOccurred((style as 'light' | 'medium' | 'heavy') || 'medium')
        break
      case 'notification':
        this.webApp.HapticFeedback.notificationOccurred((style as 'error' | 'success' | 'warning') || 'success')
        break
      case 'selection':
        this.webApp.HapticFeedback.selectionChanged()
        break
    }
  }

  // Show alert
  showAlert(message: string): Promise<void> {
    return new Promise((resolve) => {
      if (!this.webApp) {
        alert(message)
        resolve()
        return
      }

      this.webApp.showAlert(message, () => resolve())
    })
  }

  // Show confirm dialog
  showConfirm(message: string): Promise<boolean> {
    return new Promise((resolve) => {
      if (!this.webApp) {
        resolve(confirm(message))
        return
      }

      this.webApp.showConfirm(message, (confirmed) => resolve(confirmed))
    })
  }

  // Close the WebApp
  close(): void {
    if (!this.webApp) return
    this.webApp.close()
  }

  // Get WebApp instance
  getWebApp(): TelegramWebApp | null {
    return this.webApp
  }
}

// Export singleton instance
export const telegramWebApp = new TelegramWebAppService()

// Initialize function to be called from App.tsx
export const initializeTelegramWebApp = () => {
  telegramWebApp.initialize()
}

// Export utility functions
export const getTelegramUser = () => telegramWebApp.getUser()
export const getTelegramInitData = () => telegramWebApp.getInitData()
export const isTelegramEnvironment = () => telegramWebApp.isTelegramEnvironment()

// Export types
export type { TelegramWebApp }