// Authentication API calls
// This service handles authentication-related API requests

import { apiClient, APIResponse } from './client'

// Get API base URL for debugging
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'
import { TelegramAuthPayload } from '../auth/telegram'
import { User, AuthTokens } from '../../stores/authStore'

// Login request payload
export interface LoginRequest {
  initData: string
  user: {
    id: number
    first_name: string
    last_name?: string
    username?: string
    language_code?: string
    is_premium?: boolean
  }
}

// Login response from backend
export interface LoginResponse {
  user: {
    id: string
    telegram_id: number
    telegram_username?: string
    telegram_first_name: string
    telegram_last_name?: string
    created_at: string
    updated_at: string
  }
  tokens: {
    app_token: string
    centrifugo_token: string
    expires_at: string // ISO 8601 timestamp
  }
}

// Transformed user data for the store
function transformUserData(backendUser: LoginResponse['user']): User {
  return {
    id: backendUser.id,
    telegramId: backendUser.telegram_id,
    telegramUsername: backendUser.telegram_username,
    telegramFirstName: backendUser.telegram_first_name,
    telegramLastName: backendUser.telegram_last_name,
  }
}

// Transformed token data for the store
function transformTokenData(backendTokens: LoginResponse['tokens']): AuthTokens {
  return {
    appToken: backendTokens.app_token,
    centrifugoToken: backendTokens.centrifugo_token,
    expiresAt: new Date(backendTokens.expires_at).getTime(),
  }
}

// Login with Telegram initData
export async function loginWithTelegram(authPayload: TelegramAuthPayload): Promise<{
  user: User
  tokens: AuthTokens
}> {
  try {
    console.log('Attempting login with Telegram initData:', {
      userId: authPayload.user.id,
      username: authPayload.user.username,
    })

    // Prepare request payload
    const loginRequest: LoginRequest = {
      initData: authPayload.initData,
      user: authPayload.user,
    }

    // Make API call
    const response = await apiClient.post<APIResponse<LoginResponse>>(
      '/auth/telegram',
      loginRequest
    )

    if (!response.data) {
      throw new Error('Invalid response format: missing data')
    }

    // Transform response data
    const user = transformUserData(response.data.user)
    const tokens = transformTokenData(response.data.tokens)

    console.log('Login successful:', {
      userId: user.id,
      telegramId: user.telegramId,
      expiresAt: new Date(tokens.expiresAt).toISOString(),
    })

    return { user, tokens }

  } catch (error) {
    console.error('Login failed:', error)
    console.error('API Base URL:', API_BASE_URL)
    console.error('Request payload:', loginRequest)
    
    // Re-throw with more context
    if (error instanceof Error) {
      throw new Error(`Authentication failed: ${error.message}`)
    }
    
    throw new Error('Authentication failed: Unknown error')
  }
}

// Logout (client-side only, tokens are stateless)
export async function logout(): Promise<void> {
  try {
    console.log('Logging out user')
    
    // Note: Since we use stateless JWT tokens, we don't need to call the backend
    // The logout is handled entirely on the client side by clearing the tokens
    
    // In the future, if we implement token blacklisting or need to track sessions,
    // we could add a backend call here:
    // await apiClient.post('/auth/logout')
    
    console.log('Logout completed')

  } catch (error) {
    console.error('Logout error (non-critical):', error)
    // Don't throw - logout should always succeed on client side
  }
}

// Refresh tokens (if implemented in the future)
export async function refreshTokens(): Promise<AuthTokens> {
  try {
    console.log('Attempting to refresh tokens')

    // Note: This is a placeholder for future implementation
    // Currently, our JWT tokens are short-lived and the user will need to re-authenticate
    // when they expire. In the future, we might implement refresh tokens.
    
    throw new Error('Token refresh not implemented - please re-authenticate')

  } catch (error) {
    console.error('Token refresh failed:', error)
    throw error
  }
}

// Validate current authentication status
export async function validateAuth(): Promise<boolean> {
  try {
    // Make a simple authenticated request to validate tokens
    // This could be a call to /auth/me or any protected endpoint
    await apiClient.get('/garage')
    return true

  } catch (error) {
    console.log('Auth validation failed:', error)
    return false
  }
}

// Get current user profile (if needed)
export async function getCurrentUser(): Promise<User> {
  try {
    const response = await apiClient.get<APIResponse<LoginResponse['user']>>('/auth/me')
    
    if (!response.data) {
      throw new Error('Invalid response format: missing data')
    }

    return transformUserData(response.data)

  } catch (error) {
    console.error('Failed to get current user:', error)
    throw error
  }
}