import React from 'react';
import WarningNotification from '../common/WarningNotification.tsx'

interface AuthErrorDisplayProps {
  errors: Error[];
}

const AuthErrorDisplay: React.FC<AuthErrorDisplayProps> = ({ errors }) => {
  if (!errors || errors.length === 0) {
    return null;
  }

  return (
    <>
      {errors.map((err, index) => (
        <WarningNotification key={index} message={err.message} />
      ))}
    </>
  );
};

export default AuthErrorDisplay;
