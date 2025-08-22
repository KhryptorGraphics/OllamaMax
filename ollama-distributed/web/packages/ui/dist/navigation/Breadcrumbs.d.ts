import React from 'react';
export interface Crumb {
    label: string;
    href?: string;
}
export interface BreadcrumbsProps {
    items: Crumb[];
    'aria-label'?: string;
}
export declare const Breadcrumbs: React.FC<BreadcrumbsProps>;
