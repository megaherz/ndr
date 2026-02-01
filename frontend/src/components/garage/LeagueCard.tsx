import React from 'react'
import { Decimal } from 'decimal.js'
import './LeagueCard.css'

export interface LeagueCardProps {
  name: 'ROOKIE' | 'STREET' | 'PRO' | 'TOP_FUEL'
  buyin: Decimal
  available: boolean
  unavailable_reason?: string | null
  fuelBalance: Decimal
  onRaceNow: () => void
}

const LEAGUE_CONFIG = {
  ROOKIE: {
    displayName: 'ROOKIE',
    subtitle: 'LEARN THE ROPES',
    color: '#00ff41', // Matrix green
    bgGradient: 'linear-gradient(135deg, #001a0a 0%, #003d1a 50%, #00ff41 100%)',
    borderColor: '#00ff41',
    shadowColor: '#00ff4140',
    icon: '🏁',
    description: 'Perfect for beginners. Max 3 races.',
  },
  STREET: {
    displayName: 'STREET',
    subtitle: 'UNDERGROUND RACING',
    color: '#ff6b35', // Vibrant orange
    bgGradient: 'linear-gradient(135deg, #1a0f00 0%, #3d2200 50%, #ff6b35 100%)',
    borderColor: '#ff6b35',
    shadowColor: '#ff6b3540',
    icon: '🔥',
    description: 'Street-level competition with real stakes.',
  },
  PRO: {
    displayName: 'PRO',
    subtitle: 'PROFESSIONAL CIRCUIT',
    color: '#3b82f6', // Electric blue
    bgGradient: 'linear-gradient(135deg, #0a0f1a 0%, #1e3a8a 50%, #3b82f6 100%)',
    borderColor: '#3b82f6',
    shadowColor: '#3b82f640',
    icon: '⚡',
    description: 'High-stakes professional racing.',
  },
  TOP_FUEL: {
    displayName: 'TOP FUEL',
    subtitle: 'ELITE CHAMPIONSHIP',
    color: '#fbbf24', // Gold
    bgGradient: 'linear-gradient(135deg, #1a1000 0%, #3d2f00 50%, #fbbf24 100%)',
    borderColor: '#fbbf24',
    shadowColor: '#fbbf2440',
    icon: '👑',
    description: 'The ultimate racing challenge.',
  },
} as const

export const LeagueCard: React.FC<LeagueCardProps> = ({
  name,
  buyin,
  available,
  unavailable_reason,
  fuelBalance,
  onRaceNow,
}) => {
  const config = LEAGUE_CONFIG[name]
  const hasInsufficientBalance = fuelBalance.lessThan(buyin)
  const isDisabled = !available || hasInsufficientBalance

  const getUnavailableMessage = (): string => {
    if (hasInsufficientBalance) {
      return 'INSUFFICIENT FUEL'
    }
    if (unavailable_reason === 'ROOKIE_LIMIT_REACHED') {
      return 'ROOKIE LIMIT REACHED'
    }
    return unavailable_reason || 'UNAVAILABLE'
  }

  const handleRaceClick = () => {
    if (!isDisabled) {
      onRaceNow()
    }
  }

  return (
    <div 
      className={`league-card ${isDisabled ? 'league-card--disabled' : ''}`}
      style={{
        '--league-color': config.color,
        '--league-bg-gradient': config.bgGradient,
        '--league-border-color': config.borderColor,
        '--league-shadow-color': config.shadowColor,
      } as React.CSSProperties}
    >
      {/* Racing stripe decoration */}
      <div className="league-card__stripe" />
      
      {/* Header */}
      <div className="league-card__header">
        <div className="league-card__icon">{config.icon}</div>
        <div className="league-card__title-group">
          <h3 className="league-card__title">{config.displayName}</h3>
          <p className="league-card__subtitle">{config.subtitle}</p>
        </div>
      </div>

      {/* Stats */}
      <div className="league-card__stats">
        <div className="league-card__stat">
          <span className="league-card__stat-label">BUY-IN</span>
          <span className="league-card__stat-value">{buyin.toFixed(2)} FUEL</span>
        </div>
        <div className="league-card__stat">
          <span className="league-card__stat-label">YOUR BALANCE</span>
          <span className={`league-card__stat-value ${hasInsufficientBalance ? 'league-card__stat-value--insufficient' : ''}`}>
            {fuelBalance.toFixed(2)} FUEL
          </span>
        </div>
      </div>

      {/* Description */}
      <p className="league-card__description">{config.description}</p>

      {/* Action button */}
      <button
        className={`league-card__button ${isDisabled ? 'league-card__button--disabled' : ''}`}
        onClick={handleRaceClick}
        disabled={isDisabled}
      >
        {isDisabled ? (
          <span className="league-card__button-text">
            {getUnavailableMessage()}
          </span>
        ) : (
          <>
            <span className="league-card__button-text">RACE NOW</span>
            <span className="league-card__button-icon">🏎️</span>
          </>
        )}
      </button>

      {/* Disabled overlay */}
      {isDisabled && <div className="league-card__overlay" />}
    </div>
  )
}