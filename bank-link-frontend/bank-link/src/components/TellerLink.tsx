import React from 'react';
import { useTellerConnect } from 'teller-connect-react';

interface TellerLinkProps {
  onSuccess: (authorization: any) => void;
}

export default function TellerLink({ onSuccess }: TellerLinkProps) {
  const { open, ready } = useTellerConnect({
    applicationId: import.meta.env.VITE_TELLER_APPLICATION_ID,
    environment: 'sandbox',
    onSuccess,
    // You can add onEvent, onExit, etc. here if needed
  });

  // Auto-open when ready
  React.useEffect(() => {
    if (ready) {
      open();
    }
  }, [ready, open]);

  return (
    <div className="flex items-center justify-center p-6">
      {ready && (
        <div className="text-center">
          <p className="text-gray-600">Opening Teller Connect...</p>
        </div>
      )}
    </div>
  );
} 