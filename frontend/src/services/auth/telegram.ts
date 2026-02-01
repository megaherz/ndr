// Telegram initData extraction and validation utilities
// This service handles extracting and parsing Telegram initData for authentication

import { getTelegramInitData, getTelegramUser, isTelegramEnvironment } from '../telegram/webapp'

// Telegram initData structure
export interface TelegramInitData {
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

// Parsed user data for authentication
export interface TelegramAuthUser {
  id: number
  first_name: string
  last_name?: string
  username?: string
  language_code?: string
  is_premium?: boolean
}

// Authentication payload to send to backend
export interface TelegramAuthPayload {
  initData: string
  user: TelegramAuthUser
}

// Extract Telegram initData for authentication
export function extractTelegramInitData(): TelegramAuthPayload | null {
  try {
    // Check if running in Telegram environment
    if (!isTelegramEnvironment()) {
      console.warn('Not running in Telegram environment')
      
      // In development, return mock data
      if (import.meta.env.DEV) {
        return createMockAuthPayload()
      }
      
      return null
    }

    // Get raw initData string
    const initData = getTelegramInitData()
    if (!initData) {
      console.error('No Telegram initData available')
      return null
    }

    // Get user data from Telegram WebApp
    const user = getTelegramUser()
    if (!user) {
      console.error('No Telegram user data available')
      return null
    }

    // Validate required user fields
    if (!user.id || !user.first_name) {
      console.error('Invalid Telegram user data: missing required fields')
      return null
    }

    // Create authentication payload
    const authPayload: TelegramAuthPayload = {
      initData,
      user: {
        id: user.id,
        first_name: user.first_name,
        last_name: user.last_name,
        username: user.username,
        language_code: user.language_code,
        is_premium: user.is_premium,
      },
    }

    console.log('Telegram initData extracted successfully:', {
      userId: user.id,
      username: user.username,
      firstName: user.first_name,
    })

    return authPayload

  } catch (error) {
    console.error('Error extracting Telegram initData:', error)
    return null
  }
}

// Create mock authentication payload for development
function createMockAuthPayload(): TelegramAuthPayload {
  const mockUser: TelegramAuthUser = {
    id: 123456789,
    first_name: 'Test',
    last_name: 'User',
    username: 'testuser',
    language_code: 'en',
    is_premium: false,
  }

  // Create mock initData string (simplified format)
  const mockInitData = [
    `user=${encodeURIComponent(JSON.stringify(mockUser))}`,
    `auth_date=${Math.floor(Date.now() / 1000)}`,
    `hash=mock_hash_for_development`,
  ].join('&')

  return {
    initData: mockInitData,
    user: mockUser,
  }
}

// Parse initData string into structured object (for debugging/validation)
export function parseInitData(initData: string): TelegramInitData | null {
  try {
    const params = new URLSearchParams(initData)
    const parsed: TelegramInitData = {}

    // Extract query_id
    if (params.has('query_id')) {
      parsed.query_id = params.get('query_id') || undefined
    }

    // Extract and parse user data
    if (params.has('user')) {
      const userStr = params.get('user')
      if (userStr) {
        parsed.user = JSON.parse(decodeURIComponent(userStr))
      }
    }

    // Extract auth_date
    if (params.has('auth_date')) {
      const authDateStr = params.get('auth_date')
      if (authDateStr) {
        parsed.auth_date = parseInt(authDateStr, 10)
      }
    }

    // Extract hash
    if (params.has('hash')) {
      parsed.hash = params.get('hash') || undefined
    }

    return parsed

  } catch (error) {
    console.error('Error parsing initData:', error)
    return null
  }
}

// Validate initData format (basic client-side validation)
export function validateInitDataFormat(initData: string): boolean {
  try {
    const parsed = parseInitData(initData)
    
    if (!parsed) {
      return false
    }

    // Check required fields
    if (!parsed.user || !parsed.auth_date || !parsed.hash) {
      return false
    }

    // Check user required fields
    if (!parsed.user.id || !parsed.user.first_name) {
      return false
    }

    // Check auth_date is not too old (24 hours)
    const maxAge = 24 * 60 * 60 // 24 hours in seconds
    const now = Math.floor(Date.now() / 1000)
    if (now - parsed.auth_date > maxAge) {
      console.warn('Telegram initData is too old')
      return false
    }

    return true

  } catch (error) {
    console.error('Error validating initData format:', error)
    return false
  }
}

// Get user language from Telegram data (for i18n)
export function getTelegramUserLanguage(): string {
  const user = getTelegramUser()
  return user?.language_code || 'en'
}

// Check if user is premium
export function isTelegramUserPremium(): boolean {
  const user = getTelegramUser()
  return user?.is_premium || false
}