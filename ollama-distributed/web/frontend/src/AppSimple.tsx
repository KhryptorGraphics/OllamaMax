import React from 'react'
import { Header, SideNav, Breadcrumbs } from '@ollamamax/ui'
import { Outlet, useLocation, useNavigate, Routes, Route, Navigate, BrowserRouter } from 'react-router-dom'
import { Login, Register, ForgotPassword, ResetPassword, VerifyEmail } from './routes/auth'

function useFlag(name: string): boolean {
  try { return localStorage.getItem(name) === '1' || (window as any)[name] === true } catch { return false }
}

function Shell() {
  const location = useLocation()
  const navigate = useNavigate()
  const useSideNav = useFlag('USE_SHARED_SIDENAV')
  const useBreadcrumbs = useFlag('USE_SHARED_BREADCRUMBS')

  const links = [
    { label: 'Dashboard', href: '/v2' },
    { label: 'Login', href: '/v2/auth/login' },
    { label: 'Register', href: '/v2/auth/register' },
  ]

  const items = [
    { label: 'Dashboard', href: '/v2' },
    { label: 'Auth', href: '/v2/auth/login' },
    { label: 'Register', href: '/v2/auth/register' },
  ]

  const parts = location.pathname.replace(/^\/+|\/+$/g,'').split('/')
  const crumbs = [{ label: 'Home', href: '/v2' }, ...parts.slice(1).map((p, i) => ({
    label: p.replace(/-/g,' ').replace(/\b\w/g, (m) => m.toUpperCase()),
    href: '/'+parts.slice(0, i+2).join('/'),
  }))]

  return (
    <div className="omx-v2 min-h-screen flex">
      <div className="w-full">
        <Header brand={<span>OllamaMax</span>} links={links} />
        {useBreadcrumbs && <Breadcrumbs items={crumbs} />}
        <div className="flex">
          {useSideNav && (
            <SideNav items={items} activeHref={location.pathname} onNavigate={(href)=>navigate(href)} />
          )}
          <main id="main" className="flex-1 p-8" tabIndex={-1}>
            <Outlet />
          </main>
        </div>
      </div>
    </div>
  )
}

export default function AppSimple() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/v2" element={<Shell /> }>
          <Route index element={<div className="omx-v2 p-6">Dashboard</div>} />
          <Route path="auth/login" element={<Login />} />
          <Route path="auth/register" element={<Register />} />
          <Route path="auth/forgot-password" element={<ForgotPassword />} />
          <Route path="auth/reset-password" element={<ResetPassword />} />
          <Route path="auth/verify-email" element={<VerifyEmail />} />
        </Route>
        <Route path="*" element={<Navigate to="/v2" replace />} />
      </Routes>
    </BrowserRouter>
  )
}

