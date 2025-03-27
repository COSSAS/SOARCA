import React, { useState } from 'react';

import Sidebar from '../partials/Sidebar.tsx';
import Header from '../partials/Header.tsx';
import Banner from '../partials/Banner.tsx';


const Dashboard: React.FC = () => {
  const [sidebarOpen, setSidebarOpen] = useState<boolean>(false);

  return (
    <div className="flex h-screen overflow-hidden">
      <Sidebar sidebarOpen={sidebarOpen} setSidebarOpen={setSidebarOpen} />
      <div className="relative flex flex-col flex-1 overflow-y-auto overflow-x-hidden">
        <Header sidebarOpen={sidebarOpen} setSidebarOpen={setSidebarOpen} />

        {/* Main content */}
        <main className="grow">
          <div className="px-4 sm:px-6 lg:px-8 py-8 w-full max-w-9xl mx-auto">
            {/* Dashboard actions/header */}
            <div className="sm:flex sm:justify-between sm:items-center mb-8">
              {/* Page title and action buttons */}
              <div>
                <h1 className="text-2xl md:text-3xl text-gray-800 dark:text-gray-100 font-bold">Dashboard</h1>
              </div>
              <div className="grid grid-flow-col sm:auto-cols-max justify-start sm:justify-end gap-2">
                {/* Filter, datepicker, and add view buttons */}
              </div>
            </div>

            {/* Dashboard cards grid */}
            <div className="grid grid-cols-12 gap-6">
              {/* Individual dashboard cards */}
              {/* <DashboardCard01 /> */}
              {/* More dashboard cards */}
            </div>
          </div>
        </main>

        {/* Banner component */}
        <Banner />
      </div>
    </div>
  );
};

export default Dashboard;
