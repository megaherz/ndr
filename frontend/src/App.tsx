import { BrowserRouter as Router } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import './App.css'

const queryClient = new QueryClient()

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <Router>
        <div className="App">
          <h1>Nitro Drag Royale MVP</h1>
          <p>Telegram Mini App - Coming Soon!</p>
        </div>
      </Router>
    </QueryClientProvider>
  )
}

export default App