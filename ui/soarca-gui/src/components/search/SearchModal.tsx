import React from 'react';

interface SearchModalProps {
  isOpen: boolean;
  onClose: () => void;
}

const SearchModal: React.FC<SearchModalProps> = ({ isOpen, onClose }) => {
  if (!isOpen) {
    return null;
  }

  const handleContentClick = (e: React.MouseEvent) => {
    e.stopPropagation(); // Prevent modal closing when clicking inside content
  };

  return (
    <div
      id="search-modal"
      className="fixed inset-0 z-50 flex items-start justify-center p-4 pt-[20vh] bg-black/50 backdrop-blur-sm"
      onClick={onClose} // Close when clicking backdrop
      role="dialog"
      aria-modal="true"
      aria-labelledby="search-modal-title"
    >
      <div
        className="bg-white dark:bg-gray-800 rounded-lg shadow-xl p-6 w-full max-w-md"
        onClick={handleContentClick}
      >
        <h2 id="search-modal-title" className="text-lg font-semibold mb-4 text-gray-900 dark:text-white">Search</h2>
        <input
          type="text"
          placeholder="Search..."
          className="w-full p-2 border rounded bg-white dark:bg-gray-700 border-gray-300 dark:border-gray-600 text-gray-900 dark:text-white placeholder-gray-500 dark:placeholder-gray-400 focus:outline-none focus:ring-2 focus:ring-violet-500"
          aria-label="Search input"
        />
        <button
          onClick={onClose}
          className="mt-4 text-sm text-blue-600 dark:text-blue-400 hover:underline focus:outline-none"
        >
          Close
        </button>
      </div>
    </div>
  );
};

export default SearchModal;
