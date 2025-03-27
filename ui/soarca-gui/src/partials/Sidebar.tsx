import React, { useState, useEffect, useRef } from "react";
import { NavLink, useLocation } from "react-router-dom";
import { ChevronDown, X } from "lucide-react";
import SidebarLinkGroup from "./SidebarLinkGroup.tsx"; // Assuming this component exists

interface SidebarProps {
  sidebarOpen: boolean;
  setSidebarOpen: (open: boolean) => void;
  variant?: "default" | "v2";
}

const Sidebar: React.FC<SidebarProps> = ({
  sidebarOpen,
  setSidebarOpen,
  variant = "default",
}) => {
  const location = useLocation();
  const { pathname } = location;

  const trigger = useRef<HTMLButtonElement>(null);
  const sidebar = useRef<HTMLDivElement>(null);

  const storedSidebarExpanded = localStorage.getItem("sidebar-expanded");
  const [sidebarExpanded, setSidebarExpanded] = useState(
    storedSidebarExpanded === null ? false : storedSidebarExpanded === "true"
  );

  useEffect(() => {
    const clickHandler = ({ target }: MouseEvent) => {
      if (!sidebar.current || !trigger.current) return;
      if (
        !sidebarOpen ||
        sidebar.current.contains(target as Node) ||
        trigger.current.contains(target as Node)
      )
        return;
      setSidebarOpen(false);
    };
    document.addEventListener("click", clickHandler);
    return () => document.removeEventListener("click", clickHandler);
  }, [sidebarOpen, setSidebarOpen]);

  useEffect(() => {
    const keyHandler = ({ key }: KeyboardEvent) => {
      if (!sidebarOpen || key !== 'Escape') return;
      setSidebarOpen(false);
    };
    document.addEventListener("keydown", keyHandler);
    return () => document.removeEventListener("keydown", keyHandler);
  }, [sidebarOpen, setSidebarOpen]);

  useEffect(() => {
    localStorage.setItem("sidebar-expanded", sidebarExpanded.toString());
    if (sidebarExpanded) {
      document.querySelector("body")?.classList.add("sidebar-expanded");
    } else {
      document.querySelector("body")?.classList.remove("sidebar-expanded");
    }
  }, [sidebarExpanded]);

  return (
    <div className="min-w-fit">
      {/* Sidebar backdrop (mobile only) */}
      <div
        className={`fixed inset-0 bg-gray-900/30 z-40 lg:hidden lg:z-auto transition-opacity duration-200 ${sidebarOpen ? "opacity-100" : "opacity-0 pointer-events-none"
          }`}
        aria-hidden="true"
      />

      {/* Sidebar */}
      <div
        id="sidebar"
        ref={sidebar}
        className={`flex flex-col absolute z-40 left-0 top-0 lg:static lg:left-auto lg:top-auto lg:translate-x-0 h-[100dvh] overflow-y-scroll lg:overflow-y-auto no-scrollbar w-64 lg:w-20 lg:sidebar-expanded:!w-64 2xl:!w-64 shrink-0 bg-white dark:bg-gray-800 p-4 transition-all duration-200 ease-in-out ${sidebarOpen ? "translate-x-0" : "-translate-x-64"
          } ${variant === "v2"
            ? "border-r border-gray-200 dark:border-gray-700/60"
            : "rounded-r-2xl shadow-xs"
          }`}
      >
        {/* Sidebar header */}
        <div className="flex justify-between mb-10 pr-3 sm:px-2">
          {/* Close button */}
          <button
            ref={trigger}
            className="lg:hidden text-gray-500 hover:text-gray-400"
            onClick={() => setSidebarOpen(!sidebarOpen)}
            aria-controls="sidebar"
            aria-expanded={sidebarOpen}
          >
            <span className="sr-only">Close sidebar</span>
            <X className="w-6 h-6" />
          </button>
          {/* Logo */}
          <NavLink end to="/" className="block">
            {/* Replace with your actual logo component or image */}
            <div className="w-8 h-8 bg-violet-500 rounded" />
          </NavLink>
        </div>

        {/* Links */}
        <div className="space-y-8">
          {/* Pages group */}
          <div>
            <h3 className="text-xs uppercase text-gray-400 dark:text-gray-500 font-semibold pl-3">
              <span
                className="hidden lg:block lg:sidebar-expanded:hidden 2xl:hidden text-center w-6"
                aria-hidden="true"
              >
                •••
              </span>
              <span className="lg:hidden lg:sidebar-expanded:block 2xl:block">
                Pages
              </span>
            </h3>
            <ul className="mt-3">
              {/* Example SidebarLinkGroup */}
              <SidebarLinkGroup activeCondition={pathname === "/"}>
                {(handleClick, open) => (
                  <>
                    <a
                      href="#0" // Using href="#0" for non-navigational links
                      className={`block text-gray-800 dark:text-gray-100 truncate transition duration-150 ${pathname === "/"
                        ? "" // Apply specific styles if needed for active group parent
                        : "hover:text-gray-900 dark:hover:text-white"
                        }`}
                      onClick={(e) => {
                        e.preventDefault(); // Prevent default anchor behavior
                        handleClick();
                        if (!sidebarExpanded) {
                          setSidebarExpanded(true);
                        }
                      }}
                    >
                      <div className="flex items-center justify-between">
                        <div className="flex items-center">
                          {/* Replace with your actual icon */}
                          <div className="w-4 h-4 bg-gray-400 rounded" />
                          <span className="text-sm font-medium ml-4 lg:opacity-0 lg:sidebar-expanded:opacity-100 2xl:opacity-100 duration-200">
                            Menu Item
                          </span>
                        </div>
                        {/* Arrow Icon */}
                        <div className="flex shrink-0 ml-2">
                          <ChevronDown
                            className={`w-3 h-3 shrink-0 ml-1 fill-current text-gray-400 dark:text-gray-500 ${open && "rotate-180"
                              } ${!sidebarExpanded && "!opacity-0"}`}
                          />
                        </div>
                      </div>
                    </a>
                    {/* Submenu */}
                    <div className="lg:hidden lg:sidebar-expanded:block 2xl:block">
                      <ul className={`pl-9 mt-1 ${!open && "hidden"}`}> {/* Adjusted padding */}
                        <li className="mb-1 last:mb-0">
                          <NavLink
                            end
                            to="/"
                            className={({ isActive }) =>
                              "block transition duration-150 truncate " +
                              (isActive
                                ? "text-violet-500"
                                : "text-gray-500/90 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200")
                            }
                          >
                            <span className="text-sm font-medium lg:opacity-0 lg:sidebar-expanded:opacity-100 2xl:opacity-100 duration-200">
                              Submenu Item
                            </span>
                          </NavLink>
                        </li>
                        {/* Add more submenu items here if needed */}
                      </ul>
                    </div>
                  </>
                )}
              </SidebarLinkGroup>
              {/* Add more SidebarLinkGroups or individual NavLinks here */}
            </ul>
          </div>
        </div>

        {/* Expand/collapse button */}
        <div className="pt-3 hidden lg:inline-flex 2xl:hidden justify-end mt-auto">
          <div className="w-12 pl-4 pr-3 py-2"> {/* Adjusted padding for centering maybe */}
            <button
              className="text-gray-400 hover:text-gray-500 dark:text-gray-500 dark:hover:text-gray-400"
              onClick={() => setSidebarExpanded(!sidebarExpanded)}
            >
              <span className="sr-only">Expand / collapse sidebar</span>
              {/* Replace with your actual icon for expand/collapse */}
              {/* Using simple arrows for example, replace with Lucide or other icons */}
              <svg className={`w-4 h-4 fill-current sidebar-expanded:rotate-180`} viewBox="0 0 16 16">
                <path d="M11.414 7.414l-3-3A.999.999 0 10 7 5.828L9.172 8 7 10.172a.999.999 0 101.414 1.414l3-3a.999.999 0 000-1.414zM4 8a1 1 0 011-1h6a1 1 0 110 2H5a1 1 0 01-1-1z" /> {/* Example icon logic */}
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Sidebar;
