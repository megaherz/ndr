import { create } from 'zustand'
import { Decimal } from 'decimal.js'

export interface WalletBalances {
  tonBalance: Decimal
  fuelBalance: Decimal
  burnBalance: Decimal
  rookieRacesCompleted: number
}

export interface LeagueAccess {
  rookie: boolean
  street: boolean
  pro: boolean
  topFuel: boolean
}

interface WalletState {
  // State
  balances: WalletBalances | null
  leagueAccess: LeagueAccess | null
  tonWalletAddress: string | null
  isLoading: boolean
  error: string | null
  lastUpdated: number | null

  // Actions
  setBalances: (balances: WalletBalances) => void
  setLeagueAccess: (access: LeagueAccess) => void
  setTonWalletAddress: (address: string | null) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  clearError: () => void
  updateBalance: (currency: 'ton' | 'fuel' | 'burn', amount: Decimal) => void
  incrementRookieRaces: () => void
  
  // Computed
  canAccessLeague: (league: string) => boolean
  hasInsufficientBalance: (league: string) => boolean
}

// League buy-in amounts (must match backend constants)
const LEAGUE_BUYINS = {
  ROOKIE: new Decimal('10.00'),
  STREET: new Decimal('50.00'),
  PRO: new Decimal('300.00'),
  TOP_FUEL: new Decimal('3000.00'),
}

export const useWalletStore = create<WalletState>((set, get) => ({
  // Initial state
  balances: null,
  leagueAccess: null,
  tonWalletAddress: null,
  isLoading: false,
  error: null,
  lastUpdated: null,

  // Actions
  setBalances: (balances) => {
    set({ 
      balances, 
      error: null,
      lastUpdated: Date.now() 
    })
  },

  setLeagueAccess: (access) => {
    set({ 
      leagueAccess: access,
      error: null 
    })
  },

  setTonWalletAddress: (address) => {
    set({ tonWalletAddress: address })
  },

  setLoading: (loading) => {
    set({ isLoading: loading })
  },

  setError: (error) => {
    set({ error })
  },

  clearError: () => {
    set({ error: null })
  },

  updateBalance: (currency, amount) => {
    const { balances } = get()
    if (!balances) return

    const updatedBalances = { ...balances }
    switch (currency) {
      case 'ton':
        updatedBalances.tonBalance = amount
        break
      case 'fuel':
        updatedBalances.fuelBalance = amount
        break
      case 'burn':
        updatedBalances.burnBalance = amount
        break
    }

    set({ 
      balances: updatedBalances,
      lastUpdated: Date.now() 
    })
  },

  incrementRookieRaces: () => {
    const { balances } = get()
    if (!balances) return

    const updatedBalances = {
      ...balances,
      rookieRacesCompleted: Math.min(balances.rookieRacesCompleted + 1, 3)
    }

    set({ 
      balances: updatedBalances,
      lastUpdated: Date.now() 
    })
  },

  // Computed
  canAccessLeague: (league: string) => {
    const { balances, leagueAccess } = get()
    if (!balances || !leagueAccess) return false

    switch (league.toUpperCase()) {
      case 'ROOKIE':
        return leagueAccess.rookie && balances.rookieRacesCompleted < 3
      case 'STREET':
        return leagueAccess.street
      case 'PRO':
        return leagueAccess.pro
      case 'TOP_FUEL':
        return leagueAccess.topFuel
      default:
        return false
    }
  },

  hasInsufficientBalance: (league: string) => {
    const { balances } = get()
    if (!balances) return true

    const buyin = LEAGUE_BUYINS[league.toUpperCase() as keyof typeof LEAGUE_BUYINS]
    if (!buyin) return true

    return balances.fuelBalance.lessThan(buyin)
  },
}))