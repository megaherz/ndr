# Nitro Drag Royale - Frontend

React + TypeScript frontend for the Nitro Drag Royale Telegram Mini App.

## Development Setup

1. Install dependencies:
```bash
npm install
```

2. Copy environment variables:
```bash
cp .env.example .env
```

3. Start development server:
```bash
npm run dev
```

## Mobile Debugging with Eruda

Since Telegram Mini Apps don't provide direct access to the browser console, we use [Eruda](https://github.com/liriliri/eruda) for mobile debugging.

### Automatic Activation

Eruda is automatically enabled in the following scenarios:

1. **Development mode** (`npm run dev`)
2. **Debug URL parameter**: Add `?debug=true` or `?eruda=true` to the URL
3. **Environment variable**: Set `VITE_ENABLE_ERUDA=true` in `.env`
4. **Local storage**: Set `localStorage.setItem('ndr-debug', 'true')`

### Manual Control

Once the app is loaded, you can control Eruda programmatically:

```javascript
// Show debug console
NDR_DEBUG.show()

// Hide debug console
NDR_DEBUG.hide()

// Enable debug mode persistently
NDR_DEBUG.enable()

// Disable debug mode
NDR_DEBUG.disable()

// Log app information
NDR_DEBUG.info()
```

### Features Available

- **Console**: View console logs, errors, and warnings
- **Elements**: Inspect DOM structure and styles
- **Network**: Monitor API requests and responses
- **Resources**: View localStorage, sessionStorage, and cookies
- **Info**: Device and browser information
- **Sources**: View source code
- **Snippets**: Run JavaScript snippets

### Debugging in Telegram

1. Open the app in Telegram
2. Look for the Eruda floating icon (usually in bottom-right corner)
3. Tap the icon to open the debug console
4. Use the console to debug authentication, API calls, and UI issues

### Production Debugging

For production debugging, add `?debug=true` to the URL or use:

```javascript
localStorage.setItem('ndr-debug', 'true')
// Then reload the app
```

## Common Debug Scenarios

### Authentication Issues

```javascript
// Check authentication state
console.log('Auth State:', useAuthStore.getState())

// Check Telegram data
console.log('Telegram User:', window.Telegram?.WebApp?.initDataUnsafe?.user)

// Check initData
console.log('InitData:', window.Telegram?.WebApp?.initData)
```

### API Request Issues

```javascript
// Monitor network tab in Eruda for failed requests
// Check API base URL
console.log('API Base URL:', import.meta.env.VITE_API_BASE_URL)

// Test API connectivity
fetch('/api/v1/garage').then(console.log).catch(console.error)
```

### Telegram WebApp Issues

```javascript
// Check Telegram WebApp availability
console.log('Telegram WebApp:', window.Telegram?.WebApp)

// Check theme parameters
console.log('Theme:', window.Telegram?.WebApp?.themeParams)

// Check viewport
console.log('Viewport:', {
  height: window.Telegram?.WebApp?.viewportHeight,
  stableHeight: window.Telegram?.WebApp?.viewportStableHeight
})
```

## Build and Deployment

```bash
# Build for production
npm run build

# Preview production build
npm run preview

# Type checking
npm run type-check

# Linting
npm run lint
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `VITE_API_URL` | Backend API base URL | `http://localhost:8080` |
| `VITE_CENTRIFUGO_URL` | Centrifugo WebSocket URL | `ws://localhost:8000/connection/websocket` |
| `VITE_ENABLE_ERUDA` | Force enable Eruda debug console | `false` |
| `VITE_APP_VERSION` | App version for debugging | `1.0.0` |
| `VITE_AMPLITUDE_API_KEY` | Analytics API key | - |
| `VITE_ENVIRONMENT` | Environment name | `development` |

## Project Structure

```
src/
├── components/          # React components
├── pages/              # Top-level page components
├── services/           # Business logic layer
│   ├── api/           # HTTP API clients
│   ├── auth/          # Authentication services
│   ├── centrifugo/    # Real-time communication
│   ├── debug/         # Debug utilities (Eruda)
│   ├── telegram/      # Telegram WebApp integration
│   └── wallet/        # TON wallet integration
├── stores/            # Zustand state management
├── hooks/             # Custom React hooks
└── utils/             # Utility functions
```

## Troubleshooting

### Eruda Not Showing

1. Check browser console for errors
2. Verify environment variables
3. Try manual activation: `NDR_DEBUG.show()`
4. Clear localStorage and reload

### Authentication Failures

1. Check Telegram WebApp initialization
2. Verify initData extraction
3. Check API endpoint connectivity
4. Review token storage and expiration

### Network Issues

1. Check CORS configuration
2. Verify API base URL
3. Monitor network requests in Eruda
4. Test API endpoints directly

For more detailed debugging, use the Eruda console and the global `NDR_DEBUG` object.