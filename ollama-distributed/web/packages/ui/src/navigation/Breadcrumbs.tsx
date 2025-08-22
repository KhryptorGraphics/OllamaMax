import React from 'react'
import styled from 'styled-components'

export interface Crumb { label: string; href?: string }
export interface BreadcrumbsProps { items: Crumb[]; 'aria-label'?: string }

const Nav = styled.nav`
  font-size: 0.875rem; color: #475569; padding: 8px 16px;
`

const List = styled.ol`
  display: flex; align-items: center; gap: 8px; list-style: none; padding: 0; margin: 0;
`

const CrumbA = styled.a`
  color: #2563eb; text-decoration: none; &:hover{ text-decoration: underline; }
`

export const Breadcrumbs: React.FC<BreadcrumbsProps> = ({ items, 'aria-label': ariaLabel='Breadcrumb' }) => {
  return (
    <Nav aria-label={ariaLabel} className="omx-v2">
      <List>
        {items.map((c, i) => {
          const isLast = i === items.length - 1
          return (
            <li key={i} aria-current={isLast ? 'page' : undefined}>
              {c.href && !isLast ? <CrumbA href={c.href}>{c.label}</CrumbA> : <span>{c.label}</span>}
              {!isLast && <span aria-hidden style={{ margin: '0 4px' }}>/</span>}
            </li>
          )
        })}
      </List>
    </Nav>
  )
}

