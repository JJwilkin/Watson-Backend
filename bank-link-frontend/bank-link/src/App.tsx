import { useState, useEffect } from 'react'
import './App.css'
import TellerLink from './components/TellerLink'
import PlaidLink from './components/PlaidLink'

type LinkProvider = 'teller' | 'plaid' | null;

function App() {
  const [jwt, setJwt] = useState<string | null>(null)
  const [selectedProvider, setSelectedProvider] = useState<LinkProvider>(null)
  const [status, setStatus] = useState<'loading' | 'success' | 'error' | null>(null)
  const [message, setMessage] = useState<string>('')

  useEffect(() => {
    // Extract JWT from URL parameters
    const urlParams = new URLSearchParams(window.location.search)
    const jwtFromUrl = urlParams.get('jwt')
    
    // Validate JWT in the backend before setting it
    const validateJWT = async () => {
      const url = import.meta.env.VITE_BACKEND_URL || 'http://0.0.0.0:8080';
      try {
        const resp = await fetch(`${url}/validate-jwt`, {
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${jwtFromUrl}`,
            'Content-Type': 'application/json',
          },
        });
        if (resp.ok) {
          setJwt(jwtFromUrl);
        } else {
          setMessage('Invalid or expired session. Please log in again.');
        }
      } catch (err) {
        setMessage('Error validating session. Please try again.');
      }
    }
    if (jwtFromUrl) {
      validateJWT()
    }
  }, [])

  const handleTellerSuccess = async (authorization: any) => {
    setStatus('loading')
    setMessage('Connecting your bank account via Teller...')
    
    const url = import.meta.env.VITE_BACKEND_URL || 'http://0.0.0.0:8080'
    const response = await fetch(url + '/bank-link-teller/success', {
      method: 'POST',
      body: JSON.stringify(authorization),
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${jwt}`
      }
    })
    
    if (response.ok) {
      const data = await response.json()
      console.log('Teller connection successful:', data)
      setStatus('success')
      setMessage('Bank account connected successfully! You can close this window now.')
    } else {
      console.error('Failed to connect bank account via Teller')
      setStatus('error')
      setMessage('Failed to connect bank account via Teller. Please try again.')
    }
  }

  const handlePlaidSuccess = async (publicToken: string, metadata: any) => {
    setStatus('loading')
    setMessage('Connecting your bank account via Plaid...')
    
    const url = import.meta.env.VITE_BACKEND_URL || 'http://0.0.0.0:8080'
    const response = await fetch(url + '/bank-link-plaid/success', {
      method: 'POST',
      body: JSON.stringify({ public_token: publicToken, metadata }),
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${jwt}`
      }
    })
    
    if (response.ok) {
      const data = await response.json()
      console.log('Plaid connection successful:', data)
      setStatus('success')
      setMessage('Bank account connected successfully! You can close this window now.')
    } else {
      console.error('Failed to connect bank account via Plaid')
      setStatus('error')
      setMessage('Failed to connect bank account via Plaid. Please try again.')
    }
  }

  const handlePlaidExit = (error: any, metadata: any) => {
    console.log('Plaid Link exited:', { error, metadata })
    if (error) {
      setStatus('error')
      setMessage('Bank connection was cancelled or failed. Please try again.')
    }
  }

  const resetToProviderSelection = () => {
    setSelectedProvider(null)
    setStatus(null)
    setMessage('')
  }

  // Provider selection screen
  if (jwt && !selectedProvider && status === null) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-50">
        <div className="bg-white p-8 rounded-lg shadow-lg max-w-md w-full mx-4">
          <h2 className="text-2xl font-bold text-gray-800 mb-6 text-center">
            Connect Your Bank Account
          </h2>
          <p className="text-gray-600 mb-6 text-center">
            Choose your preferred bank connection service:
          </p>
          
          <div className="space-y-4">
            <button
              onClick={() => setSelectedProvider('teller')}
              className="w-full p-4 border-2 border-gray-200 rounded-lg hover:border-indigo-500 hover:bg-indigo-50 transition-colors"
            >
              <div className="flex items-center justify-center space-x-3">
                <span className="text-2xl">🏦</span>
                <div className="text-left">
                  <h3 className="font-semibold text-gray-800">Teller</h3>
                  <p className="text-sm text-gray-600">Secure bank connection via Teller</p>
                </div>
              </div>
            </button>
            
            <button
              onClick={() => setSelectedProvider('plaid')}
              className="w-full p-4 border-2 border-gray-200 rounded-lg hover:border-blue-500 hover:bg-blue-50 transition-colors"
            >
              <div className="flex items-center justify-center space-x-3">
                <span className="text-2xl">💳</span>
                <div className="text-left">
                  <h3 className="font-semibold text-gray-800">Plaid</h3>
                  <p className="text-sm text-gray-600">Secure bank connection via Plaid</p>
                </div>
              </div>
            </button>
          </div>
        </div>
      </div>
    )
  }

  return (
    <>
      {/* Teller Link */}
      {jwt && selectedProvider === 'teller' && status === null && (
        <TellerLink onSuccess={handleTellerSuccess} />
      )}
      
      {/* Plaid Link */}
      {jwt && selectedProvider === 'plaid' && status === null && (
        <PlaidLink 
          onSuccess={handlePlaidSuccess}
          onExit={handlePlaidExit}
          jwt={jwt}
        />
      )}
      
      {/* Loading State */}
      {status === 'loading' && (
        <div className="flex items-center justify-center min-h-screen">
          <div className="text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-indigo-600 mx-auto mb-4"></div>
            <p className="text-gray-600">{message}</p>
          </div>
        </div>
      )}
      
      {/* Success State */}
      {status === 'success' && (
        <div className="flex items-center justify-center min-h-screen">
          <div className="text-center">
            <div className="text-green-600 text-6xl mb-4">✓</div>
            <h2 className="text-2xl font-bold text-gray-800 mb-2">Success!</h2>
            <p className="text-gray-600 mb-4">{message}</p>
            <button 
              onClick={resetToProviderSelection}
              className="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700"
            >
              Connect Another Account
            </button>
          </div>
        </div>
      )}
      
      {/* Error State */}
      {status === 'error' && (
        <div className="flex items-center justify-center min-h-screen">
          <div className="text-center">
            <div className="text-red-600 text-6xl mb-4">✗</div>
            <h2 className="text-2xl font-bold text-gray-800 mb-2">Error</h2>
            <p className="text-gray-600 mb-4">{message}</p>
            <div className="space-x-4">
              <button 
                onClick={resetToProviderSelection}
                className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
              >
                Try Again
              </button>
              <button 
                onClick={() => window.close()}
                className="px-4 py-2 bg-gray-600 text-white rounded hover:bg-gray-700"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  )
}

export default App
