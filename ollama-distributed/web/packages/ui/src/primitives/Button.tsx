import React from 'react'
import styled from 'styled-components'

export type ButtonVariant = 'primary' | 'secondary' | 'danger' | 'ghost'

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant
  'aria-label'?: string
}

const Root = styled.button<{ $variant: ButtonVariant }>`
  --bg: var(--omx-color-brand-500);
  --bg-hover: var(--omx-color-brand-600);
  --fg: #fff;
  --border: transparent;

  ${(p) => p.$variant === 'secondary' && `--bg: var(--omx-color-bg-surface); --fg: var(--omx-color-text-default); --border: var(--omx-color-text-muted);`}
  ${(p) => p.$variant === 'danger' && `--bg: #ef4444; --bg-hover: #dc2626;`}
  ${(p) => p.$variant === 'ghost' && `--bg: transparent; --bg-hover: rgba(0,0,0,0.05); --fg: var(--omx-color-text-default);`}

  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: var(--omx-spacing-2) var(--omx-spacing-4);
  border-radius: var(--omx-radius-md);
  background: var(--bg);
  color: var(--fg);
  border: 1px solid var(--border);
  transition: background 150ms ease;

  &:hover { background: var(--bg-hover); }
  &:focus-visible { outline: 2px solid var(--omx-color-brand-500); outline-offset: 2px; }
` 

export const Button: React.FC<ButtonProps> = ({ variant = 'primary', children, ...rest }) => {
  return (
    <Root className="omx-v2" $variant={variant} {...rest}>
      {children}
    </Root>
  )
}

