import React, { useState } from 'react';
import {
  Menu,
  Search,
  Sun,
  Moon,
  Bell,
  User,
  ChevronDown,
  MoreHorizontal
} from 'lucide-react';

interface HeaderProps {
  userName: string;
  sidebarOpen: boolean;
  setSidebarOpen: (open: boolean) => void;
  onThemeChange?: () => void;
}

const Header: React.FC<HeaderProps> = ({
  userName,
  sidebarOpen,
  setSidebarOpen,
  onThemeChange,
}) => {
  const [modalOpen, setModalOpen] = useState<boolean>(false);
  const [notificationsOpen, setNotificationsOpen] = useState<boolean>(false);
  const [userMenuOpen, setUserMenuOpen] = useState<boolean>(false);
  const [mobileMenuOpen, setMobileMenuOpen] = useState<boolean>(false);

  const handleThemeToggle = onThemeChange ?? (() => console.log('Theme Toggle Clicked'));

  const handleNotificationsClick = () => {
    setNotificationsOpen(!notificationsOpen);
    setUserMenuOpen(false);
    setMobileMenuOpen(false);
  };

  const handleUserProfileClick = () => {
    setUserMenuOpen(!userMenuOpen);
    setNotificationsOpen(false);
    setMobileMenuOpen(false);
  };

  const handleMobileMenuClick = () => {
    setMobileMenuOpen(!mobileMenuOpen);
    setNotificationsOpen(false);
    setUserMenuOpen(false);
  };

  return (
    <header className="sticky top-0 z-30 bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700/60">
      <div className="flex flex-col items-center justify-between lg:flex-row lg:px-6">

        <div className="flex items-center justify-between w-full gap-2 px-3 py-3 border-b border-gray-200 dark:border-gray-700/60 sm:gap-4 lg:justify-normal lg:border-b-0 lg:px-0 lg:py-4 h-16">

          <button
            className={`lg:hidden transition-colors duration-150 ${sidebarOpen
              ? 'text-violet-600 dark:text-violet-400'
              : 'text-gray-500 hover:text-gray-600 dark:hover:text-gray-400'
              }`}
            aria-controls="sidebar"
            aria-expanded={sidebarOpen}
            onClick={(e) => {
              e.stopPropagation();
              setSidebarOpen(!sidebarOpen);
            }}
          >
            <span className="sr-only">Open sidebar</span>
            <Menu className="w-6 h-6" />
          </button>

          <div className="flex-1 flex justify-start lg:hidden"></div>

          <button
            className={`flex items-center justify-center w-10 h-10 rounded-lg lg:hidden transition-colors duration-150 ${mobileMenuOpen
              ? 'text-violet-600 dark:text-violet-400 bg-violet-100 dark:bg-violet-900/30'
              : 'text-gray-700 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'
              }`}
            aria-label="More options"
            onClick={handleMobileMenuClick}
          >
            <MoreHorizontal size={24} />
          </button>

          <div className="hidden lg:flex flex-grow"></div>

        </div>

        <div className="hidden items-center justify-between w-full gap-4 px-5 py-4 lg:flex lg:justify-end lg:px-0 lg:shadow-none h-16">
          <div className="flex items-center gap-3">

            <button
              className={`w-8 h-8 flex items-center justify-center rounded-full transition-colors duration-150 ${modalOpen
                ? 'bg-violet-100 dark:bg-violet-900/30'
                : 'hover:bg-gray-100 lg:hover:bg-gray-200 dark:hover:bg-gray-700/50 dark:lg:hover:bg-gray-700'
                }`}
              onClick={(e) => {
                e.stopPropagation();
                setModalOpen(true);
              }}
              aria-controls="search-modal"
              aria-haspopup="dialog"
              aria-expanded={modalOpen}
            >
              <span className="sr-only">Open search</span>
              <Search className={`w-4 h-4 transition-colors duration-150 ${modalOpen
                ? 'text-violet-600 dark:text-violet-400'
                : 'text-gray-500/80 dark:text-gray-400/80'
                }`}
              />
            </button>

            <button
              className="relative flex items-center justify-center text-gray-500 transition-colors bg-white border border-gray-200 rounded-full h-11 w-11 hover:bg-gray-100 hover:text-gray-700 dark:bg-gray-800 dark:border-gray-700/60 dark:text-gray-400 dark:hover:bg-gray-700 dark:hover:text-white"
              onClick={handleThemeToggle}
              aria-label="Toggle theme"
            >
              <Sun className="hidden dark:block" size={20} />
              <Moon className="dark:hidden" size={20} />
            </button>

            <div className="relative">
              <button
                className={`relative flex items-center justify-center rounded-full h-11 w-11 transition-colors duration-150 border ${notificationsOpen
                  ? 'text-violet-600 dark:text-violet-400 bg-violet-100 dark:bg-violet-900/30 border-violet-300 dark:border-violet-500/30'
                  : 'text-gray-500 dark:text-gray-400 bg-white dark:bg-gray-800 border-gray-200 dark:border-gray-700/60 hover:bg-gray-100 dark:hover:bg-gray-700 hover:text-gray-700 dark:hover:text-white'
                  }`}
                onClick={handleNotificationsClick}
                aria-label="View notifications"
                aria-expanded={notificationsOpen}
              >
                <span className="absolute right-0 top-0.5 z-10 flex h-2 w-2 rounded-full bg-orange-400">
                  <span className="absolute inline-flex w-full h-full bg-orange-400 rounded-full opacity-75 animate-ping"></span>
                </span>
                <Bell size={20} />
              </button>
            </div>

            <div className="w-px h-6 bg-gray-200 dark:bg-gray-700/60"></div>

          </div>

          <div className="relative">
            <button
              className={`flex items-center transition-colors duration-150 ${userMenuOpen ? '' : 'dark:text-gray-400'}`}
              onClick={handleUserProfileClick}
              aria-label="User menu"
              aria-haspopup="true"
              aria-expanded={userMenuOpen}
            >
              <span className={`mr-3 flex h-11 w-11 items-center justify-center overflow-hidden rounded-full transition-colors duration-150 ${userMenuOpen ? 'bg-violet-100 dark:bg-violet-900/30' : 'bg-gray-200 dark:bg-gray-700'} text-gray-600 dark:text-gray-400`}>
                <User size={24} className={`transition-colors duration-150 ${userMenuOpen ? 'text-violet-600 dark:text-violet-400' : 'text-gray-600 dark:text-gray-400'}`} />
              </span>
              <span className={`block mr-1 font-medium text-sm transition-colors duration-150 ${userMenuOpen ? 'text-violet-600 dark:text-violet-400' : 'text-gray-800 dark:text-gray-100'}`}>
                {userName}
              </span>
              <ChevronDown
                className={`transition-colors duration-150 ${userMenuOpen ? 'text-violet-500 dark:text-violet-300' : 'text-gray-500 dark:text-gray-400'}`}
                size={18}
                strokeWidth={1.5}
              />
            </button>
          </div>
        </div>
      </div>

      {modalOpen && (
        <div
          id="search-modal"
          className="fixed inset-0 z-50 flex items-start justify-center p-4 pt-[20vh] bg-black/50 backdrop-blur-sm"
          onClick={() => setModalOpen(false)}
          role="dialog"
          aria-modal="true"
          aria-labelledby="search-modal-title"
        >
          <div
            className="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 w-full max-w-md"
            onClick={e => e.stopPropagation()}
          >
            <h2 id="search-modal-title" className="text-lg font-semibold mb-4 text-gray-900 dark:text-white">Search</h2>
            <input type="text" placeholder="Search..." className="w-full p-2 border rounded bg-white dark:bg-gray-700 border-gray-300 dark:border-gray-600 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400" />
            <button onClick={() => setModalOpen(false)} className="mt-4 text-sm text-blue-600 dark:text-blue-400">Close</button>
          </div>
        </div>
      )}
    </header>
  );
};

export default Header;
