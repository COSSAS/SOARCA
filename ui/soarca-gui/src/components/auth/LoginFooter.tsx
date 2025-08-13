import React from 'react';
import { ArrowLeft, HelpCircle, KeyRound } from 'lucide-react';

const LoginFooter: React.FC = () => {
  const handleBackToSite = () => {
    // Use your preferred navigation method, window.location is simple
    window.location.href = 'https://soarca.com';
  }

  return (
    <>
      <div className="py-5">
        <div className="grid grid-cols-2 gap-1">
          <div className="text-center sm:text-left whitespace-nowrap">
            <button className="transition duration-200 mx-5 px-5 py-4 cursor-pointer font-normal text-sm rounded-lg text-gray-500 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-slate-700 focus:outline-none focus:bg-gray-200 focus:ring-2 focus:ring-gray-400 focus:ring-opacity-50 ring-inset">
              <KeyRound className="w-4 h-4 inline-block align-text-top" />
              <span className="inline-block ml-1">Forgot Password</span>
            </button>
          </div>
          <div className="text-center sm:text-right whitespace-nowrap">
            <button className="transition duration-200 mx-5 px-5 py-4 cursor-pointer font-normal text-sm rounded-lg text-gray-500 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-slate-700 focus:outline-none focus:bg-gray-200 focus:ring-2 focus:ring-gray-400 focus:ring-opacity-50 ring-inset">
              <HelpCircle className="w-4 h-4 inline-block align-text-bottom" />
              <span className="inline-block ml-1">Help</span>
            </button>
          </div>
        </div>
      </div>
      <div className="py-5">
        <div className="grid grid-cols-1 gap-1">
          <div className="text-center sm:text-left whitespace-nowrap">
            <button onClick={handleBackToSite} className="transition duration-200 mx-5 px-5 py-4 cursor-pointer font-normal text-sm rounded-lg text-gray-500 hover:bg-gray-200 dark:text-gray-400 dark:hover:bg-slate-700 focus:outline-none focus:bg-gray-300 focus:ring-2 focus:ring-gray-400 focus:ring-opacity-50 ring-inset">
              <ArrowLeft className="w-4 h-4 inline-block align-text-top" />
              <span className="inline-block ml-1">Back to SOARCA.com</span>
            </button>
          </div>
        </div>
      </div>
    </>
  );
};

export default LoginFooter;
