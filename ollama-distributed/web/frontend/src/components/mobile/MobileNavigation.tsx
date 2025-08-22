import React, { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import { 
  Home, 
  Settings, 
  Monitor, 
  Shield, 
  Menu, 
  X, 
  ChevronRight,
  Bell,
  User,
  Search,
  Cpu,
  BarChart3,
  Server,
  Activity,
  CheckSquare,
  RefreshCw,
  Zap
} from 'lucide-react';

interface MobileNavigationProps {
  onNavigate?: (path: string) => void;
  className?: string;
}

interface NavItem {
  id: string;
  label: string;
  path: string;
  icon: React.ComponentType<any>;
  badge?: number;
  description?: string;
}

const navigationItems: NavItem[] = [
  {
    id: 'dashboard',
    label: 'Dashboard',
    path: '/dashboard',
    icon: Home,
    description: 'System overview and metrics'
  },
  {
    id: 'models',
    label: 'Models',
    path: '/models',
    icon: Cpu,
    description: 'AI model management'
  },
  {
    id: 'nodes',
    label: 'Nodes',
    path: '/nodes',
    icon: Server,
    description: 'Cluster node management'
  },
  {
    id: 'monitoring',
    label: 'Monitoring',
    path: '/monitoring',
    icon: Monitor,
    badge: 3,
    description: 'System monitoring and observability'
  },
  {
    id: 'tasks',
    label: 'Tasks',
    path: '/tasks',
    icon: CheckSquare,
    description: 'Task management and scheduling'
  },
  {
    id: 'transfers',
    label: 'Transfers',
    path: '/transfers',
    icon: RefreshCw,
    description: 'Data transfer management'
  },
  {
    id: 'security',
    label: 'Security',
    path: '/security',
    icon: Shield,
    description: 'Security settings and audit'
  },
  {
    id: 'performance',
    label: 'Performance',
    path: '/performance',
    icon: Zap,
    description: 'Performance optimization'
  },
  {
    id: 'settings',
    label: 'Settings',
    path: '/settings',
    icon: Settings,
    description: 'Application settings'
  }
];

export const MobileNavigation: React.FC<MobileNavigationProps> = ({
  onNavigate,
  className = ''
}) => {
  const navigate = useNavigate();
  const location = useLocation();
  const [isDrawerOpen, setIsDrawerOpen] = useState(false);
  const [searchQuery, setSearchQuery] = useState('');
  const [filteredItems, setFilteredItems] = useState(navigationItems);

  // Handle navigation
  const handleNavigate = (path: string) => {
    navigate(path);
    onNavigate?.(path);
    setIsDrawerOpen(false);
  };

  // Filter navigation items based on search
  useEffect(() => {
    if (searchQuery.trim()) {
      const filtered = navigationItems.filter(item =>
        item.label.toLowerCase().includes(searchQuery.toLowerCase()) ||
        item.description?.toLowerCase().includes(searchQuery.toLowerCase())
      );
      setFilteredItems(filtered);
    } else {
      setFilteredItems(navigationItems);
    }
  }, [searchQuery]);

  // Close drawer on location change
  useEffect(() => {
    setIsDrawerOpen(false);
  }, [location.pathname]);

  // Handle touch gestures for drawer
  useEffect(() => {
    let startX = 0;
    let currentX = 0;
    let isDragging = false;

    const handleTouchStart = (e: TouchEvent) => {
      if (e.touches.length !== 1) return;
      startX = e.touches[0].clientX;
      
      // Only start gesture if starting from the left edge
      if (startX < 20) {
        isDragging = true;
      }
    };

    const handleTouchMove = (e: TouchEvent) => {
      if (!isDragging || e.touches.length !== 1) return;
      
      currentX = e.touches[0].clientX;
      const deltaX = currentX - startX;
      
      // Open drawer if swiping right from left edge
      if (deltaX > 80 && !isDrawerOpen) {
        setIsDrawerOpen(true);
        isDragging = false;
      }
      
      // Close drawer if swiping left when open
      if (deltaX < -80 && isDrawerOpen) {
        setIsDrawerOpen(false);
        isDragging = false;
      }
    };

    const handleTouchEnd = () => {
      isDragging = false;
    };

    document.addEventListener('touchstart', handleTouchStart, { passive: true });
    document.addEventListener('touchmove', handleTouchMove, { passive: true });
    document.addEventListener('touchend', handleTouchEnd, { passive: true });

    return () => {
      document.removeEventListener('touchstart', handleTouchStart);
      document.removeEventListener('touchmove', handleTouchMove);
      document.removeEventListener('touchend', handleTouchEnd);
    };
  }, [isDrawerOpen]);

  // Keyboard navigation
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isDrawerOpen) {
        setIsDrawerOpen(false);
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [isDrawerOpen]);

  return (
    <>
      {/* Mobile Header */}
      <div className={`lg:hidden bg-white border-b border-gray-200 px-4 py-3 flex items-center justify-between ${className}`}>
        <div className="flex items-center space-x-3">
          <button
            onClick={() => setIsDrawerOpen(true)}
            className="p-2 rounded-md text-gray-600 hover:text-gray-900 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
            aria-label="Open navigation menu"
          >
            <Menu className="w-6 h-6" />
          </button>
          
          <div className="flex items-center space-x-2">
            <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
              <span className="text-white text-sm font-bold">O</span>
            </div>
            <span className="font-semibold text-gray-900">OllamaMax</span>
          </div>
        </div>

        <div className="flex items-center space-x-2">
          <button
            className="p-2 rounded-md text-gray-600 hover:text-gray-900 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500 relative"
            aria-label="Notifications"
          >
            <Bell className="w-5 h-5" />
            <span className="absolute -top-1 -right-1 w-4 h-4 bg-red-500 text-white text-xs rounded-full flex items-center justify-center">
              3
            </span>
          </button>
          
          <button
            className="p-2 rounded-md text-gray-600 hover:text-gray-900 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
            aria-label="User profile"
          >
            <User className="w-5 h-5" />
          </button>
        </div>
      </div>

      {/* Mobile Drawer */}
      {isDrawerOpen && (
        <div className="lg:hidden fixed inset-0 z-50 flex">
          {/* Backdrop */}
          <div
            className="fixed inset-0 bg-black bg-opacity-50 transition-opacity"
            onClick={() => setIsDrawerOpen(false)}
            aria-hidden="true"
          />
          
          {/* Drawer */}
          <div className="relative flex-1 flex flex-col max-w-xs w-full bg-white shadow-xl">
            {/* Drawer Header */}
            <div className="px-4 py-6 border-b border-gray-200">
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <div className="w-10 h-10 bg-blue-600 rounded-xl flex items-center justify-center">
                    <span className="text-white text-lg font-bold">O</span>
                  </div>
                  <div>
                    <h2 className="text-lg font-semibold text-gray-900">OllamaMax</h2>
                    <p className="text-sm text-gray-600">Distributed AI Platform</p>
                  </div>
                </div>
                
                <button
                  onClick={() => setIsDrawerOpen(false)}
                  className="p-2 rounded-md text-gray-400 hover:text-gray-600 hover:bg-gray-100 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  aria-label="Close navigation menu"
                >
                  <X className="w-6 h-6" />
                </button>
              </div>
              
              {/* Search */}
              <div className="mt-4 relative">
                <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" />
                <input
                  type="text"
                  placeholder="Search navigation..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
            </div>

            {/* Navigation Items */}
            <nav className="flex-1 px-4 py-4 overflow-y-auto">
              <div className="space-y-1">
                {filteredItems.map((item) => {
                  const isActive = location.pathname === item.path;
                  const Icon = item.icon;
                  
                  return (
                    <button
                      key={item.id}
                      onClick={() => handleNavigate(item.path)}
                      className={`
                        w-full flex items-center px-3 py-3 rounded-lg text-left transition-colors group
                        ${isActive 
                          ? 'bg-blue-50 text-blue-700 border-l-4 border-blue-600' 
                          : 'text-gray-700 hover:bg-gray-50 hover:text-gray-900'
                        }
                      `}
                      aria-current={isActive ? 'page' : undefined}
                    >
                      <Icon className={`
                        w-5 h-5 mr-3 flex-shrink-0
                        ${isActive ? 'text-blue-600' : 'text-gray-400 group-hover:text-gray-600'}
                      `} />
                      
                      <div className="flex-1 min-w-0">
                        <div className="flex items-center justify-between">
                          <span className="font-medium truncate">{item.label}</span>
                          {item.badge && (
                            <span className="ml-2 bg-red-100 text-red-800 text-xs font-medium px-2 py-1 rounded-full">
                              {item.badge}
                            </span>
                          )}
                        </div>
                        {item.description && (
                          <p className="text-sm text-gray-500 truncate mt-0.5">
                            {item.description}
                          </p>
                        )}
                      </div>
                      
                      <ChevronRight className="w-4 h-4 text-gray-400 ml-2" />
                    </button>
                  );
                })}
              </div>
            </nav>

            {/* Drawer Footer */}
            <div className="px-4 py-4 border-t border-gray-200">
              <div className="flex items-center space-x-3 p-3 bg-gray-50 rounded-lg">
                <div className="w-8 h-8 bg-green-500 rounded-full flex items-center justify-center">
                  <span className="w-2 h-2 bg-white rounded-full"></span>
                </div>
                <div className="flex-1 min-w-0">
                  <p className="text-sm font-medium text-gray-900">System Status</p>
                  <p className="text-xs text-gray-600">All systems operational</p>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Bottom Tab Bar for very small screens */}
      <div className="sm:hidden fixed bottom-0 left-0 right-0 bg-white border-t border-gray-200 z-40">
        <div className="grid grid-cols-4 gap-0">
          {navigationItems.slice(0, 4).map((item) => {
            const isActive = location.pathname === item.path;
            const Icon = item.icon;
            
            return (
              <button
                key={item.id}
                onClick={() => handleNavigate(item.path)}
                className={`
                  flex flex-col items-center px-2 py-2 text-xs transition-colors relative
                  ${isActive 
                    ? 'text-blue-600 bg-blue-50' 
                    : 'text-gray-600 hover:text-gray-900'
                  }
                `}
                aria-current={isActive ? 'page' : undefined}
              >
                <Icon className="w-5 h-5 mb-1" />
                <span className="truncate">{item.label}</span>
                {item.badge && (
                  <span className="absolute -top-1 -right-1 w-4 h-4 bg-red-500 text-white text-xs rounded-full flex items-center justify-center">
                    {item.badge}
                  </span>
                )}
                {isActive && (
                  <div className="absolute top-0 left-0 right-0 h-0.5 bg-blue-600"></div>
                )}
              </button>
            );
          })}
        </div>
      </div>
    </>
  );
};