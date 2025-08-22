import React from 'react';
export type ButtonVariant = 'primary' | 'secondary' | 'danger' | 'ghost';
export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    variant?: ButtonVariant;
    'aria-label'?: string;
}
export declare const Button: React.FC<ButtonProps>;
