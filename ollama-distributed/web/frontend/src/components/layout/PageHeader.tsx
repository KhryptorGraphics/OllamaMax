/**
 * PageHeader - Dynamic page headers with breadcrumbs and actions
 * Features: Dynamic breadcrumbs, action buttons, search and filter integration
 */

import React from 'react'
import styled from 'styled-components'
import { useLocation, Link } from 'react-router-dom'
import { 
  ChevronRight, 
  Download, 
  Share2, 
  Filter, 
  RefreshCw,
  MoreHorizontal,
  Search,
  SortAsc,
  SortDesc,
  Grid,
  List
} from 'lucide-react'

import { Button } from '../../design-system/components/Button/Button'
import { Input } from '../../design-system/components/Input/Input'

// Styled Components
const HeaderContainer = styled.div`
  padding: 1rem 1.5rem;
  background-color: ${({ theme }) => theme.colors.background.primary};
  border-bottom: 1px solid ${({ theme }) => theme.colors.border.primary};
`

const HeaderTop = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1rem;

  @media (max-width: 768px) {
    flex-direction: column;
    align-items: flex-start;
    gap: 1rem;
  }
`

const TitleSection = styled.div`
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
`

const PageTitle = styled.h1`
  font-size: 1.5rem;
  font-weight: 600;
  color: ${({ theme }) => theme.colors.text.primary};
  margin: 0;
  line-height: 1.2;
`

const PageSubtitle = styled.p`
  font-size: 0.875rem;
  color: ${({ theme }) => theme.colors.text.secondary};
  margin: 0;
`

const Breadcrumbs = styled.nav`
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
`

const BreadcrumbList = styled.ol`
  display: flex;
  align-items: center;
  gap: 0.5rem;
  list-style: none;
  margin: 0;
  padding: 0;
`

const BreadcrumbItem = styled.li`
  display: flex;
  align-items: center;
  gap: 0.5rem;
`

const BreadcrumbLink = styled(Link)`
  color: ${({ theme }) => theme.colors.text.secondary};
  text-decoration: none;
  font-size: 0.875rem;
  border-radius: ${({ theme }) => theme.radius.sm};
  padding: 0.25rem 0.5rem;
  transition: ${({ theme }) => theme.transitions.colors};

  &:hover {
    color: ${({ theme }) => theme.colors.text.primary};
    background-color: ${({ theme }) => theme.colors.interactive.ghost.hover};
  }

  &:focus-visible {
    outline: 2px solid ${({ theme }) => theme.colors.border.focus};
    outline-offset: 2px;
  }
`

const BreadcrumbCurrent = styled.span`
  color: ${({ theme }) => theme.colors.text.primary};
  font-size: 0.875rem;
  font-weight: 500;
  padding: 0.25rem 0.5rem;
`

const BreadcrumbSeparator = styled(ChevronRight)`
  width: 0.75rem;
  height: 0.75rem;
  color: ${({ theme }) => theme.colors.text.tertiary};
`

const ActionsSection = styled.div`
  display: flex;
  align-items: center;
  gap: 0.75rem;

  @media (max-width: 768px) {
    width: 100%;
    justify-content: flex-end;
  }
`

const HeaderBottom = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;

  @media (max-width: 768px) {
    flex-direction: column;
    align-items: stretch;
    gap: 0.75rem;
  }
`

const FilterSection = styled.div`
  display: flex;
  align-items: center;
  gap: 0.75rem;
  flex: 1;

  @media (max-width: 768px) {
    flex-wrap: wrap;
  }
`

const SearchContainer = styled.div`
  position: relative;
  max-width: 300px;
  flex: 1;

  @media (max-width: 768px) {
    max-width: none;
    width: 100%;
  }
`

const ViewToggle = styled.div`
  display: flex;
  align-items: center;
  border: 1px solid ${({ theme }) => theme.colors.border.primary};
  border-radius: ${({ theme }) => theme.radius.md};
  overflow: hidden;
`

const ViewButton = styled.button<{ $active: boolean }>`
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.5rem;
  border: none;
  background-color: ${({ $active, theme }) =>
    $active ? theme.colors.interactive.primary.default : 'transparent'};
  color: ${({ $active, theme }) =>
    $active ? theme.colors.text.inverse : theme.colors.text.secondary};
  transition: ${({ theme }) => theme.transitions.colors};

  &:hover {
    background-color: ${({ $active, theme }) =>
      $active
        ? theme.colors.interactive.primary.hover
        : theme.colors.interactive.ghost.hover};
  }

  &:focus-visible {
    outline: 2px solid ${({ theme }) => theme.colors.border.focus};
    outline-offset: -2px;
  }

  svg {
    width: 1rem;
    height: 1rem;
  }
`

const SortButton = styled.button<{ $direction?: 'asc' | 'desc' | null }>`
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  border: 1px solid ${({ theme }) => theme.colors.border.primary};
  border-radius: ${({ theme }) => theme.radius.md};
  background-color: ${({ $direction, theme }) =>
    $direction ? theme.colors.interactive.primary.default : 'transparent'};
  color: ${({ $direction, theme }) =>
    $direction ? theme.colors.text.inverse : theme.colors.text.primary};
  font-size: 0.875rem;
  transition: ${({ theme }) => theme.transitions.colors};

  &:hover {
    background-color: ${({ $direction, theme }) =>
      $direction
        ? theme.colors.interactive.primary.hover
        : theme.colors.interactive.ghost.hover};
  }

  &:focus-visible {
    outline: 2px solid ${({ theme }) => theme.colors.border.focus};
    outline-offset: 2px;
  }

  svg {
    width: 0.875rem;
    height: 0.875rem;
  }
`

// Types
interface PageHeaderProps {
  title?: string
  subtitle?: string
  showBreadcrumbs?: boolean
  showSearch?: boolean
  showFilters?: boolean
  showViewToggle?: boolean
  showSort?: boolean
  actions?: React.ReactNode
  onSearch?: (query: string) => void
  onFilter?: () => void
  onExport?: () => void
  onRefresh?: () => void
  onSort?: (field: string, direction: 'asc' | 'desc') => void
  onViewChange?: (view: 'grid' | 'list') => void
  currentView?: 'grid' | 'list'
  sortField?: string
  sortDirection?: 'asc' | 'desc' | null
}

// Route-based page configurations
const PAGE_CONFIGS: Record<string, Partial<PageHeaderProps>> = {
  '/dashboard': {
    title: 'Dashboard',
    subtitle: 'System overview and key metrics'
  },
  '/models': {
    title: 'Models',
    subtitle: 'Manage and monitor AI models',
    showSearch: true,
    showFilters: true,
    showViewToggle: true,
    showSort: true
  },
  '/nodes': {
    title: 'Nodes',
    subtitle: 'Network nodes and cluster management',
    showSearch: true,
    showFilters: true,
    showViewToggle: true
  },
  '/monitoring': {
    title: 'Monitoring',
    subtitle: 'System performance and health metrics',
    showFilters: true
  },
  '/security': {
    title: 'Security',
    subtitle: 'Security settings and audit logs',
    showSearch: true,
    showFilters: true
  },
  '/settings': {
    title: 'Settings',
    subtitle: 'Application configuration and preferences'
  }
}

// Generate breadcrumbs from pathname
const generateBreadcrumbs = (pathname: string) => {
  const segments = pathname.split('/').filter(Boolean)
  const breadcrumbs = [{ label: 'Home', path: '/dashboard' }]

  let currentPath = ''
  segments.forEach((segment, index) => {
    currentPath += `/${segment}`
    const isLast = index === segments.length - 1

    // Capitalize and format segment
    const label = segment
      .replace(/-/g, ' ')
      .replace(/\b\w/g, (char) => char.toUpperCase())

    breadcrumbs.push({
      label,
      path: currentPath,
      isLast
    })
  })

  return breadcrumbs
}

// Component
export const PageHeader: React.FC<PageHeaderProps> = ({
  title,
  subtitle,
  showBreadcrumbs = true,
  showSearch = false,
  showFilters = false,
  showViewToggle = false,
  showSort = false,
  actions,
  onSearch,
  onFilter,
  onExport,
  onRefresh,
  onSort,
  onViewChange,
  currentView = 'grid',
  sortField = 'name',
  sortDirection = null,
  ...props
}) => {
  const location = useLocation()
  
  // Get page config based on current route
  const pageConfig = PAGE_CONFIGS[location.pathname] || {}
  const finalProps = { ...pageConfig, ...props }
  
  const finalTitle = title || finalProps.title || 'Page'
  const finalSubtitle = subtitle || finalProps.subtitle
  const finalShowSearch = showSearch || finalProps.showSearch || false
  const finalShowFilters = showFilters || finalProps.showFilters || false
  const finalShowViewToggle = showViewToggle || finalProps.showViewToggle || false
  const finalShowSort = showSort || finalProps.showSort || false

  // State
  const [searchQuery, setSearchQuery] = React.useState('')

  // Generate breadcrumbs
  const breadcrumbs = generateBreadcrumbs(location.pathname)

  // Handlers
  const handleSearch = (event: React.FormEvent) => {
    event.preventDefault()
    onSearch?.(searchQuery)
  }

  const handleSortToggle = () => {
    const newDirection = sortDirection === 'asc' ? 'desc' : 'asc'
    onSort?.(sortField, newDirection)
  }

  return (
    <HeaderContainer>
      <HeaderTop>
        <TitleSection>
          {showBreadcrumbs && breadcrumbs.length > 1 && (
            <Breadcrumbs aria-label="Breadcrumb navigation">
              <BreadcrumbList>
                {breadcrumbs.map((crumb, index) => (
                  <BreadcrumbItem key={crumb.path}>
                    {crumb.isLast ? (
                      <BreadcrumbCurrent aria-current="page">
                        {crumb.label}
                      </BreadcrumbCurrent>
                    ) : (
                      <>
                        <BreadcrumbLink to={crumb.path}>
                          {crumb.label}
                        </BreadcrumbLink>
                        {index < breadcrumbs.length - 1 && (
                          <BreadcrumbSeparator aria-hidden="true" />
                        )}
                      </>
                    )}
                  </BreadcrumbItem>
                ))}
              </BreadcrumbList>
            </Breadcrumbs>
          )}
          
          <PageTitle>{finalTitle}</PageTitle>
          {finalSubtitle && <PageSubtitle>{finalSubtitle}</PageSubtitle>}
        </TitleSection>

        <ActionsSection>
          {onRefresh && (
            <Button
              variant="ghost"
              size="sm"
              onClick={onRefresh}
              aria-label="Refresh data"
            >
              <RefreshCw size={16} />
            </Button>
          )}

          {onExport && (
            <Button
              variant="ghost"
              size="sm"
              onClick={onExport}
              aria-label="Export data"
            >
              <Download size={16} />
              Export
            </Button>
          )}

          <Button
            variant="ghost"
            size="sm"
            onClick={() => {}}
            aria-label="Share page"
          >
            <Share2 size={16} />
          </Button>

          <Button
            variant="ghost"
            size="sm"
            onClick={() => {}}
            aria-label="More options"
          >
            <MoreHorizontal size={16} />
          </Button>

          {actions}
        </ActionsSection>
      </HeaderTop>

      {(finalShowSearch || finalShowFilters || finalShowViewToggle || finalShowSort) && (
        <HeaderBottom>
          <FilterSection>
            {finalShowSearch && (
              <SearchContainer>
                <form onSubmit={handleSearch}>
                  <Input
                    type="search"
                    placeholder="Search..."
                    value={searchQuery}
                    onChange={(e) => setSearchQuery(e.target.value)}
                    aria-label="Search items"
                  />
                </form>
              </SearchContainer>
            )}

            {finalShowFilters && (
              <Button
                variant="ghost"
                size="sm"
                onClick={onFilter}
                aria-label="Open filters"
              >
                <Filter size={16} />
                Filters
              </Button>
            )}

            {finalShowSort && (
              <SortButton
                onClick={handleSortToggle}
                $direction={sortDirection}
                aria-label={`Sort by ${sortField} ${
                  sortDirection === 'asc' ? 'descending' : 'ascending'
                }`}
              >
                Sort
                {sortDirection === 'asc' ? (
                  <SortAsc />
                ) : sortDirection === 'desc' ? (
                  <SortDesc />
                ) : (
                  <SortAsc />
                )}
              </SortButton>
            )}
          </FilterSection>

          {finalShowViewToggle && (
            <ViewToggle>
              <ViewButton
                $active={currentView === 'grid'}
                onClick={() => onViewChange?.('grid')}
                aria-label="Grid view"
                aria-pressed={currentView === 'grid'}
              >
                <Grid />
              </ViewButton>
              <ViewButton
                $active={currentView === 'list'}
                onClick={() => onViewChange?.('list')}
                aria-label="List view"
                aria-pressed={currentView === 'list'}
              >
                <List />
              </ViewButton>
            </ViewToggle>
          )}
        </HeaderBottom>
      )}
    </HeaderContainer>
  )
}

export default PageHeader