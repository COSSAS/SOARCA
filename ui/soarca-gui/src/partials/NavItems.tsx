import React, { ReactElement } from "react";
import {
  Home,
  Users,
  UserCircle,
  FolderOpen,
  Calendar,
  BookOpen,
  ShieldCheck,
  Settings,
  Puzzle,
  AlertTriangle,
  Database
} from "lucide-react";

export interface NavItem {
  icon: ReactElement;
  label: string;
  href: string;
  children?: NavItem[];
}

const iconClasses = "shrink-0 size-5";

const cloneIcon = (icon: ReactElement) => React.cloneElement(icon, { className: iconClasses });

export const navItems: NavItem[] = [
  {
    icon: cloneIcon(<Home />),
    label: 'Dashboard',
    href: '/dashboard'
  },
  {
    icon: cloneIcon(<AlertTriangle />),
    label: 'Incidents',
    href: '/incidents'
  },
  {
    icon: cloneIcon(<ShieldCheck />),
    label: 'Playbooks',
    href: '/playbooks'
  },
  {
    icon: cloneIcon(<Puzzle />),
    label: 'Integrations',
    href: '/integrations'
  },
  {
    icon: cloneIcon(<Database />),
    label: 'Assets',
    href: '/assets'
  },
  {
    icon: cloneIcon(<Users />),
    label: 'Users',
    href: '/users',
    children: [
      {
        icon: cloneIcon(<Users />),
        label: 'All Users',
        href: '/users/all'
      },
      {
        icon: cloneIcon(<Users />),
        label: 'Active Users',
        href: '/users/active'
      }
    ]
  },
  {
    icon: cloneIcon(<FolderOpen />),
    label: 'Projects',
    href: '/projects'
  },
  {
    icon: cloneIcon(<Calendar />),
    label: 'Calendar',
    href: '/calendar'
  },
  {
    icon: cloneIcon(<BookOpen />),
    label: 'Documentation',
    href: '/docs'
  },
  {
    icon: cloneIcon(<Settings />),
    label: 'Settings',
    href: '/settings',
    children: [
      {
        icon: cloneIcon(<UserCircle />),
        label: 'Profile',
        href: '/settings/profile'
      },
      {
        icon: cloneIcon(<Settings />),
        label: 'General',
        href: '/settings/general'
      },
      {
        icon: cloneIcon(<Puzzle />),
        label: 'Integration Settings',
        href: '/settings/integrations'
      }
    ]
  },
];
