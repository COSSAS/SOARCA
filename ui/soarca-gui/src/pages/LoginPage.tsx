import React, { useState } from 'react';
import AuthErrorDisplay from '../components/auth/AuthErrorDisplay';
import LoginForm from '../components/auth/LoginForm';
import LoginFooter from '../components/auth/LoginFooter';

const homeLink = "/";
const soarcaLogoUrlPath = "/assets/logos/soarca-logo.svg";

const LoginPage: React.FC = () => {
  const [errors, setErrors] = useState<Error[]>([]);

  const handleLoginSuccess = () => {
    console.log("Login successful!");
    setErrors([]);
    // TODO: Redirect user or update global auth state here
    // Example: navigate('/dashboard'); or authContext.login(userData);
  };

  const handleLoginError = (receivedErrors: Error[]) => {
    setErrors(receivedErrors);
  };

  return (
    <div className="min-h-screen bg-gray-100 dark:bg-slate-800 flex flex-col justify-center sm:py-12 font-sans">
      <div className="p-10 xs:p-0 mx-auto md:w-full md:max-w-md">
        <div className="flex justify-center mb-5">
          <a href={homeLink}>
            {/* Ensure the image path is correctly handled by your build process (e.g., placing it in the public folder) */}
            <img src={soarcaLogoUrlPath} alt="SOARCA Logo" className="w-30 h-auto" />
          </a>
        </div>
        <div className="bg-white dark:bg-slate-900 shadow w-full rounded-lg divide-y divide-gray-200 dark:divide-slate-700">
          <div className="px-5 pt-7">
            <AuthErrorDisplay errors={errors} />
          </div>
          {/* Assuming LoginForm, AuthErrorDisplay, LoginFooter are correctly implemented and imported */}
          <LoginForm onSubmitSuccess={handleLoginSuccess} onSubmitError={handleLoginError} />
          <LoginFooter />
        </div>
      </div>
    </div>
  );
};

export default LoginPage;
