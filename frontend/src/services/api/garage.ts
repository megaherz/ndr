import { apiClient, queryKeys, APIResponse } from './client'
import { Decimal } from 'decimal.js'

// Types based on the garage.json contract
export interface GarageUser {
  id: string
  display_name: string
}

export interface GarageWallet {
  fuel_balance: string // Decimal string from API
  burn_balance: string // Decimal string from API
  rookie_races_completed: number
}

export interface GarageLeague {
  name: 'ROOKIE' | 'STREET' | 'PRO' | 'TOP_FUEL'
  buyin: string // Decimal string from API
  available: boolean
  unavailable_reason?: string | null
}

export interface GarageResponse {
  user: GarageUser
  wallet: GarageWallet
  leagues: GarageLeague[]
}

// Transform API response to store-compatible format
export interface TransformedGarageData {
  user: GarageUser
  wallet: {
    fuelBalance: Decimal
    burnBalance: Decimal
    rookieRacesCompleted: number
  }
  leagues: Array<{
    name: 'ROOKIE' | 'STREET' | 'PRO' | 'TOP_FUEL'
    buyin: Decimal
    available: boolean
    unavailable_reason?: string | null
  }>
}

/**
 * Fetch garage state (balances, league access)
 * GET /api/v1/garage
 */
export const fetchGarageState = async (): Promise<TransformedGarageData> => {
  // API client returns the full response: { success: true, data: {...}, timestamp: "..." }
  const apiResponse = await apiClient.get<APIResponse<GarageResponse>>('/garage')
  
  // Extract the garage data from the API response
  const garageData = (apiResponse as APIResponse<GarageResponse>).data
  
  // Validate response structure
  if (!garageData || !garageData.user || !garageData.wallet || !garageData.leagues) {
    throw new Error(`Invalid garage response structure. Expected data with user, wallet, leagues. Got: ${JSON.stringify(garageData)}`)
  }
  
  // Transform decimal strings to Decimal objects for consistent handling
  return {
    user: garageData.user,
    wallet: {
      fuelBalance: new Decimal(garageData.wallet.fuel_balance || '0'),
      burnBalance: new Decimal(garageData.wallet.burn_balance || '0'),
      rookieRacesCompleted: garageData.wallet.rookie_races_completed || 0,
    },
    leagues: garageData.leagues.map((league) => ({
      ...league,
      buyin: new Decimal(league.buyin || '0'),
    })),
  }
}

// React Query hook for garage state
export const garageQueries = {
  // Query key for garage state
  state: () => queryKeys.garage,
  
  // Query options for React Query
  stateOptions: () => ({
    queryKey: garageQueries.state(),
    queryFn: fetchGarageState,
    staleTime: 30 * 1000, // 30 seconds - garage state changes frequently during gameplay
    gcTime: 5 * 60 * 1000, // 5 minutes cache time
    retry: (failureCount: number, error: unknown) => {
      // Don't retry on authentication errors
      if (error && typeof error === 'object' && 'code' in error && error.code === 'UNAUTHORIZED') {
        return false
      }
      // Retry up to 2 times for other errors (garage is critical)
      return failureCount < 2
    },
    retryDelay: (attemptIndex: number) => Math.min(1000 * 2 ** attemptIndex, 5000),
  }),
}

// Error codes that can be returned by the garage API
export const GARAGE_ERROR_CODES = {
  UNAUTHORIZED: 'UNAUTHORIZED',
  WALLET_NOT_FOUND: 'WALLET_NOT_FOUND',
  INTERNAL_ERROR: 'INTERNAL_ERROR',
} as const

export type GarageErrorCode = typeof GARAGE_ERROR_CODES[keyof typeof GARAGE_ERROR_CODES]