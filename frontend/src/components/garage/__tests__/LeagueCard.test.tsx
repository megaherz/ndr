import { describe, it, expect, vi } from 'vitest'
import { render, screen, fireEvent } from '@testing-library/react'
import { Decimal } from 'decimal.js'
import { LeagueCard } from '../LeagueCard'

describe('LeagueCard', () => {
  const defaultProps = {
    name: 'ROOKIE' as const,
    buyin: new Decimal('10.00'),
    available: true,
    fuelBalance: new Decimal('50.00'),
    onRaceNow: vi.fn(),
  }

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('should render league information correctly', () => {
    render(<LeagueCard {...defaultProps} />)

    expect(screen.getByText('ROOKIE')).toBeInTheDocument()
    expect(screen.getByText('LEARN THE ROPES')).toBeInTheDocument()
    expect(screen.getByText('10.00 FUEL')).toBeInTheDocument()
    expect(screen.getByText('50.00 FUEL')).toBeInTheDocument()
    expect(screen.getByText('Perfect for beginners. Max 3 races.')).toBeInTheDocument()
  })

  it('should show RACE NOW button when available and sufficient balance', () => {
    render(<LeagueCard {...defaultProps} />)

    const button = screen.getByRole('button')
    expect(button).toHaveTextContent('RACE NOW')
    expect(button).not.toBeDisabled()
  })

  it('should call onRaceNow when button is clicked', () => {
    const onRaceNow = vi.fn()
    render(<LeagueCard {...defaultProps} onRaceNow={onRaceNow} />)

    const button = screen.getByRole('button')
    fireEvent.click(button)

    expect(onRaceNow).toHaveBeenCalledTimes(1)
  })

  it('should show insufficient balance message when balance is too low', () => {
    const props = {
      ...defaultProps,
      fuelBalance: new Decimal('5.00'), // Less than 10.00 buyin
    }

    render(<LeagueCard {...props} />)

    const button = screen.getByRole('button')
    expect(button).toHaveTextContent('INSUFFICIENT FUEL')
    expect(button).toBeDisabled()
  })

  it('should show unavailable message when league is not available', () => {
    const props = {
      ...defaultProps,
      available: false,
      unavailable_reason: 'ROOKIE_LIMIT_REACHED',
    }

    render(<LeagueCard {...props} />)

    const button = screen.getByRole('button')
    expect(button).toHaveTextContent('ROOKIE LIMIT REACHED')
    expect(button).toBeDisabled()
  })

  it('should render different league configurations correctly', () => {
    const streetProps = {
      ...defaultProps,
      name: 'STREET' as const,
      buyin: new Decimal('50.00'),
    }

    render(<LeagueCard {...streetProps} />)

    expect(screen.getByText('STREET')).toBeInTheDocument()
    expect(screen.getByText('UNDERGROUND RACING')).toBeInTheDocument()
    expect(screen.getByText('Street-level competition with real stakes.')).toBeInTheDocument()
  })

  it('should apply disabled styling when card is disabled', () => {
    const props = {
      ...defaultProps,
      available: false,
    }

    const { container } = render(<LeagueCard {...props} />)
    const card = container.querySelector('.league-card')
    
    expect(card).toHaveClass('league-card--disabled')
  })

  it('should highlight insufficient balance in red', () => {
    const props = {
      ...defaultProps,
      fuelBalance: new Decimal('5.00'), // Less than buyin
    }

    const { container } = render(<LeagueCard {...props} />)
    const balanceValue = container.querySelector('.league-card__stat-value--insufficient')
    
    expect(balanceValue).toBeInTheDocument()
    expect(balanceValue).toHaveTextContent('5.00 FUEL')
  })
})