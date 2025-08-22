import React from 'react';
export interface NavLink {
    label: string;
    href: string;
}
export interface UserMenuProps {
    name?: string;
    onLogout?: () => void;
}
export interface HeaderProps {
    brand?: React.ReactNode;
    links?: NavLink[];
    onToggleMenu?: () => void;
    user?: UserMenuProps;
}
export declare const Header: React.FC<HeaderProps>;
