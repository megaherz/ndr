import { create } from 'zustand'
import { Decimal } from 'decimal.js'

export interface MatchParticipant {
  userId?: string
  displayName: string
  isGhost: boolean
  heat1Score?: Decimal
  heat2Score?: Decimal
  heat3Score?: Decimal
  totalScore?: Decimal
  currentPosition?: number
  finalPosition?: number
  prizeAmount: Decimal
  burnReward: Decimal
}

export interface HeatData {
  heatNumber: number
  status: 'waiting' | 'countdown' | 'active' | 'completed'
  duration: number
  startedAt?: number
  targetLine?: Decimal // For Heat 2 and 3
  playerScore?: Decimal
  playerLocked: boolean
}

export interface MatchData {
  id: string
  league: string
  status: 'forming' | 'in_progress' | 'completed' | 'aborted'
  participants: MatchParticipant[]
  currentHeat: number
  heats: HeatData[]
  prizePool: Decimal
  crashSeed?: string
  crashSeedHash?: string
}

interface MatchState {
  // Current match state
  currentMatch: MatchData | null
  queueTicketId: string | null
  isInQueue: boolean
  isInMatch: boolean
  estimatedWaitSeconds: number
  
  // UI state
  isLoading: boolean
  error: string | null
  
  // Real-time connection
  isConnected: boolean
  lastEventTimestamp: number | null

  // Actions
  setCurrentMatch: (match: MatchData | null) => void
  setQueueTicketId: (ticketId: string | null) => void
  setInQueue: (inQueue: boolean) => void
  setInMatch: (inMatch: boolean) => void
  setEstimatedWaitSeconds: (seconds: number) => void
  setLoading: (loading: boolean) => void
  setError: (error: string | null) => void
  setConnected: (connected: boolean) => void
  setLastEventTimestamp: (timestamp: number) => void
  clearError: () => void
  
  // Match updates
  updateHeatData: (heatNumber: number, data: Partial<HeatData>) => void
  updateParticipant: (userId: string | undefined, updates: Partial<MatchParticipant>) => void
  updatePlayerScore: (heatNumber: number, score: Decimal) => void
  setPlayerLocked: (heatNumber: number, locked: boolean) => void
  
  // Reset
  resetMatch: () => void
  resetQueue: () => void
}

export const useMatchStore = create<MatchState>((set, get) => ({
  // Initial state
  currentMatch: null,
  queueTicketId: null,
  isInQueue: false,
  isInMatch: false,
  estimatedWaitSeconds: 0,
  isLoading: false,
  error: null,
  isConnected: false,
  lastEventTimestamp: null,

  // Actions
  setCurrentMatch: (match) => {
    set({ 
      currentMatch: match,
      isInMatch: match !== null,
      error: null 
    })
  },

  setQueueTicketId: (ticketId) => {
    set({ queueTicketId: ticketId })
  },

  setInQueue: (inQueue) => {
    set({ isInQueue: inQueue })
  },

  setInMatch: (inMatch) => {
    set({ isInMatch: inMatch })
  },

  setEstimatedWaitSeconds: (seconds) => {
    set({ estimatedWaitSeconds: seconds })
  },

  setLoading: (loading) => {
    set({ isLoading: loading })
  },

  setError: (error) => {
    set({ error })
  },

  setConnected: (connected) => {
    set({ isConnected: connected })
  },

  setLastEventTimestamp: (timestamp) => {
    set({ lastEventTimestamp: timestamp })
  },

  clearError: () => {
    set({ error: null })
  },

  // Match updates
  updateHeatData: (heatNumber, data) => {
    const { currentMatch } = get()
    if (!currentMatch) return

    const updatedHeats = currentMatch.heats.map(heat => 
      heat.heatNumber === heatNumber 
        ? { ...heat, ...data }
        : heat
    )

    set({
      currentMatch: {
        ...currentMatch,
        heats: updatedHeats
      }
    })
  },

  updateParticipant: (userId, updates) => {
    const { currentMatch } = get()
    if (!currentMatch) return

    const updatedParticipants = currentMatch.participants.map(participant =>
      participant.userId === userId
        ? { ...participant, ...updates }
        : participant
    )

    set({
      currentMatch: {
        ...currentMatch,
        participants: updatedParticipants
      }
    })
  },

  updatePlayerScore: (heatNumber, score) => {
    const { currentMatch } = get()
    if (!currentMatch) return

    const updatedHeats = currentMatch.heats.map(heat =>
      heat.heatNumber === heatNumber
        ? { ...heat, playerScore: score }
        : heat
    )

    set({
      currentMatch: {
        ...currentMatch,
        heats: updatedHeats
      }
    })
  },

  setPlayerLocked: (heatNumber, locked) => {
    const { currentMatch } = get()
    if (!currentMatch) return

    const updatedHeats = currentMatch.heats.map(heat =>
      heat.heatNumber === heatNumber
        ? { ...heat, playerLocked: locked }
        : heat
    )

    set({
      currentMatch: {
        ...currentMatch,
        heats: updatedHeats
      }
    })
  },

  // Reset
  resetMatch: () => {
    set({
      currentMatch: null,
      isInMatch: false,
      error: null,
    })
  },

  resetQueue: () => {
    set({
      queueTicketId: null,
      isInQueue: false,
      estimatedWaitSeconds: 0,
      error: null,
    })
  },
}))