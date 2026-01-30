import { useAuthStore } from '../../stores/authStore'

// API configuration
const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1'

// Request types
export interface APIErrorResponse {
  error: string
  message: string
  code?: string
  details?: Record<string, unknown>
}

export interface APIResponse<T> {
  data: T
  success: boolean
  timestamp: string
}

// HTTP client class
class APIClient {
  private baseURL: string

  constructor(baseURL: string) {
    this.baseURL = baseURL
  }

  private async getAuthHeaders(): Promise<Record<string, string>> {
    const { tokens, isTokenExpired } = useAuthStore.getState()
    
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    }

    if (tokens && !isTokenExpired()) {
      headers.Authorization = `Bearer ${tokens.appToken}`
    }

    return headers
  }

  private async handleResponse<T>(response: Response): Promise<T> {
    const contentType = response.headers.get('content-type')
    const isJSON = contentType?.includes('application/json')

    if (!response.ok) {
      let errorData: APIErrorResponse
      
      if (isJSON) {
        errorData = await response.json()
      } else {
        errorData = {
          error: 'HTTP_ERROR',
          message: `HTTP ${response.status}: ${response.statusText}`,
          code: 'HTTP_ERROR',
        }
      }

      // Handle authentication errors
      if (response.status === 401) {
        const { logout } = useAuthStore.getState()
        logout()
        errorData.message = 'Authentication required. Please log in again.'
      }

      throw new APIError(errorData.message, errorData.code || 'API_ERROR', errorData.details)
    }

    if (!isJSON) {
      throw new APIError('Invalid response format', 'INVALID_RESPONSE')
    }

    const data = await response.json()
    return data
  }

  async get<T>(endpoint: string, params?: Record<string, string>): Promise<T> {
    const url = new URL(`${this.baseURL}${endpoint}`)
    
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        url.searchParams.append(key, value)
      })
    }

    const headers = await this.getAuthHeaders()

    const response = await fetch(url.toString(), {
      method: 'GET',
      headers,
    })

    return this.handleResponse<T>(response)
  }

  async post<T>(endpoint: string, data?: unknown): Promise<T> {
    const headers = await this.getAuthHeaders()

    const response = await fetch(`${this.baseURL}${endpoint}`, {
      method: 'POST',
      headers,
      body: data ? JSON.stringify(data) : undefined,
    })

    return this.handleResponse<T>(response)
  }

  async put<T>(endpoint: string, data?: unknown): Promise<T> {
    const headers = await this.getAuthHeaders()

    const response = await fetch(`${this.baseURL}${endpoint}`, {
      method: 'PUT',
      headers,
      body: data ? JSON.stringify(data) : undefined,
    })

    return this.handleResponse<T>(response)
  }

  async delete<T>(endpoint: string): Promise<T> {
    const headers = await this.getAuthHeaders()

    const response = await fetch(`${this.baseURL}${endpoint}`, {
      method: 'DELETE',
      headers,
    })

    return this.handleResponse<T>(response)
  }
}

// Custom API Error class
export class APIError extends Error {
  public readonly code: string
  public readonly details?: Record<string, unknown>

  constructor(message: string, code: string = 'API_ERROR', details?: Record<string, unknown>) {
    super(message)
    this.name = 'APIError'
    this.code = code
    this.details = details
  }
}

// Export singleton instance
export const apiClient = new APIClient(API_BASE_URL)

// React Query configuration helpers
export const queryConfig = {
  retry: (failureCount: number, error: unknown) => {
    // Don't retry on authentication errors
    if (error instanceof APIError && error.code === 'UNAUTHORIZED') {
      return false
    }
    // Retry up to 3 times for other errors
    return failureCount < 3
  },
  retryDelay: (attemptIndex: number) => Math.min(1000 * 2 ** attemptIndex, 30000),
  staleTime: 5 * 60 * 1000, // 5 minutes
  gcTime: 10 * 60 * 1000, // 10 minutes (formerly cacheTime)
}

// Query key factories for consistent cache management
export const queryKeys = {
  // Auth
  auth: ['auth'] as const,
  
  // Garage
  garage: ['garage'] as const,
  
  // Payments
  payments: ['payments'] as const,
  paymentHistory: (limit?: number, offset?: number) => 
    ['payments', 'history', { limit, offset }] as const,
  
  // Matches
  matches: ['matches'] as const,
  match: (matchId: string) => ['matches', matchId] as const,
  
  // User
  user: ['user'] as const,
  userProfile: (userId: string) => ['user', userId] as const,
} as const

// Mutation key factories
export const mutationKeys = {
  // Auth
  login: ['auth', 'login'] as const,
  logout: ['auth', 'logout'] as const,
  
  // Payments
  createDeposit: ['payments', 'deposit'] as const,
  createWithdrawal: ['payments', 'withdrawal'] as const,
  
  // Matchmaking (handled via RPC, but tracked for UI state)
  joinMatchmaking: ['matchmaking', 'join'] as const,
  cancelMatchmaking: ['matchmaking', 'cancel'] as const,
  
  // Match actions (handled via RPC, but tracked for UI state)
  earnPoints: ['match', 'earn-points'] as const,
  giveUp: ['match', 'give-up'] as const,
} as const

// Type helpers for React Query
export type QueryKey = typeof queryKeys[keyof typeof queryKeys]
export type MutationKey = typeof mutationKeys[keyof typeof mutationKeys]