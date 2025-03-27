import React, { useState, useEffect, useRef } from "react";
import { ChevronDown, X } from "lucide-react";
import SidebarLinkGroup from "./SidebarLinkGroup";
import { navItems } from "./NavItems.tsx";

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
  const trigger = useRef<HTMLButtonElement>(null);
  const sidebar = useRef<HTMLDivElement>(null);

  const storedSidebarExpanded = typeof window !== 'undefined' ? localStorage.getItem("sidebar-expanded") : null;
  const [sidebarExpanded, setSidebarExpanded] = useState(
    storedSidebarExpanded === null ? false : storedSidebarExpanded === "true"
  );

  const currentPath = typeof window !== 'undefined' ? window.location.pathname : '';

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
    if (typeof window !== 'undefined') {
      localStorage.setItem("sidebar-expanded", sidebarExpanded.toString());
      if (sidebarExpanded) {
        document.querySelector("body")?.classList.add("sidebar-expanded");
      } else {
        document.querySelector("body")?.classList.remove("sidebar-expanded");
      }
    }
  }, [sidebarExpanded]);

  return (
    <div className="min-w-fit">
      <div
        className={`fixed inset-0 bg-gray-900/30 z-40 lg:hidden lg:z-auto transition-opacity duration-200 ${sidebarOpen ? "opacity-100" : "opacity-0 pointer-events-none"
          }`}
        aria-hidden="true"
      />

      <div
        id="sidebar"
        ref={sidebar}
        className={`flex flex-col absolute z-40 left-0 top-0 lg:static lg:left-auto lg:top-auto lg:translate-x-0 h-[100dvh] overflow-y-scroll lg:overflow-y-auto no-scrollbar w-64 lg:w-20 lg:sidebar-expanded:!w-64 2xl:!w-64 shrink-0 bg-white dark:bg-gray-800 p-4 transition-all duration-200 ease-in-out ${sidebarOpen ? "translate-x-0" : "-translate-x-64"
          } ${variant === "v2"
            ? "border-r border-gray-200 dark:border-gray-700/60"
            : ""
          }`}
      >
        <div className="flex justify-between mb-10 pr-3 sm:px-2">
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
          <a href="/" className="block">
            <div className="w-8 h-8 bg-violet-500 rounded" />
          </a>
        </div>

        <div className="space-y-8">
          <div>
            <h3 className="text-xs uppercase text-gray-400 dark:text-gray-500 font-semibold pl-3 mb-3">
              <span
                className="hidden lg:block lg:sidebar-expanded:hidden 2xl:hidden text-center w-6"
                aria-hidden="true"
              >
                •••
              </span>
              <span className="lg:hidden lg:sidebar-expanded:block 2xl:block">
                Menu
              </span>
            </h3>
            <ul className="mt-3">
              {navItems.map((item) => {
                const isActive = currentPath === item.href || currentPath.startsWith(item.href + '/');
                const isGroupActive = item.children ? currentPath.startsWith(item.href) : false;

                return (
                  <React.Fragment key={item.label}>
                    {item.children ? (
                      <SidebarLinkGroup activeCondition={isGroupActive}>
                        {(handleClick, open) => (
                          <>
                            <a
                              href="#0"
                              className={`block text-gray-800 dark:text-gray-100 truncate transition duration-150 ${isGroupActive
                                ? "font-medium" // Group parent isn't styled like active link, just bolded maybe
                                : "hover:text-gray-900 dark:hover:text-white"
                                }`}
                              onClick={(e) => {
                                e.preventDefault();
                                handleClick();
                                // Expand sidebar automatically when a group is clicked if it's collapsed
                                if (!sidebarExpanded) {
                                  setSidebarExpanded(true);
                                }
                              }}
                            >
                              <div className="flex items-center justify-between px-2 py-1">
                                <div className="flex items-center">
                                  {item.icon}
                                  <span className="text-sm font-medium ml-3 lg:opacity-0 lg:sidebar-expanded:opacity-100 2xl:opacity-100 duration-200">
                                    {item.label}
                                  </span>
                                </div>
                                <div className="flex shrink-0 ml-2">
                                  <ChevronDown
                                    className={`w-3 h-3 shrink-0 ml-1 fill-current text-gray-400 dark:text-gray-500 ${open && "rotate-180"} transition-transform duration-200 lg:opacity-0 lg:sidebar-expanded:opacity-100 2xl:opacity-100`}
                                  />
                                </div>
                              </div>
                            </a>
                            <div className="lg:hidden lg:sidebar-expanded:block 2xl:block">
                              <ul className={`pl-8 pr-2 mt-1 ${!open && "hidden"}`}>
                                {item.children?.map(child => {
                                  const isChildActive = currentPath === child.href;
                                  return (
                                    <li key={child.label} className="mb-1 last:mb-0">
                                      <a
                                        href={child.href}
                                        className={`block transition duration-150 truncate rounded px-2 py-1 ${isChildActive
                                          ? "text-violet-600 dark:text-violet-400 bg-violet-100 dark:bg-violet-900/30 font-medium"
                                          : "text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200 hover:bg-gray-100 dark:hover:bg-gray-700/50"
                                          }`}
                                      >
                                        <span className="text-sm lg:opacity-0 lg:sidebar-expanded:opacity-100 2xl:opacity-100 duration-200">
                                          {child.label}
                                        </span>
                                      </a>
                                    </li>
                                  );
                                })}
                              </ul>
                            </div>
                          </>
                        )}
                      </SidebarLinkGroup>
                    ) : (
                      <li className={`px-1 py-1 rounded-lg mb-0.5 last:mb-0 ${isActive ? 'bg-gray-100 dark:bg-gray-700/50' : ''}`}>
                        <a
                          href={item.href}
                          className={`block text-gray-800 dark:text-gray-100 truncate transition duration-150 rounded px-2 py-1 ${isActive
                            ? "text-violet-600 dark:text-violet-400 font-medium" // Active link styling
                            : "hover:text-gray-900 dark:hover:text-white hover:bg-gray-100 dark:hover:bg-gray-700/50" // Hover for non-active
                            }`}
                        >
                          <div className="flex items-center">
                            {item.icon}
                            <span className="text-sm font-medium ml-3 lg:opacity-0 lg:sidebar-expanded:opacity-100 2xl:opacity-100 duration-200">
                              {item.label}
                            </span>
                          </div>
                        </a>
                      </li>
                    )}
                  </React.Fragment>
                );
              })}
            </ul>
          </div>
        </div>

        <div className="pt-3 hidden lg:inline-flex justify-end mt-auto">
          <div className="w-12 pl-4 pr-3 py-2">
            <button
              className="text-gray-400 hover:text-gray-500 dark:text-gray-500 dark:hover:text-gray-400"
              onClick={() => setSidebarExpanded(!sidebarExpanded)}
            >
              <span className="sr-only">Expand / collapse sidebar</span>
              <svg
                className={`w-4 h-4 fill-current transition-transform duration-200 ${sidebarExpanded && 'rotate-180'}`}
                viewBox="0 0 16 16"
              >
                <path d="M11.414 7.414l-3-3A.999.999 0 10 7 5.828L9.172 8 7 10.172a.999.999 0 101.414 1.414l3-3a.999.999 0 000-1.414zM4 8a1 1 0 011-1h6a1 1 0 110 2H5a1 1 0 01-1-1z" />
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Sidebar;
