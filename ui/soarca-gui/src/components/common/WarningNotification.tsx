import React from 'react';

interface WarningNotificationProps {
  message: string;
}

const WarningNotification: React.FC<WarningNotificationProps> = ({ message }) => (
  <div className="bg-orange-100 border-l-4 border-orange-500 text-orange-700 p-4 mb-4" role="alert">
    <p>{message}</p>
  </div>
);

export default WarningNotification;
