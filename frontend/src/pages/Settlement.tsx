import React from 'react'
import { useParams } from 'react-router-dom'

const Settlement: React.FC = () => {
  const { matchId } = useParams<{ matchId: string }>()

  return (
    <div className="settlement-page">
      <h1>Race Results</h1>
      <p>Match ID: {matchId}</p>
      <p>Final standings and prizes...</p>
      {/* Will be implemented in Phase 3: User Story 1 */}
    </div>
  )
}

export default Settlement