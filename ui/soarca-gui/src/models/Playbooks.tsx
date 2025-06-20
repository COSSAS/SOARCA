export interface Playbook {
  id: string;
  name: string;
  description: string;
  status: 'Active' | 'Inactive' | 'Draft';
  lastModified: string; // Or Date
  createdBy: string;
}
