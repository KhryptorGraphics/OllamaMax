import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App'
import './index.css'

// Import accessibility provider
import { AccessibilityProvider } from '@/components/accessibility'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <AccessibilityProvider>
      <App />
    </AccessibilityProvider>
  </React.StrictMode>
)