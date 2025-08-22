import React from 'react'
import { render, screen, fireEvent } from '@testing-library/react'
import '@testing-library/jest-dom'
import { Header, SideNav, Breadcrumbs } from '@ollamamax/ui'

function wrapper(children: React.ReactNode) {
  return <div className="omx-v2">{children}</div>
}

describe('Shared UI components', () => {
  test('Header renders brand and links', () => {
    render(wrapper(<Header brand={<span>Brand</span>} links={[{label:'Home', href:'/'},{label:'Login', href:'/login'}]} />))
    expect(screen.getByRole('banner')).toBeInTheDocument()
    expect(screen.getByText('Brand')).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'Home' })).toHaveAttribute('href', '/')
  })

  test('SideNav toggles collapsed state', () => {
    render(wrapper(<SideNav items={[{label:'Home', href:'/'},{label:'Auth', href:'/login'}]} activeHref="/" />))
    const nav = screen.getByRole('navigation', { name: 'Sidebar' })
    expect(nav).toBeInTheDocument()
    const btn = screen.getByRole('button')
    fireEvent.click(btn)
  })

  test('Breadcrumbs render and last is current page', () => {
    render(wrapper(<Breadcrumbs items={[{label:'Home', href:'/'},{label:'Dashboard'}]} />))
    const nav = screen.getByRole('navigation', { name: 'Breadcrumb' })
    expect(nav).toBeInTheDocument()
    const items = screen.getAllByRole('listitem')
    expect(items[1]).toHaveAttribute('aria-current', 'page')
  })
})

