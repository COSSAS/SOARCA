import React from 'react';
import PlaybookInventory from '../components/PlaybookInventory';

const PlaybookInventoryPage: React.FC = () => {
  return (
    <div className="px-4 sm:px-6 lg:px-8 py-8 w-full max-w-9xl mx-auto">
      <div className="sm:flex sm:justify-between sm:items-center mb-8">
        <div>
          <h1 className="text-2xl md:text-3xl text-gray-800 dark:text-gray-100 font-bold">Playbook Inventory</h1>
        </div>

        <div className="grid grid-flow-col sm:auto-cols-max justify-start sm:justify-end gap-2">
          {/* Placeholder for potential action buttons */}
        </div>
      </div>

      {/* Removed grid grid-cols-12 gap-6 from this div */}
      <div>
        {/* Removed col-span-12 from this div as parent is no longer a grid */}
        <PlaybookInventory />
      </div>
    </div>
  );
};

export default PlaybookInventoryPage;
