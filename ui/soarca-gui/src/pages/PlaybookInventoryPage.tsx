import React from 'react';
import PlaybookInventory from '../components/PlaybookInventory'; // Adjust the import path as needed

const PlaybookInventoryPage: React.FC = () => {
  return (
    <div className="px-4 sm:px-6 lg:px-8 py-8 w-full max-w-9xl mx-auto">
      {/* Page header */}
      <div className="sm:flex sm:justify-between sm:items-center mb-8">
        {/* Left side */}
        <div>
          <h1 className="text-2xl md:text-3xl text-gray-800 dark:text-gray-100 font-bold">Playbook Inventory</h1>
        </div>

        {/* Right side (placeholder for buttons) */}
        <div className="grid grid-flow-col sm:auto-cols-max justify-start sm:justify-end gap-2">
          {/* You can add Filter, Add Playbook, etc. buttons here if needed */}
        </div>
      </div>

      {/* Content grid */}
      <div className="grid grid-cols-12 gap-6">
        {/* Playbook Inventory Component */}
        <div className="col-span-12 bg-white dark:bg-slate-700 shadow-md rounded-sm border border-gray-200 dark:border-slate-600 p-4">
          <PlaybookInventory />
        </div>
      </div>
    </div>
  );
};

export default PlaybookInventoryPage;
