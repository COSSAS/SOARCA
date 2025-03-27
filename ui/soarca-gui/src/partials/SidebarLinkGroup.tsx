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
      className={`pl-4 pr-3 py-2 rounded-lg mb-0.5 last:mb-0 bg-linear-to-r ${activeCondition &&
        'from-violet-500/[0.12] dark:from-violet-500/[0.24] to-violet-500/[0.04]'
        }`}
    >
      {children(handleClick, open)}
    </li>
  );
};

export default SidebarLinkGroup;
