import { Routes, Route, Navigate } from 'react-router-dom';

import LoginPage from './pages/LoginPage';
import DashboardLayout from './layouts/DashboardLayout';
import PlaybookInventoryPage from './pages/PlaybookInventoryPage.tsx';

function App() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />

      <Route path="/" element={<DashboardLayout />}>
        {/* Redirect base path "/" to "/playbooks" */}
        <Route index element={<Navigate to="/playbooks" replace />} />

        {/* Playbook Inventory Route */}
        <Route path="playbooks" element={<PlaybookInventoryPage />} />

        {/* Catch-all for unknown routes within the layout, redirects to playbooks */}
        <Route path="*" element={<Navigate to="/playbooks" replace />} />
      </Route>

    </Routes>
  );
}

export default App;
