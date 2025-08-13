import React, { useState, FormEvent } from 'react';
import { ArrowRight } from 'lucide-react';

interface LoginFormProps {
  onSubmitSuccess: () => void;
  onSubmitError: (errors: Error[]) => void;
}

const LoginForm: React.FC<LoginFormProps> = ({ onSubmitSuccess, onSubmitError }) => {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setIsLoading(true);
    onSubmitError([]);

    try {
      const response = await fetch('/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, password }),
      });

      setIsLoading(false);

      if (!response.ok) {
        let errorsToShow = [new Error('Login failed. Please check your credentials.')];
        try {
          const errorData = await response.json();
          if (Array.isArray(errorData.errors) && errorData.errors.length > 0) {
            errorsToShow = errorData.errors.map((msg: string) => new Error(msg));
          } else if (errorData.message) {
            errorsToShow = [new Error(errorData.message)];
          }
        } catch (parseError) {
          console.error("Failed to parse error response:", parseError);
        }
        onSubmitError(errorsToShow);

      } else {
        onSubmitSuccess();
      }
    } catch (error) {
      setIsLoading(false);
      onSubmitError([new Error('A network error occurred. Please try again.')]);
      console.error('Login error:', error);
    }
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="px-5 py-7"
    >
      <label htmlFor="email" className="font-semibold text-sm text-gray-600 dark:text-gray-200 pb-1 block">E-mail</label>
      <input
        id="email"
        name="email"
        type="email"
        required
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        disabled={isLoading}
        className="border rounded-lg px-3 py-2 mt-1 mb-5 text-sm w-full dark:bg-slate-700 dark:text-gray-100 dark:border-slate-600 focus:outline-none focus:ring-2 focus:ring-blue-500"
      />
      <label htmlFor="password" className="font-semibold text-sm text-gray-600 dark:text-gray-200 pb-1 block">Password</label>
      <input
        id="password"
        name="password"
        type="password"
        required
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        disabled={isLoading}
        className="border rounded-lg px-3 py-2 mt-1 mb-5 text-sm w-full dark:bg-slate-700 dark:text-gray-100 dark:border-slate-600 focus:outline-none focus:ring-2 focus:ring-blue-500"
      />
      <button
        type="submit"
        disabled={isLoading}
        className="transition duration-200 bg-blue-500 hover:bg-blue-600 focus:bg-blue-700 focus:shadow-sm focus:ring-4 focus:ring-blue-500 focus:ring-opacity-50 text-white w-full py-2.5 rounded-lg text-sm shadow-sm hover:shadow-md font-semibold text-center inline-flex items-center justify-center disabled:opacity-50 disabled:cursor-not-allowed"
      >
        <span className="inline-block mr-2">{isLoading ? 'Logging in...' : 'Login'}</span>
        {!isLoading && <ArrowRight className="w-4 h-4 inline-block" />}
        {isLoading && (
          <svg className="animate-spin h-4 w-4 inline-block text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
            <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
            <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
        )}
      </button>
    </form>
  );
};

export default LoginForm;
