import { useTonConnectUI, useTonWallet, useTonAddress } from '@tonconnect/ui-react'
import { useCallback, useEffect } from 'react'
import { useWalletStore } from '../../stores/walletStore'

// TON Connect configuration
export const TON_CONNECT_CONFIG = {
  manifestUrl: import.meta.env.VITE_TON_CONNECT_MANIFEST_URL || 'https://your-domain.com/tonconnect-manifest.json',
  buttonRootId: 'ton-connect-button',
  uiPreferences: {
    theme: 'DARK' as const,
    borderRadius: 's' as const,
  },
}

// Hook for TON wallet connection management
export const useTonWalletConnection = () => {
  const [tonConnectUI] = useTonConnectUI()
  const wallet = useTonWallet()
  const address = useTonAddress()
  const { setTonWalletAddress } = useWalletStore()

  // Update wallet address in store when connection changes
  useEffect(() => {
    if (address) {
      setTonWalletAddress(address)
    } else {
      setTonWalletAddress(null)
    }
  }, [address, setTonWalletAddress])

  const connectWallet = useCallback(async () => {
    try {
      await tonConnectUI.connectWallet()
    } catch (error) {
      console.error('Failed to connect TON wallet:', error)
      throw error
    }
  }, [tonConnectUI])

  const disconnectWallet = useCallback(async () => {
    try {
      await tonConnectUI.disconnect()
      setTonWalletAddress(null)
    } catch (error) {
      console.error('Failed to disconnect TON wallet:', error)
      throw error
    }
  }, [tonConnectUI, setTonWalletAddress])

  return {
    wallet,
    address,
    isConnected: !!wallet,
    connectWallet,
    disconnectWallet,
  }
}

// Transaction types for TON operations
export interface TonTransaction {
  to: string
  value: string // In nanotons
  body?: string
  stateInit?: string
}

// Hook for sending TON transactions
export const useTonTransactions = () => {
  const [tonConnectUI] = useTonConnectUI()
  const wallet = useTonWallet()

  const sendTransaction = useCallback(async (transaction: TonTransaction) => {
    if (!wallet) {
      throw new Error('Wallet not connected')
    }

    try {
      const result = await tonConnectUI.sendTransaction({
        validUntil: Math.floor(Date.now() / 1000) + 600, // 10 minutes
        messages: [
          {
            address: transaction.to,
            amount: transaction.value,
            payload: transaction.body,
            stateInit: transaction.stateInit,
          },
        ],
      })

      return result
    } catch (error) {
      console.error('Transaction failed:', error)
      throw error
    }
  }, [tonConnectUI, wallet])

  return {
    sendTransaction,
    isWalletConnected: !!wallet,
  }
}

// Utility functions for TON amount conversion
export const tonToNanoton = (ton: string): string => {
  const tonAmount = parseFloat(ton)
  if (isNaN(tonAmount)) {
    throw new Error('Invalid TON amount')
  }
  return (tonAmount * 1_000_000_000).toString()
}

export const nanotonToTon = (nanoton: string): string => {
  const nanotonAmount = parseInt(nanoton, 10)
  if (isNaN(nanotonAmount)) {
    throw new Error('Invalid nanoton amount')
  }
  return (nanotonAmount / 1_000_000_000).toFixed(9)
}

// Format TON address for display
export const formatTonAddress = (address: string, length: number = 8): string => {
  if (!address) return ''
  if (address.length <= length * 2) return address
  return `${address.slice(0, length)}...${address.slice(-length)}`
}

// Validate TON address format
export const isValidTonAddress = (address: string): boolean => {
  // Basic validation - TON addresses are typically 48 characters
  // This is a simplified validation, you might want to use a proper TON address validator
  const tonAddressRegex = /^[A-Za-z0-9_-]{48}$/
  return tonAddressRegex.test(address)
}

// TON Connect manifest configuration
export const createTonConnectManifest = () => ({
  url: window.location.origin,
  name: 'Nitro Drag Royale',
  iconUrl: `${window.location.origin}/icon-512x512.png`,
  termsOfUseUrl: `${window.location.origin}/terms`,
  privacyPolicyUrl: `${window.location.origin}/privacy`,
})

// Hook for wallet balance monitoring
export const useTonBalance = () => {
  const wallet = useTonWallet()
  const address = useTonAddress()

  // This would typically fetch balance from TON blockchain
  // For now, we'll return a placeholder
  const getBalance = useCallback(async (): Promise<string> => {
    if (!address) {
      return '0'
    }

    try {
      // In a real implementation, you would call TON API here
      // For example: const response = await fetch(`https://toncenter.com/api/v2/getAddressBalance?address=${address}`)
      // const data = await response.json()
      // return data.result
      
      // Placeholder - return mock balance
      return '0'
    } catch (error) {
      console.error('Failed to fetch TON balance:', error)
      return '0'
    }
  }, [address])

  return {
    address,
    getBalance,
    isConnected: !!wallet,
  }
}

// Error handling for TON Connect operations
export class TonConnectError extends Error {
  public code: string

  constructor(message: string, code: string = 'TON_CONNECT_ERROR') {
    super(message)
    this.name = 'TonConnectError'
    this.code = code
  }
}

// Common error codes
export const TON_CONNECT_ERROR_CODES = {
  USER_REJECTED: 'USER_REJECTED_ERROR',
  WALLET_NOT_CONNECTED: 'WALLET_NOT_CONNECTED_ERROR',
  TRANSACTION_FAILED: 'TRANSACTION_FAILED_ERROR',
  INVALID_ADDRESS: 'INVALID_ADDRESS_ERROR',
  INSUFFICIENT_BALANCE: 'INSUFFICIENT_BALANCE_ERROR',
} as const

// Hook for handling TON Connect errors
export const useTonConnectErrorHandler = () => {
  const handleError = useCallback((error: any): TonConnectError => {
    if (error instanceof TonConnectError) {
      return error
    }

    // Map common TON Connect errors
    if (error?.message?.includes('user rejected')) {
      return new TonConnectError('User rejected the transaction', TON_CONNECT_ERROR_CODES.USER_REJECTED)
    }

    if (error?.message?.includes('wallet not connected')) {
      return new TonConnectError('Wallet not connected', TON_CONNECT_ERROR_CODES.WALLET_NOT_CONNECTED)
    }

    if (error?.message?.includes('insufficient balance')) {
      return new TonConnectError('Insufficient balance', TON_CONNECT_ERROR_CODES.INSUFFICIENT_BALANCE)
    }

    // Generic error
    return new TonConnectError(error?.message || 'Unknown TON Connect error', 'UNKNOWN_ERROR')
  }, [])

  return { handleError }
}