import React from 'react'
import { useParams } from 'react-router-dom'

const Race: React.FC = () => {
  const { matchId } = useParams<{ matchId: string }>()

  return (
    <div className="race-page">
      <h1>Race HUD</h1>
      <p>Match ID: {matchId}</p>
      <p>Race in progress...</p>
      {/* Will be implemented in Phase 3: User Story 1 with Pixi.js */}
    </div>
  )
}

export default Race