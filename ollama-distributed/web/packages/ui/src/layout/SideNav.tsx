import React, { useEffect, useState } from 'react'
import styled from 'styled-components'

export interface SideNavItem {
  label: string
  href: string
  icon?: React.ReactNode
  roles?: string[] // visible if user has one of roles (resolved by host app)
}

export interface SideNavProps {
  items: SideNavItem[]
  activeHref?: string
  collapsedKey?: string // localStorage key for collapsed state
  defaultCollapsed?: boolean
  onNavigate?: (href: string) => void
}

const Wrapper = styled.aside<{collapsed:boolean}>`
  position: sticky; top: 0; height: 100dvh; z-index: 40;
  width: ${(p)=> p.collapsed ? '64px' : '240px'};
  background: var(--omx-color-bg-subtle, #0ea5e91a);
  border-right: 1px solid rgba(0,0,0,0.06);
  transition: width .2s ease;
  display: flex; flex-direction: column; gap: 4px; padding: 8px;
`

const Item = styled.a<{active:boolean;collapsed:boolean}>`
  display: flex; align-items: center; gap: 8px; padding: 8px; border-radius: 8px;
  color: ${(p)=> p.active? '#0f172a':'#334155'}; text-decoration: none;
  background: ${(p)=> p.active? 'rgba(14,165,233,0.12)' : 'transparent'};
  &:hover { background: rgba(14,165,233,0.10); }
  > span.label { display: ${(p)=> p.collapsed? 'none':'inline'}; }
`

const Toggle = styled.button`
  background: transparent; border: none; cursor: pointer; padding: 8px; text-align: left;
`

export const SideNav: React.FC<SideNavProps> = ({ items, activeHref, collapsedKey='omx-sidenav', defaultCollapsed=false, onNavigate }) => {
  const [collapsed, setCollapsed] = useState(defaultCollapsed)
  useEffect(() => {
    const raw = localStorage.getItem(collapsedKey)
    if (raw != null) setCollapsed(raw === '1')
  }, [collapsedKey])

  const toggle = () => {
    setCollapsed((c)=>{
      const v = !c
      try { localStorage.setItem(collapsedKey, v ? '1' : '0') } catch{}
      return v
    })
  }

  return (
    <Wrapper className="omx-v2" role="navigation" aria-label="Sidebar" collapsed={collapsed}>
      <Toggle onClick={toggle} aria-label={collapsed? 'Expand sidebar':'Collapse sidebar'}>{collapsed? '»' : '«'}</Toggle>
      <div role="list">
        {items.map((it) => (
          <Item key={it.href} href={it.href} active={activeHref===it.href} collapsed={collapsed} onClick={(e)=>{ if(onNavigate){ e.preventDefault(); onNavigate(it.href) }}} role="listitem">
            {it.icon && <span aria-hidden>{it.icon}</span>}
            <span className="label">{it.label}</span>
          </Item>
        ))}
      </div>
    </Wrapper>
  )
}

