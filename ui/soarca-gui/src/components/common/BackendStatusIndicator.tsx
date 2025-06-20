import React, { useState, useEffect } from 'react';

type Status = 'checking' | 'ok' | 'error';

const BackendStatusIndicator: React.FC = () => {
  const [status, setStatus] = useState<Status>('checking');
  const soarcaUri = process.env.SOARCA_URI;

  useEffect(() => {
    const checkStatus = async () => {
      if (!soarcaUri) {
        console.error("SOARCA_URI environment variable is not defined.");
        setStatus('error');
        return;
      }

      try {
        const response = await fetch(`${soarcaUri}/status/ping`);
        if (response.ok) {
          const text = await response.text();
          setStatus(text.trim() === 'pong' ? 'ok' : 'error');
        } else {
          setStatus('error');
        }
      } catch (error) {
        console.error('Error pinging backend status:', error);
        setStatus('error');
      }
    };

    checkStatus();
    const intervalId = setInterval(checkStatus, 30000);

    return () => clearInterval(intervalId);
  }, [soarcaUri]);

  const renderIndicator = () => {
    switch (status) {
      case 'ok':
        return (
          <div
            className="w-3 h-3 rounded-full bg-green-500 animate-pulse"
            title="Backend Status: Connected"
          />
        );
      case 'error':
        return (
          <div
            className="w-3 h-3 rounded-full bg-red-500"
            title="Backend Status: Error"
          />
        );
      case 'checking':
      default:
        return (
          <div
            className="w-3 h-3 rounded-full bg-yellow-500"
            title="Backend Status: Checking..."
          />
        );
    }
  };

  return (
    <div className="flex items-center" aria-live="polite" aria-label={`Backend status is ${status}`}>
      {renderIndicator()}
    </div>
  );
};

export default BackendStatusIndicator;
