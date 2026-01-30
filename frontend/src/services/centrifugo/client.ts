import { Centrifuge, Subscription, PublicationContext, JoinContext, LeaveContext } from 'centrifuge'
import { useAuthStore } from '../../stores/authStore'
import { useMatchStore } from '../../stores/matchStore'

// Centrifugo configuration
const CENTRIFUGO_URL = import.meta.env.VITE_CENTRIFUGO_URL || 'ws://localhost:8000/connection/websocket'

// Event types
export interface CentrifugoEvent {
  event: string
  data: unknown
  timestamp: number
}

// Event handlers type
export type EventHandler = (data: unknown) => void

// Client wrapper class
class CentrifugoClient {
  private centrifuge: Centrifuge | null = null
  private subscriptions: Map<string, Subscription> = new Map()
  private eventHandlers: Map<string, EventHandler[]> = new Map()
  private isConnected = false
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5

  constructor() {
    this.setupEventHandlers()
  }

  // Initialize connection
  async connect(): Promise<void> {
    const { tokens, isTokenExpired } = useAuthStore.getState()
    
    if (!tokens || isTokenExpired()) {
      throw new Error('No valid authentication token available')
    }

    if (this.centrifuge) {
      this.disconnect()
    }

    this.centrifuge = new Centrifuge(CENTRIFUGO_URL, {
      token: tokens.centrifugoToken,
      debug: import.meta.env.DEV,
    })

    this.setupConnectionHandlers()
    
    try {
      await this.centrifuge.connect()
      this.isConnected = true
      this.reconnectAttempts = 0
      
      // Update match store connection status
      useMatchStore.getState().setConnected(true)
      
      console.log('Connected to Centrifugo')
    } catch (error) {
      console.error('Failed to connect to Centrifugo:', error)
      throw error
    }
  }

  // Disconnect from Centrifugo
  disconnect(): void {
    if (this.centrifuge) {
      // Unsubscribe from all channels
      this.subscriptions.forEach((sub) => {
        sub.unsubscribe()
      })
      this.subscriptions.clear()

      this.centrifuge.disconnect()
      this.centrifuge = null
    }

    this.isConnected = false
    useMatchStore.getState().setConnected(false)
    console.log('Disconnected from Centrifugo')
  }

  // Setup connection event handlers
  private setupConnectionHandlers(): void {
    if (!this.centrifuge) return

    this.centrifuge.on('connected', () => {
      console.log('Centrifugo connection established')
      this.isConnected = true
      this.reconnectAttempts = 0
      useMatchStore.getState().setConnected(true)
    })

    this.centrifuge.on('disconnected', (ctx) => {
      console.log('Centrifugo disconnected:', ctx)
      this.isConnected = false
      useMatchStore.getState().setConnected(false)
      
      // Attempt reconnection
      this.handleReconnection()
    })

    this.centrifuge.on('error', (ctx) => {
      console.error('Centrifugo error:', ctx)
      useMatchStore.getState().setError(`Connection error: ${ctx.error}`)
    })
  }

  // Handle reconnection logic
  private async handleReconnection(): Promise<void> {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnection attempts reached')
      useMatchStore.getState().setError('Connection lost. Please refresh the page.')
      return
    }

    this.reconnectAttempts++
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 10000)
    
    console.log(`Attempting to reconnect in ${delay}ms (attempt ${this.reconnectAttempts})`)
    
    setTimeout(async () => {
      try {
        await this.connect()
      } catch (error) {
        console.error('Reconnection failed:', error)
      }
    }, delay)
  }

  // Subscribe to a channel
  async subscribe(channel: string): Promise<void> {
    if (!this.centrifuge || !this.isConnected) {
      throw new Error('Not connected to Centrifugo')
    }

    if (this.subscriptions.has(channel)) {
      console.log(`Already subscribed to channel: ${channel}`)
      return
    }

    const subscription = this.centrifuge.newSubscription(channel)

    subscription.on('publication', (ctx: PublicationContext) => {
      this.handleEvent(channel, ctx.data)
    })

    subscription.on('join', (ctx: JoinContext) => {
      console.log(`User joined channel ${channel}:`, ctx.info)
    })

    subscription.on('leave', (ctx: LeaveContext) => {
      console.log(`User left channel ${channel}:`, ctx.info)
    })

    subscription.on('error', (ctx) => {
      console.error(`Subscription error for channel ${channel}:`, ctx)
    })

    try {
      await subscription.subscribe()
      this.subscriptions.set(channel, subscription)
      console.log(`Subscribed to channel: ${channel}`)
    } catch (error) {
      console.error(`Failed to subscribe to channel ${channel}:`, error)
      throw error
    }
  }

  // Unsubscribe from a channel
  async unsubscribe(channel: string): Promise<void> {
    const subscription = this.subscriptions.get(channel)
    if (subscription) {
      await subscription.unsubscribe()
      this.subscriptions.delete(channel)
      console.log(`Unsubscribed from channel: ${channel}`)
    }
  }

  // Handle incoming events
  private handleEvent(channel: string, data: unknown): void {
    try {
      const event = data as CentrifugoEvent
      console.log(`Received event on ${channel}:`, event)

      // Update last event timestamp
      useMatchStore.getState().setLastEventTimestamp(event.timestamp)

      // Call registered event handlers
      const handlers = this.eventHandlers.get(event.event) || []
      handlers.forEach(handler => {
        try {
          handler(event.data)
        } catch (error) {
          console.error(`Error in event handler for ${event.event}:`, error)
        }
      })

      // Built-in event handlers
      this.handleBuiltInEvents(event).catch(error => {
        console.error('Error in built-in event handler:', error)
      }).catch(console.error)
    } catch (error) {
      console.error('Error handling event:', error)
    }
  }

  // Handle built-in events that update stores
  private async handleBuiltInEvents(event: CentrifugoEvent): Promise<void> {
    const matchStore = useMatchStore.getState()
    const { Decimal } = await import('decimal.js')
    
    // Type guard for event data
    const data = event.data as Record<string, unknown>
    
    switch (event.event) {
      case 'match_found':
        // Navigate to race when match is found
        matchStore.setCurrentMatch({
          id: data.match_id as string,
          league: data.league as string,
          status: 'forming',
          participants: [],
          currentHeat: 1,
          heats: [
            { heatNumber: 1, status: 'waiting', duration: 25, playerLocked: false },
            { heatNumber: 2, status: 'waiting', duration: 25, playerLocked: false },
            { heatNumber: 3, status: 'waiting', duration: 25, playerLocked: false },
          ],
          prizePool: new Decimal(0),
        })
        matchStore.resetQueue()
        break

      case 'heat_started':
        matchStore.updateHeatData(data.heat_number as number, {
          status: 'active',
          startedAt: Date.now(),
          targetLine: data.target_line ? 
            new Decimal(data.target_line as string) : 
            undefined,
        })
        break

      case 'heat_ended':
        matchStore.updateHeatData(data.heat_number as number, {
          status: 'completed',
        })
        
        // Update standings
        if (data.standings && Array.isArray(data.standings)) {
          data.standings.forEach((standing: unknown) => {
            const standingData = standing as Record<string, unknown>
            if (standingData.user_id) {
              matchStore.updateParticipant(standingData.user_id as string, {
                currentPosition: standingData.position as number,
                totalScore: new Decimal(standingData.total_score as string),
              })
            }
          })
        }
        break

      case 'match_settled':
        // Update final standings and navigate to settlement
        if (data.final_standings && Array.isArray(data.final_standings)) {
          data.final_standings.forEach((standing: unknown) => {
            const standingData = standing as Record<string, unknown>
            if (standingData.user_id) {
              matchStore.updateParticipant(standingData.user_id as string, {
                finalPosition: standingData.position as number,
                prizeAmount: new Decimal(standingData.prize_amount as string),
                burnReward: new Decimal(standingData.burn_reward as string),
              })
            }
          })
        }
        break

      case 'balance_updated': {
        // Update wallet balances
        const walletStore = (await import('../../stores/walletStore')).useWalletStore.getState()
        walletStore.setBalances({
          tonBalance: new Decimal(data.ton_balance as string),
          fuelBalance: new Decimal(data.fuel_balance as string),
          burnBalance: new Decimal(data.burn_balance as string),
          rookieRacesCompleted: walletStore.balances?.rookieRacesCompleted || 0,
        })
        break
      }
    }
  }

  // RPC call wrapper
  async rpc(method: string, data: unknown): Promise<unknown> {
    if (!this.centrifuge || !this.isConnected) {
      throw new Error('Not connected to Centrifugo')
    }

    try {
      const result = await this.centrifuge.rpc(method, data)
      return result.data
    } catch (error) {
      console.error(`RPC call failed for ${method}:`, error)
      throw error
    }
  }

  // Register event handler
  on(event: string, handler: EventHandler): void {
    if (!this.eventHandlers.has(event)) {
      this.eventHandlers.set(event, [])
    }
    this.eventHandlers.get(event)!.push(handler)
  }

  // Unregister event handler
  off(event: string, handler: EventHandler): void {
    const handlers = this.eventHandlers.get(event)
    if (handlers) {
      const index = handlers.indexOf(handler)
      if (index > -1) {
        handlers.splice(index, 1)
      }
    }
  }

  // Setup default event handlers
  private setupEventHandlers(): void {
    // These will be implemented in Phase 3 when we build the actual game logic
  }

  // Utility methods
  isChannelSubscribed(channel: string): boolean {
    return this.subscriptions.has(channel)
  }

  getConnectionStatus(): boolean {
    return this.isConnected
  }

  getSubscribedChannels(): string[] {
    return Array.from(this.subscriptions.keys())
  }
}

// Export singleton instance
export const centrifugoClient = new CentrifugoClient()

// Convenience functions for common operations
export const subscribeToUserChannel = async (userId: string) => {
  await centrifugoClient.subscribe(`user:${userId}`)
}

export const subscribeToMatchChannel = async (matchId: string) => {
  await centrifugoClient.subscribe(`match:${matchId}`)
}

export const unsubscribeFromUserChannel = async (userId: string) => {
  await centrifugoClient.unsubscribe(`user:${userId}`)
}

export const unsubscribeFromMatchChannel = async (matchId: string) => {
  await centrifugoClient.unsubscribe(`match:${matchId}`)
}

// RPC convenience functions
export const joinMatchmaking = async (league: string, clientReqId: string) => {
  return centrifugoClient.rpc('matchmaking.join', {
    league,
    client_req_id: clientReqId,
  })
}

export const cancelMatchmaking = async (queueTicketId: string, clientReqId: string) => {
  return centrifugoClient.rpc('matchmaking.cancel', {
    queue_ticket_id: queueTicketId,
    client_req_id: clientReqId,
  })
}

export const earnPoints = async (matchId: string, heatNumber: number, clientReqId: string) => {
  return centrifugoClient.rpc('match.earn_points', {
    match_id: matchId,
    heat_number: heatNumber,
    client_req_id: clientReqId,
  })
}

export const giveUp = async (matchId: string, clientReqId: string) => {
  return centrifugoClient.rpc('match.give_up', {
    match_id: matchId,
    client_req_id: clientReqId,
  })
}