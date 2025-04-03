import React, { useState, useMemo, useEffect } from 'react';
import { Playbook } from '../models/Playbooks.tsx';
import PlaybookTableRow from './tables/PlaybookTableRow.tsx';
import EditPlaybookModal from './modals/EditPlaybookModal.tsx';
import InventoryControls from './InventorControls.tsx';

// Mock Data - Replace with actual API call
const MOCK_PLAYBOOKS: Playbook[] = [
  { id: 'pb-001', name: 'Phishing Triage', description: 'Analyzes potential phishing emails.', status: 'Active', lastModified: '2025-03-15T10:00:00Z', createdBy: 'admin' },
  { id: 'pb-002', name: 'Malware Containment', description: 'Isolates endpoints infected with known malware.', status: 'Active', lastModified: '2025-04-01T14:30:00Z', createdBy: 'security.ops' },
  { id: 'pb-003', name: 'Suspicious Login Alert', description: 'Investigates and responds to unusual login activity.', status: 'Inactive', lastModified: '2024-11-20T09:15:00Z', createdBy: 'admin' },
  { id: 'pb-004', name: 'Vulnerability Scan Remediation', description: 'Creates tickets for high-severity vulnerabilities.', status: 'Draft', lastModified: '2025-04-02T11:05:00Z', createdBy: 'dev.team' },
  { id: 'pb-005', name: 'Cloud Resource Misconfiguration', description: 'Detects and notifies on insecure cloud settings.', status: 'Active', lastModified: '2025-03-28T16:45:00Z', createdBy: 'cloud.sec' },
];

const PlaybookInventory: React.FC = () => {
  const [playbooks, setPlaybooks] = useState<Playbook[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedPlaybookIds, setSelectedPlaybookIds] = useState<Set<string>>(new Set());
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingPlaybook, setEditingPlaybook] = useState<Playbook | null>(null);

  // Simulate fetching data
  useEffect(() => {
    // In a real app, fetch data here, e.g., using fetch or axios
    setPlaybooks(MOCK_PLAYBOOKS);
  }, []);

  const filteredPlaybooks = useMemo(() => {
    if (!searchTerm) return playbooks;
    return playbooks.filter(pb =>
      pb.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      pb.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
      pb.id.toLowerCase().includes(searchTerm.toLowerCase()) ||
      pb.createdBy.toLowerCase().includes(searchTerm.toLowerCase())
    );
  }, [playbooks, searchTerm]);

  const handleSelectRow = (id: string) => {
    setSelectedPlaybookIds(prev => {
      const newSelection = new Set(prev);
      if (newSelection.has(id)) {
        newSelection.delete(id);
      } else {
        newSelection.add(id);
      }
      return newSelection;
    });
  };

  const handleSelectAll = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.checked) {
      const allIds = new Set(filteredPlaybooks.map(pb => pb.id));
      setSelectedPlaybookIds(allIds);
    } else {
      setSelectedPlaybookIds(new Set());
    }
  };

  const handleEditClick = (playbook: Playbook) => {
    setEditingPlaybook(playbook);
    setIsModalOpen(true);
  };

  const handleCloseModal = () => {
    setIsModalOpen(false);
    setEditingPlaybook(null);
  };

  const handleSaveChanges = (updatedPlaybook: Playbook) => {
    // TODO: Implement actual save logic (e.g., API call)
    console.log('Saving:', updatedPlaybook);
    setPlaybooks(prev =>
      prev.map(pb => pb.id === updatedPlaybook.id ? { ...pb, ...updatedPlaybook, lastModified: new Date().toISOString() } : pb) // Update lastModified on save
    );
    handleCloseModal();
  };

  const handleDeleteSelected = () => {
    // TODO: Implement actual delete logic (e.g., API call)
    console.log('Deleting:', Array.from(selectedPlaybookIds));
    setPlaybooks(prev => prev.filter(pb => !selectedPlaybookIds.has(pb.id)));
    setSelectedPlaybookIds(new Set()); // Clear selection
  };

  const updateSelectedStatus = (status: Playbook['status']) => {
    // TODO: Implement actual status update logic (e.g., API call)
    console.log(`Updating status to ${status} for:`, Array.from(selectedPlaybookIds));
    setPlaybooks(prev => prev.map(pb =>
      selectedPlaybookIds.has(pb.id) ? { ...pb, status: status, lastModified: new Date().toISOString() } : pb
    ));
    setSelectedPlaybookIds(new Set()); // Clear selection
  };

  const handleActivateSelected = () => updateSelectedStatus('Active');
  const handleDeactivateSelected = () => updateSelectedStatus('Inactive');


  const isAllSelected = filteredPlaybooks.length > 0 && selectedPlaybookIds.size === filteredPlaybooks.length;


  return (
    <div className="relative overflow-x-auto shadow-md sm:rounded-lg">
      <InventoryControls
        searchTerm={searchTerm}
        onSearchChange={setSearchTerm}
        selectedCount={selectedPlaybookIds.size}
        onDeleteSelected={handleDeleteSelected}
        onActivateSelected={handleActivateSelected}
        onDeactivateSelected={handleDeactivateSelected}
      />
      <table className="w-full text-sm text-left rtl:text-right text-gray-500 dark:text-gray-400">
        <thead className="text-xs text-gray-700 uppercase bg-gray-50 dark:bg-gray-700 dark:text-gray-400">
          <tr>
            <th scope="col" className="p-4">
              <div className="flex items-center">
                <input
                  id="checkbox-all-search"
                  type="checkbox"
                  checked={isAllSelected}
                  onChange={handleSelectAll}
                  disabled={filteredPlaybooks.length === 0}
                  className="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 dark:focus:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600" />
                <label htmlFor="checkbox-all-search" className="sr-only">checkbox</label>
              </div>
            </th>
            <th scope="col" className="px-6 py-3">Name / Details</th>
            <th scope="col" className="px-6 py-3">Description</th>
            <th scope="col" className="px-6 py-3">Last Modified</th>
            <th scope="col" className="px-6 py-3">Status</th>
            <th scope="col" className="px-6 py-3">Action</th>
          </tr>
        </thead>
        <tbody>
          {filteredPlaybooks.map(playbook => (
            <PlaybookTableRow
              key={playbook.id}
              playbook={playbook}
              isSelected={selectedPlaybookIds.has(playbook.id)}
              onSelect={handleSelectRow}
              onEdit={handleEditClick}
            />
          ))}
          {filteredPlaybooks.length === 0 && (
            <tr>
              <td colSpan={6} className="px-6 py-4 text-center text-gray-500 dark:text-gray-400">
                No playbooks found.
              </td>
            </tr>
          )}
        </tbody>
      </table>

      <EditPlaybookModal
        isOpen={isModalOpen}
        onClose={handleCloseModal}
        playbook={editingPlaybook}
        onSave={handleSaveChanges}
      />
    </div>
  );
};

export default PlaybookInventory;
