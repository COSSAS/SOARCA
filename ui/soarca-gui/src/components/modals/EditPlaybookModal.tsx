import React, { useState, useEffect } from 'react';
import { Playbook } from '../types';

interface EditPlaybookModalProps {
  isOpen: boolean;
  onClose: () => void;
  playbook: Playbook | null;
  onSave: (updatedPlaybook: Playbook) => void;
}

const EditPlaybookModal: React.FC<EditPlaybookModalProps> = ({ isOpen, onClose, playbook, onSave }) => {
  const [formData, setFormData] = useState<Partial<Playbook>>({});

  useEffect(() => {
    if (playbook) {
      setFormData({
        id: playbook.id,
        name: playbook.name,
        description: playbook.description,
        status: playbook.status,
      });
    } else {
      setFormData({}); // Reset form if no playbook is selected
    }
  }, [playbook]);

  if (!isOpen || !playbook) return null;

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({ ...prev, [name]: value }));
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    // Add validation if needed
    onSave(formData as Playbook); // Assume formData is valid Playbook for simplicity
  };

  return (
    <div
      id="editPlaybookModal"
      tabIndex={-1}
      aria-hidden={!isOpen}
      className={`fixed top-0 left-0 right-0 z-50 flex items-center justify-center w-full p-4 overflow-x-hidden overflow-y-auto md:inset-0 h-[calc(100%-1rem)] max-h-full bg-gray-900 bg-opacity-50 dark:bg-opacity-80 ${isOpen ? '' : 'hidden'}`}
    >
      <div className="relative w-full max-w-2xl max-h-full">
        <form onSubmit={handleSubmit} className="relative bg-white rounded-lg shadow dark:bg-gray-700">
          <div className="flex items-start justify-between p-4 border-b rounded-t dark:border-gray-600">
            <h3 className="text-xl font-semibold text-gray-900 dark:text-white">
              Edit Playbook
            </h3>
            <button
              type="button"
              className="text-gray-400 bg-transparent hover:bg-gray-200 hover:text-gray-900 rounded-lg text-sm w-8 h-8 ms-auto inline-flex justify-center items-center dark:hover:bg-gray-600 dark:hover:text-white"
              onClick={onClose}
            >
              <svg className="w-3 h-3" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 14 14">
                <path stroke="currentColor" strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="m1 1 6 6m0 0 6 6M7 7l6-6M7 7l-6 6" />
              </svg>
              <span className="sr-only">Close modal</span>
            </button>
          </div>
          <div className="p-6 space-y-6">
            <div className="grid grid-cols-6 gap-6">
              <div className="col-span-6">
                <label htmlFor="name" className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Playbook Name</label>
                <input
                  type="text"
                  name="name"
                  id="name"
                  value={formData.name || ''}
                  onChange={handleChange}
                  className="shadow-xs bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-600 focus:border-blue-600 block w-full p-2.5 dark:bg-gray-600 dark:border-gray-500 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                  placeholder="My Awesome Playbook"
                  required
                />
              </div>
              <div className="col-span-6">
                <label htmlFor="description" className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Description</label>
                <textarea
                  name="description"
                  id="description"
                  rows={3}
                  value={formData.description || ''}
                  onChange={handleChange}
                  className="shadow-xs bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-600 focus:border-blue-600 block w-full p-2.5 dark:bg-gray-600 dark:border-gray-500 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                  placeholder="What this playbook does..."
                />
              </div>
              <div className="col-span-6 sm:col-span-3">
                <label htmlFor="status" className="block mb-2 text-sm font-medium text-gray-900 dark:text-white">Status</label>
                <select
                  name="status"
                  id="status"
                  value={formData.status || ''}
                  onChange={handleChange}
                  className="shadow-xs bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-600 focus:border-blue-600 block w-full p-2.5 dark:bg-gray-600 dark:border-gray-500 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
                  required
                >
                  <option value="Draft">Draft</option>
                  <option value="Active">Active</option>
                  <option value="Inactive">Inactive</option>
                </select>
              </div>
            </div>
          </div>
          <div className="flex items-center p-6 space-x-3 rtl:space-x-reverse border-t border-gray-200 rounded-b dark:border-gray-600">
            <button
              type="submit"
              className="text-white bg-blue-700 hover:bg-blue-800 focus:ring-4 focus:outline-none focus:ring-blue-300 font-medium rounded-lg text-sm px-5 py-2.5 text-center dark:bg-blue-600 dark:hover:bg-blue-700 dark:focus:ring-blue-800"
            >
              Save Changes
            </button>
            <button
              type="button"
              onClick={onClose}
              className="text-gray-500 bg-white hover:bg-gray-100 focus:ring-4 focus:outline-none focus:ring-blue-300 rounded-lg border border-gray-200 text-sm font-medium px-5 py-2.5 hover:text-gray-900 focus:z-10 dark:bg-gray-700 dark:text-gray-300 dark:border-gray-500 dark:hover:text-white dark:hover:bg-gray-600 dark:focus:ring-gray-600"
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default EditPlaybookModal;
