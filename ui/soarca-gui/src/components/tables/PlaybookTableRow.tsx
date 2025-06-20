import React from 'react';
import { Playbook } from '../types';

interface PlaybookTableRowProps {
  playbook: Playbook;
  isSelected: boolean;
  onSelect: (id: string) => void;
  onEdit: (playbook: Playbook) => void;
}

const PlaybookTableRow: React.FC<PlaybookTableRowProps> = ({ playbook, isSelected, onSelect, onEdit }) => {

  const getStatusBadge = (status: Playbook['status']) => {
    switch (status) {
      case 'Active':
        return <div className="h-2.5 w-2.5 rounded-full bg-green-500 me-2"></div>;
      case 'Inactive':
        return <div className="h-2.5 w-2.5 rounded-full bg-red-500 me-2"></div>;
      case 'Draft':
        return <div className="h-2.5 w-2.5 rounded-full bg-yellow-400 me-2"></div>;
      default:
        return <div className="h-2.5 w-2.5 rounded-full bg-gray-400 me-2"></div>;
    }
  };

  return (
    <tr className="bg-white border-b dark:bg-gray-800 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-600">
      <td className="w-4 p-4">
        <div className="flex items-center">
          <input
            id={`checkbox-table-search-${playbook.id}`}
            type="checkbox"
            checked={isSelected}
            onChange={() => onSelect(playbook.id)}
            className="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 dark:focus:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600"
          />
          <label htmlFor={`checkbox-table-search-${playbook.id}`} className="sr-only">checkbox</label>
        </div>
      </td>
      <th scope="row" className="px-6 py-4 font-medium text-gray-900 whitespace-nowrap dark:text-white">
        {playbook.name}
        <div className="font-normal text-gray-500 text-xs mt-1">ID: {playbook.id}</div>
        <div className="font-normal text-gray-500 text-xs mt-1">Created By: {playbook.createdBy}</div>
      </th>
      <td className="px-6 py-4">
        {playbook.description}
      </td>
      <td className="px-6 py-4">
        {/* Display Last Modified Date - format as needed */}
        {new Date(playbook.lastModified).toLocaleDateString()}
      </td>
      <td className="px-6 py-4">
        <div className="flex items-center">
          {getStatusBadge(playbook.status)} {playbook.status}
        </div>
      </td>
      <td className="px-6 py-4">
        <button
          type="button"
          onClick={() => onEdit(playbook)}
          className="font-medium text-blue-600 dark:text-blue-500 hover:underline"
        >
          Edit Playbook
        </button>
      </td>
    </tr>
  );
};

export default PlaybookTableRow;
