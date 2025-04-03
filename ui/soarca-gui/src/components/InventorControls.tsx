import React, { useState } from 'react';

interface InventoryControlsProps {
  searchTerm: string;
  onSearchChange: (term: string) => void;
  selectedCount: number;
  onDeleteSelected: () => void; // Add other bulk actions as needed
  onActivateSelected: () => void;
  onDeactivateSelected: () => void;
}

const InventoryControls: React.FC<InventoryControlsProps> = ({
  searchTerm,
  onSearchChange,
  selectedCount,
  onDeleteSelected,
  onActivateSelected,
  onDeactivateSelected
}) => {
  const [isActionDropdownOpen, setIsActionDropdownOpen] = useState(false);

  const handleActionClick = (action: () => void) => {
    action();
    setIsActionDropdownOpen(false); // Close dropdown after action
  };

  return (
    <div className="flex items-center justify-between flex-column md:flex-row flex-wrap space-y-4 md:space-y-0 py-4 bg-white dark:bg-gray-900">
      <div className="relative">
        <button
          id="dropdownActionButton"
          onClick={() => setIsActionDropdownOpen(!isActionDropdownOpen)}
          className="inline-flex items-center text-gray-500 bg-white border border-gray-300 focus:outline-none hover:bg-gray-100 focus:ring-4 focus:ring-gray-100 font-medium rounded-lg text-sm px-3 py-1.5 dark:bg-gray-800 dark:text-gray-400 dark:border-gray-600 dark:hover:bg-gray-700 dark:hover:border-gray-600 dark:focus:ring-gray-700"
          type="button"
        >
          <span className="sr-only">Action button</span>
          Action {selectedCount > 0 ? `(${selectedCount})` : ''}
          <svg className="w-2.5 h-2.5 ms-2.5" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 10 6">
            <path stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="m1 1 4 4 4-4" />
          </svg>
        </button>
        {/* Dropdown menu */}
        <div
          id="dropdownAction"
          className={`z-10 ${isActionDropdownOpen ? 'block' : 'hidden'} absolute mt-1 bg-white divide-y divide-gray-100 rounded-lg shadow w-44 dark:bg-gray-700 dark:divide-gray-600`}
        >
          <ul className="py-1 text-sm text-gray-700 dark:text-gray-200" aria-labelledby="dropdownActionButton">
            <li>
              <button
                onClick={() => handleActionClick(onActivateSelected)}
                disabled={selectedCount === 0}
                className={`block w-full text-left px-4 py-2 hover:bg-gray-100 dark:hover:bg-gray-600 dark:hover:text-white ${selectedCount === 0 ? 'opacity-50 cursor-not-allowed' : ''}`}
              >
                Activate Selected
              </button>
            </li>
            <li>
              <button
                onClick={() => handleActionClick(onDeactivateSelected)}
                disabled={selectedCount === 0}
                className={`block w-full text-left px-4 py-2 hover:bg-gray-100 dark:hover:bg-gray-600 dark:hover:text-white ${selectedCount === 0 ? 'opacity-50 cursor-not-allowed' : ''}`}
              >
                Deactivate Selected
              </button>
            </li>
          </ul>
          <div className="py-1">
            <button
              onClick={() => handleActionClick(onDeleteSelected)}
              disabled={selectedCount === 0}
              className={`block w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-gray-100 dark:hover:bg-gray-600 dark:text-red-500 dark:hover:text-white ${selectedCount === 0 ? 'opacity-50 cursor-not-allowed' : ''}`}
            >
              Delete Selected
            </button>
          </div>
        </div>
      </div>
      <label htmlFor="table-search" className="sr-only">Search</label>
      <div className="relative">
        <div className="absolute inset-y-0 rtl:inset-r-0 start-0 flex items-center ps-3 pointer-events-none">
          <svg className="w-4 h-4 text-gray-500 dark:text-gray-400" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 20 20">
            <path stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="m19 19-4-4m0-7A7 7 0 1 1 1 8a7 7 0 0 1 14 0Z" />
          </svg>
        </div>
        <input
          type="text"
          id="table-search-playbooks"
          value={searchTerm}
          onChange={(e) => onSearchChange(e.target.value)}
          className="block pt-2 ps-10 text-sm text-gray-900 border border-gray-300 rounded-lg w-80 bg-gray-50 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
          placeholder="Search for playbooks"
        />
      </div>
    </div>
  );
}

export default InventoryControls;
