import React from 'react'
import styled from 'styled-components'

export interface NavLink {
  label: string
  href: string
}

export interface UserMenuProps {
  name?: string
  onLogout?: () => void
}

export interface HeaderProps {
  brand?: React.ReactNode
  links?: NavLink[]
  onToggleMenu?: () => void
  user?: UserMenuProps
}

const Bar = styled.header`
  position: sticky; top: 0; z-index: 50;
  display: flex; align-items: center; justify-content: space-between;
  padding: var(--omx-spacing-3) var(--omx-spacing-4);
  background: var(--omx-color-bg-default, #fff);
  border-bottom: 1px solid rgba(0,0,0,0.06);
`

const Brand = styled.div`
  display: flex; align-items: center; gap: var(--omx-spacing-2);
  font-weight: 600; color: var(--omx-color-text-default, #0f172a);
`

const Nav = styled.nav`
  display: none; gap: var(--omx-spacing-4);
  @media (min-width: 768px) { display: flex; }
`

const LinkA = styled.a`
  color: #2563eb; text-decoration: none;
  &:hover { text-decoration: underline; }
`

const Right = styled.div`
  display: flex; align-items: center; gap: var(--omx-spacing-3);
`

const IconBtn = styled.button`
  background: transparent; border: none; cursor: pointer; font-size: 1.25rem;
`

export const Header: React.FC<HeaderProps> = ({ brand = 'OllamaMax', links = [], onToggleMenu, user }) => {
  return (
    <Bar className="omx-v2" role="banner">
      <Brand>
        <IconBtn aria-label="Open menu" onClick={onToggleMenu} style={{ display: 'inline-flex' }}>
          <span aria-hidden>☰</span>
        </IconBtn>
        <span>{brand}</span>
      </Brand>
      <Nav aria-label="Primary">
        {links.map((l) => (
          <LinkA key={l.href} href={l.href}>{l.label}</LinkA>
        ))}
      </Nav>
      <Right>
        <IconBtn aria-label="Notifications"><span aria-hidden>🔔</span></IconBtn>
        {user?.name && (
          <div aria-label="User menu">{user.name} {user.onLogout && (
            <button onClick={user.onLogout} aria-label="Logout" style={{ marginLeft: 8 }}>Logout</button>
          )}</div>
        )}
      </Right>
    </Bar>
  )
}

