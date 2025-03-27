import React, { useState, ReactNode } from 'react';

interface SidebarLinkGroupProps {
  children: (handleClick: () => void, open: boolean) => ReactNode;
  activeCondition: boolean;
}

const SidebarLinkGroup: React.FC<SidebarLinkGroupProps> = ({
  children,
  activeCondition,
}) => {
  const [open, setOpen] = useState<boolean>(activeCondition);

  const handleClick = (): void => {
    setOpen(!open);
  };

  return (
    <li
      className={`px-1 py-1 rounded-lg mb-0.5 last:mb-0 ${activeCondition ? 'bg-gray-100 dark:bg-gray-700/50' : ''
        }`}
    >
      {children(handleClick, open)}
    </li>
  );
};

export default SidebarLinkGroup;
