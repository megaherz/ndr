import React from 'react'
import { useQuery } from '@tanstack/react-query'
import { useNavigate } from 'react-router-dom'
import { Decimal } from 'decimal.js'
import { LeagueCard } from '../components/garage/LeagueCard'
import { garageQueries } from '../services/api/garage'
import { useWalletStore } from '../stores/walletStore'
import { useAuthStore } from '../stores/authStore'
import './Garage.css'

const Garage: React.FC = () => {
  const navigate = useNavigate()
  const { setBalances, setLeagueAccess } = useWalletStore()
  const { isAuthenticated, isTokenExpired } = useAuthStore()

  // Only fetch garage state when user is authenticated
  const shouldFetch = isAuthenticated && !isTokenExpired()

  // Fetch garage state using React Query
  const {
    data: garageData,
    isLoading,
    error,
    refetch
  } = useQuery({
    ...garageQueries.stateOptions(),
    enabled: shouldFetch, // Only run query when authenticated
  })

  // Update wallet store when garage data changes
  React.useEffect(() => {
    if (garageData) {
      // Update wallet balances
      setBalances({
        tonBalance: new Decimal(0), // TON balance not included in garage response
        fuelBalance: garageData.wallet.fuelBalance,
        burnBalance: garageData.wallet.burnBalance,
        rookieRacesCompleted: garageData.wallet.rookieRacesCompleted,
      })

      // Update league access based on availability
      setLeagueAccess({
        rookie: garageData.leagues.find(l => l.name === 'ROOKIE')?.available ?? false,
        street: garageData.leagues.find(l => l.name === 'STREET')?.available ?? false,
        pro: garageData.leagues.find(l => l.name === 'PRO')?.available ?? false,
        topFuel: garageData.leagues.find(l => l.name === 'TOP_FUEL')?.available ?? false,
      })
    }
  }, [garageData, setBalances, setLeagueAccess])

  const handleRaceNow = (leagueName: string) => {
    console.log(`Starting race in ${leagueName} league`)
    // Navigate to matchmaking with the selected league
    navigate('/matchmaking', { state: { league: leagueName } })
  }

  const handleGasStationClick = () => {
    navigate('/gas-station')
  }

  // Show loading while authentication is in progress or data is loading
  if (!shouldFetch || isLoading) {
    return (
      <div className="garage">
        <div className="garage__loading">
          <div className="garage__loading-spinner" />
          <h2>{!shouldFetch ? 'Authenticating...' : 'Loading Garage...'}</h2>
          <p>{!shouldFetch ? 'Verifying your credentials' : 'Preparing your racing experience'}</p>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="garage">
        <div className="garage__error">
          <div className="garage__error-icon">⚠️</div>
          <h2>Garage Unavailable</h2>
          <p>Failed to load garage data. Please try again.</p>
          <button 
            className="garage__retry-button"
            onClick={() => refetch()}
          >
            Retry
          </button>
        </div>
      </div>
    )
  }

  if (!garageData) {
    return (
      <div className="garage">
        <div className="garage__error">
          <div className="garage__error-icon">🚫</div>
          <h2>No Data Available</h2>
          <p>Unable to load garage information.</p>
        </div>
      </div>
    )
  }

  return (
    <div className="garage">
      {/* Header */}
      <header className="garage__header">
        <div className="garage__title-section">
          <h1 className="garage__title">
            <span className="garage__title-main">GARAGE</span>
            <span className="garage__title-sub">SELECT YOUR LEAGUE</span>
          </h1>
          <div className="garage__racing-stripes">
            <div className="garage__stripe garage__stripe--1" />
            <div className="garage__stripe garage__stripe--2" />
            <div className="garage__stripe garage__stripe--3" />
          </div>
        </div>

        {/* User info */}
        <div className="garage__user-info">
          <div className="garage__user-avatar">
            {garageData.user.display_name.charAt(0).toUpperCase()}
          </div>
          <div className="garage__user-details">
            <h3 className="garage__user-name">{garageData.user.display_name}</h3>
            <p className="garage__user-id">ID: {garageData.user.id.slice(0, 8)}...</p>
          </div>
        </div>
      </header>

      {/* Balance display */}
      <section className="garage__balance-section">
        <div className="garage__balance-card">
          <div className="garage__balance-header">
            <h3 className="garage__balance-title">YOUR WALLET</h3>
            <button 
              className="garage__gas-station-button"
              onClick={handleGasStationClick}
              title="Go to Gas Station for deposits/withdrawals"
            >
              ⛽ GAS STATION
            </button>
          </div>
          
          <div className="garage__balances">
            <div className="garage__balance-item garage__balance-item--fuel">
              <span className="garage__balance-label">FUEL</span>
              <span className="garage__balance-value">
                {garageData.wallet.fuelBalance.toFixed(2)}
              </span>
            </div>
            
            <div className="garage__balance-item garage__balance-item--burn">
              <span className="garage__balance-label">BURN</span>
              <span className="garage__balance-value">
                {garageData.wallet.burnBalance.toFixed(2)}
              </span>
            </div>
          </div>

          {/* Rookie races progress */}
          {garageData.wallet.rookieRacesCompleted > 0 && (
            <div className="garage__rookie-progress">
              <span className="garage__rookie-label">ROOKIE RACES</span>
              <div className="garage__rookie-counter">
                {garageData.wallet.rookieRacesCompleted} / 3
              </div>
            </div>
          )}
        </div>
      </section>

      {/* League selection */}
      <section className="garage__leagues-section">
        <h2 className="garage__leagues-title">CHOOSE YOUR LEAGUE</h2>
        
        <div className="garage__leagues-grid">
          {garageData.leagues.map((league) => (
            <LeagueCard
              key={league.name}
              name={league.name}
              buyin={league.buyin}
              available={league.available}
              unavailable_reason={league.unavailable_reason}
              fuelBalance={garageData.wallet.fuelBalance}
              onRaceNow={() => handleRaceNow(league.name)}
            />
          ))}
        </div>
      </section>

      {/* Footer info */}
      <footer className="garage__footer">
        <p className="garage__footer-text">
          Ready to race? Select a league and hit the track! 🏁
        </p>
      </footer>
    </div>
  )
}

export default Garage