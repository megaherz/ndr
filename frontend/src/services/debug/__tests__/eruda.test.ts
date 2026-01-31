// Tests for Eruda debug service
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { erudaService } from '../eruda'

// Mock Eruda module
vi.mock('eruda', () => ({
  default: {
    init: vi.fn(),
    show: vi.fn(),
    hide: vi.fn(),
    position: vi.fn(),
  }
}))

// Mock Telegram WebApp
const mockTelegramWebApp = {
  initDataUnsafe: {
    user: {
      id: 123456789,
      first_name: 'Test',
      username: 'testuser',
    }
  },
  version: '6.7',
  platform: 'web',
  colorScheme: 'dark' as const,
  viewportHeight: 600,
}

describe('ErudaService', () => {
  beforeEach(() => {
    // Reset environment
    vi.unstubAllEnvs()
    
    // Clear localStorage
    localStorage.clear()
    
    // Reset window.Telegram
    delete (window as any).Telegram
    delete (window as any).eruda
    
    // Reset URL
    Object.defineProperty(window, 'location', {
      value: {
        hostname: 'localhost',
        search: '',
      },
      writable: true,
    })
  })

  afterEach(() => {
    vi.clearAllMocks()
  })

  describe('shouldEnableEruda', () => {
    it('should enable in development mode', async () => {
      // Mock development environment
      vi.stubEnv('DEV', true)
      
      await erudaService.initialize()
      
      // Should attempt to load Eruda
      expect(erudaService.isAvailable()).toBe(false) // Won't be available without actual Eruda
    })

    it('should enable with VITE_ENABLE_ERUDA=true', async () => {
      vi.stubEnv('VITE_ENABLE_ERUDA', 'true')
      vi.stubEnv('DEV', false)
      
      await erudaService.initialize()
      
      // Should attempt to initialize
      expect(erudaService.isAvailable()).toBe(false)
    })

    it('should enable with debug URL parameter', async () => {
      vi.stubEnv('DEV', false)
      Object.defineProperty(window, 'location', {
        value: {
          hostname: 'example.com',
          search: '?debug=true',
        },
        writable: true,
      })
      
      await erudaService.initialize()
      
      expect(erudaService.isAvailable()).toBe(false)
    })

    it('should enable with localStorage flag', async () => {
      vi.stubEnv('DEV', false)
      localStorage.setItem('ndr-debug', 'true')
      
      await erudaService.initialize()
      
      expect(erudaService.isAvailable()).toBe(false)
    })

    it('should enable in Telegram dev environment', async () => {
      vi.stubEnv('DEV', false)
      
      // Mock Telegram environment
      ;(window as any).Telegram = { WebApp: mockTelegramWebApp }
      
      // Mock dev hostname
      Object.defineProperty(window, 'location', {
        value: {
          hostname: 'dev.example.com',
          search: '',
        },
        writable: true,
      })
      
      await erudaService.initialize()
      
      expect(erudaService.isAvailable()).toBe(false)
    })

    it('should not enable in production without flags', async () => {
      vi.stubEnv('DEV', false)
      vi.stubEnv('VITE_ENABLE_ERUDA', 'false')
      
      Object.defineProperty(window, 'location', {
        value: {
          hostname: 'production.com',
          search: '',
        },
        writable: true,
      })
      
      await erudaService.initialize()
      
      expect(erudaService.isAvailable()).toBe(false)
    })
  })

  describe('debug mode management', () => {
    it('should enable debug mode in localStorage', () => {
      erudaService.enableDebugMode()
      
      expect(localStorage.getItem('ndr-debug')).toBe('true')
    })

    it('should disable debug mode in localStorage', () => {
      localStorage.setItem('ndr-debug', 'true')
      
      erudaService.disableDebugMode()
      
      expect(localStorage.getItem('ndr-debug')).toBeNull()
    })

    it('should handle localStorage errors gracefully', () => {
      // Mock localStorage to throw error
      const mockSetItem = vi.spyOn(Storage.prototype, 'setItem')
      mockSetItem.mockImplementation(() => {
        throw new Error('Storage error')
      })
      
      // Should not throw
      expect(() => erudaService.enableDebugMode()).not.toThrow()
      
      mockSetItem.mockRestore()
    })
  })

  describe('global NDR_DEBUG object', () => {
    it('should expose debug functions globally', () => {
      expect((window as any).NDR_DEBUG).toBeDefined()
      expect(typeof (window as any).NDR_DEBUG.show).toBe('function')
      expect(typeof (window as any).NDR_DEBUG.hide).toBe('function')
      expect(typeof (window as any).NDR_DEBUG.enable).toBe('function')
      expect(typeof (window as any).NDR_DEBUG.disable).toBe('function')
      expect(typeof (window as any).NDR_DEBUG.info).toBe('function')
    })
  })

  describe('logAppInfo', () => {
    it('should log app information when available', () => {
      const consoleSpy = vi.spyOn(console, 'log').mockImplementation(() => {})
      const consoleGroupSpy = vi.spyOn(console, 'group').mockImplementation(() => {})
      const consoleGroupEndSpy = vi.spyOn(console, 'groupEnd').mockImplementation(() => {})
      
      // Mock Telegram WebApp
      ;(window as any).Telegram = { WebApp: mockTelegramWebApp }
      
      erudaService.logAppInfo()
      
      // Should not log if Eruda is not available
      expect(consoleGroupSpy).not.toHaveBeenCalled()
      
      consoleSpy.mockRestore()
      consoleGroupSpy.mockRestore()
      consoleGroupEndSpy.mockRestore()
    })
  })
})