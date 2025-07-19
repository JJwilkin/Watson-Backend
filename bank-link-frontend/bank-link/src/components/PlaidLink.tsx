import { useState, useEffect, useCallback } from 'react';
import {
  usePlaidLink,
  type PlaidLinkOptions,
  type PlaidLinkOnSuccess,
} from 'react-plaid-link';

interface PlaidLinkProps {
  onSuccess: (publicToken: string, metadata: any) => void;
  onExit?: (error: any, metadata: any) => void;
  jwt: string;
}

export default function PlaidLink({ onSuccess, onExit, jwt }: PlaidLinkProps) {
  const [linkToken, setLinkToken] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Fetch link token from backend
  useEffect(() => {
    const fetchLinkToken = async () => {
      try {
        setLoading(true);
        const backendUrl = import.meta.env.VITE_BACKEND_URL || 'http://0.0.0.0:8080';
        const response = await fetch(`${backendUrl}/create-link-token`, {
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${jwt}`,
            'Content-Type': 'application/json',
          },
        });

        if (!response.ok) {
          throw new Error(`Failed to fetch link token: ${response.statusText}`);
        }

        const data = await response.json();
        setLinkToken(data.link_token);
      } catch (err) {
        console.error('Error fetching link token:', err);
        setError(err instanceof Error ? err.message : 'Failed to fetch link token');
      } finally {
        setLoading(false);
      }
    };

    if (jwt) {
      fetchLinkToken();
    }
  }, [jwt]);

  const handleOnSuccess: PlaidLinkOnSuccess = useCallback(
    (publicToken, metadata) => {
      console.log('Plaid Link success:', { publicToken, metadata });
      onSuccess(publicToken, metadata);
    },
    [onSuccess]
  );

  const handleOnExit = useCallback(
    (error: any, metadata: any) => {
      console.log('Plaid Link exit:', { error, metadata });
      if (onExit) {
        onExit(error, metadata);
      }
    },
    [onExit]
  );

  const config: PlaidLinkOptions = {
    token: linkToken!,
    onSuccess: handleOnSuccess,
    onExit: handleOnExit,
  };

  const { open, ready } = usePlaidLink(config);

  // Auto-open when ready
  useEffect(() => {
    if (ready && linkToken) {
      open();
    }
  }, [ready, linkToken, open]);

  if (loading) {
    return (
      <div className="flex items-center justify-center p-6">
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Preparing Plaid Link...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="flex items-center justify-center p-6">
        <div className="text-center">
          <div className="text-red-600 text-4xl mb-4">⚠️</div>
          <h3 className="text-lg font-semibold text-gray-800 mb-2">Error</h3>
          <p className="text-gray-600 mb-4">{error}</p>
          <button 
            onClick={() => window.location.reload()} 
            className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
          >
            Try Again
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex items-center justify-center p-6">
      <div className="text-center">
        <div className="text-blue-600 text-4xl mb-4">🏦</div>
        <p className="text-gray-600 mb-4">Opening Plaid Link...</p>
        <button 
          onClick={() => open()} 
          disabled={!ready}
          className="px-6 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {ready ? 'Connect Bank Account' : 'Loading...'}
        </button>
      </div>
    </div>
  );
}