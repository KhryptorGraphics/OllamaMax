import React from 'react';
export interface SideNavItem {
    label: string;
    href: string;
    icon?: React.ReactNode;
    roles?: string[];
}
export interface SideNavProps {
    items: SideNavItem[];
    activeHref?: string;
    collapsedKey?: string;
    defaultCollapsed?: boolean;
    onNavigate?: (href: string) => void;
}
export declare const SideNav: React.FC<SideNavProps>;
