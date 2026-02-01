import { describe, it, expect, vi, beforeEach } from 'vitest'
import { Decimal } from 'decimal.js'
import { fetchGarageState, garageQueries } from '../garage'
import { apiClient } from '../client'

// Mock the API client
vi.mock('../client', () => ({
  apiClient: {
    get: vi.fn(),
  },
  queryKeys: {
    garage: ['garage'],
  },
}))

describe('Garage API Service', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  describe('fetchGarageState', () => {
    it('should fetch and transform garage state correctly', async () => {
      // Mock API response with full API response structure
      const mockApiResponse = {
        success: true,
        data: {
          user: {
            id: 'test-user-id',
            display_name: 'Test User',
          },
          wallet: {
            fuel_balance: '100.50',
            burn_balance: '25.75',
            rookie_races_completed: 2,
          },
          leagues: [
            {
              name: 'ROOKIE' as const,
              buyin: '10.00',
              available: true,
              unavailable_reason: null,
            },
            {
              name: 'STREET' as const,
              buyin: '50.00',
              available: false,
              unavailable_reason: 'INSUFFICIENT_BALANCE',
            },
          ],
        },
        timestamp: '2026-02-01T16:00:00Z',
      }

      vi.mocked(apiClient.get).mockResolvedValue(mockApiResponse)

      // Call the function
      const result = await fetchGarageState()

      // Verify API call
      expect(apiClient.get).toHaveBeenCalledWith('/garage')

      // Verify transformation
      expect(result).toEqual({
        user: {
          id: 'test-user-id',
          display_name: 'Test User',
        },
        wallet: {
          fuelBalance: new Decimal('100.50'),
          burnBalance: new Decimal('25.75'),
          rookieRacesCompleted: 2,
        },
        leagues: [
          {
            name: 'ROOKIE',
            buyin: new Decimal('10.00'),
            available: true,
            unavailable_reason: null,
          },
          {
            name: 'STREET',
            buyin: new Decimal('50.00'),
            available: false,
            unavailable_reason: 'INSUFFICIENT_BALANCE',
          },
        ],
      })
    })

    it('should handle API errors correctly', async () => {
      const mockError = new Error('API Error')
      vi.mocked(apiClient.get).mockRejectedValue(mockError)

      await expect(fetchGarageState()).rejects.toThrow('API Error')
    })
  })

  describe('garageQueries', () => {
    it('should provide correct query key', () => {
      expect(garageQueries.state()).toEqual(['garage'])
    })

    it('should provide query options with correct configuration', () => {
      const options = garageQueries.stateOptions()

      expect(options.queryKey).toEqual(['garage'])
      expect(options.queryFn).toBe(fetchGarageState)
      expect(options.staleTime).toBe(30 * 1000) // 30 seconds
      expect(options.gcTime).toBe(5 * 60 * 1000) // 5 minutes
      expect(typeof options.retry).toBe('function')
      expect(typeof options.retryDelay).toBe('function')
    })

    it('should not retry on unauthorized errors', () => {
      const options = garageQueries.stateOptions()
      const retryFunction = options.retry as (failureCount: number, error: unknown) => boolean

      // Test unauthorized error
      const unauthorizedError = { code: 'UNAUTHORIZED' }
      expect(retryFunction(1, unauthorizedError)).toBe(false)

      // Test other errors
      const otherError = new Error('Network error')
      expect(retryFunction(1, otherError)).toBe(true)
      expect(retryFunction(2, otherError)).toBe(false) // Max 2 retries
    })
  })
})