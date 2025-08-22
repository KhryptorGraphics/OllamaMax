/**
 * Layout Components - Export index
 * Comprehensive layout system for Sprint C
 */

// Main layout components
export { AppLayout } from './AppLayout'
export { PageHeader } from './PageHeader'
export { Sidebar } from './Sidebar'
export { 
  Navigation,
  NAVIGATION_CONFIG,
  USER_ROLES,
  hasPermission,
  filterMenuByPermissions,
  filterGroupsByPermissions,
  useNavigation,
  NotificationBadge,
  NavigationItem
} from './Navigation'

// Types
export type {
  MenuItem,
  MenuGroup,
  NavigationConfig,
  Permission,
  UserRole
} from './Navigation'