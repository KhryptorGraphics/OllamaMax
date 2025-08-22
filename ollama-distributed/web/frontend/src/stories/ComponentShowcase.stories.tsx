import { Meta } from '@storybook/react'
import { Button } from '../design-system/components/Button/Button'
import { Input } from '../design-system/components/Input/Input'
import { Alert } from '../design-system/components/Alert/Alert'
import { Card } from '../design-system/components/Card/Card'
import { Badge } from '../design-system/components/Badge/Badge'
import { useState } from 'react'
import { 
  Search, 
  Mail, 
  Lock, 
  User, 
  Calendar, 
  MapPin, 
  Star, 
  Heart, 
  Share,
  Download,
  Settings,
  Bell,
  CheckCircle,
  AlertTriangle,
  Info
} from 'lucide-react'

export default {
  title: 'Design System/Component Showcase',
  parameters: {
    docs: {
      page: () => (
        <div className="p-6">
          <h1 className="text-3xl font-bold mb-6">Component Showcase</h1>
          <p className="text-lg text-muted-foreground mb-8">
            Explore real-world combinations of components working together to create 
            cohesive user interfaces. These examples demonstrate best practices for 
            component composition and interaction patterns.
          </p>
        </div>
      )
    }
  }
} as Meta

export const DashboardExample = () => {
  const [notifications, setNotifications] = useState([
    { id: 1, type: 'info', title: 'System Update', message: 'New features available', time: '2m ago' },
    { id: 2, type: 'warning', title: 'Maintenance', message: 'Scheduled downtime tonight', time: '1h ago' },
    { id: 3, type: 'success', title: 'Backup Complete', message: 'Daily backup finished', time: '2h ago' }
  ])

  const dismissNotification = (id: number) => {
    setNotifications(prev => prev.filter(n => n.id !== id))
  }

  return (
    <div className="p-6 bg-background min-h-screen">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-3xl font-bold">Dashboard</h1>
            <p className="text-muted-foreground">Welcome back! Here's what's happening.</p>
          </div>
          <div className="flex items-center space-x-4">
            <div className="relative">
              <Search className="w-4 h-4 absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground" />
              <input 
                className="pl-9 pr-4 py-2 w-64 border rounded-md bg-background" 
                placeholder="Search..." 
              />
            </div>
            <Button variant="outline" size="sm">
              <Bell className="w-4 h-4 mr-2" />
              Notifications
              <Badge variant="destructive" size="sm" className="ml-2">3</Badge>
            </Button>
            <Button>
              <Settings className="w-4 h-4 mr-2" />
              Settings
            </Button>
          </div>
        </div>

        {/* Stats Cards */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          <Card>
            <Card.Content className="pt-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">Total Users</p>
                  <p className="text-2xl font-bold">12,345</p>
                  <p className="text-xs text-success flex items-center">
                    <CheckCircle className="w-3 h-3 mr-1" />
                    +12% from last month
                  </p>
                </div>
                <div className="w-12 h-12 bg-primary/10 rounded-full flex items-center justify-center">
                  <User className="w-6 h-6 text-primary" />
                </div>
              </div>
            </Card.Content>
          </Card>

          <Card>
            <Card.Content className="pt-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">Active Sessions</p>
                  <p className="text-2xl font-bold">1,234</p>
                  <p className="text-xs text-warning flex items-center">
                    <AlertTriangle className="w-3 h-3 mr-1" />
                    -5% from yesterday
                  </p>
                </div>
                <div className="w-12 h-12 bg-success/10 rounded-full flex items-center justify-center">
                  <Star className="w-6 h-6 text-success" />
                </div>
              </div>
            </Card.Content>
          </Card>

          <Card>
            <Card.Content className="pt-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">Revenue</p>
                  <p className="text-2xl font-bold">$45,678</p>
                  <p className="text-xs text-success flex items-center">
                    <CheckCircle className="w-3 h-3 mr-1" />
                    +23% from last month
                  </p>
                </div>
                <div className="w-12 h-12 bg-warning/10 rounded-full flex items-center justify-center">
                  <Calendar className="w-6 h-6 text-warning" />
                </div>
              </div>
            </Card.Content>
          </Card>

          <Card>
            <Card.Content className="pt-6">
              <div className="flex items-center justify-between">
                <div>
                  <p className="text-sm text-muted-foreground">Conversion Rate</p>
                  <p className="text-2xl font-bold">3.24%</p>
                  <p className="text-xs text-success flex items-center">
                    <CheckCircle className="w-3 h-3 mr-1" />
                    +0.5% increase
                  </p>
                </div>
                <div className="w-12 h-12 bg-destructive/10 rounded-full flex items-center justify-center">
                  <MapPin className="w-6 h-6 text-destructive" />
                </div>
              </div>
            </Card.Content>
          </Card>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Main Content */}
          <div className="lg:col-span-2 space-y-6">
            {/* Notifications */}
            <Card>
              <Card.Header>
                <Card.Title>Recent Notifications</Card.Title>
                <Card.Description>
                  Stay updated with the latest system alerts and updates
                </Card.Description>
              </Card.Header>
              <Card.Content className="space-y-4">
                {notifications.map(notification => (
                  <Alert 
                    key={notification.id}
                    variant={notification.type as any}
                    dismissible
                    onDismiss={() => dismissNotification(notification.id)}
                    title={notification.title}
                  >
                    <div className="flex items-center justify-between">
                      <span>{notification.message}</span>
                      <span className="text-xs text-muted-foreground">{notification.time}</span>
                    </div>
                  </Alert>
                ))}
                {notifications.length === 0 && (
                  <div className="text-center py-8 text-muted-foreground">
                    <Bell className="w-12 h-12 mx-auto mb-4 opacity-50" />
                    <p>No new notifications</p>
                  </div>
                )}
              </Card.Content>
            </Card>

            {/* Activity Feed */}
            <Card>
              <Card.Header>
                <Card.Title>Recent Activity</Card.Title>
                <Card.Description>
                  Track what's happening in your organization
                </Card.Description>
              </Card.Header>
              <Card.Content>
                <div className="space-y-4">
                  {[
                    { icon: User, action: 'New user registered', detail: 'John Doe joined the platform', time: '2 minutes ago', type: 'success' },
                    { icon: Settings, action: 'System configuration updated', detail: 'Database settings modified', time: '15 minutes ago', type: 'info' },
                    { icon: Download, action: 'Report generated', detail: 'Monthly analytics report', time: '1 hour ago', type: 'default' },
                    { icon: AlertTriangle, action: 'Security alert', detail: 'Unusual login attempt detected', time: '2 hours ago', type: 'warning' }
                  ].map((activity, index) => (
                    <div key={index} className="flex items-start space-x-3">
                      <div className={`w-8 h-8 rounded-full flex items-center justify-center ${
                        activity.type === 'success' ? 'bg-success/10' :
                        activity.type === 'warning' ? 'bg-warning/10' :
                        activity.type === 'info' ? 'bg-info/10' : 'bg-muted'
                      }`}>
                        <activity.icon className="w-4 h-4" />
                      </div>
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium">{activity.action}</p>
                        <p className="text-xs text-muted-foreground">{activity.detail}</p>
                        <p className="text-xs text-muted-foreground mt-1">{activity.time}</p>
                      </div>
                    </div>
                  ))}
                </div>
              </Card.Content>
            </Card>
          </div>

          {/* Sidebar */}
          <div className="space-y-6">
            {/* Quick Actions */}
            <Card>
              <Card.Header>
                <Card.Title>Quick Actions</Card.Title>
              </Card.Header>
              <Card.Content className="space-y-3">
                <Button className="w-full justify-start">
                  <User className="w-4 h-4 mr-2" />
                  Add New User
                </Button>
                <Button variant="outline" className="w-full justify-start">
                  <Download className="w-4 h-4 mr-2" />
                  Export Data
                </Button>
                <Button variant="outline" className="w-full justify-start">
                  <Settings className="w-4 h-4 mr-2" />
                  System Settings
                </Button>
                <Button variant="outline" className="w-full justify-start">
                  <Calendar className="w-4 h-4 mr-2" />
                  Schedule Backup
                </Button>
              </Card.Content>
            </Card>

            {/* System Status */}
            <Card>
              <Card.Header>
                <Card.Title>System Status</Card.Title>
              </Card.Header>
              <Card.Content className="space-y-4">
                <div className="flex items-center justify-between">
                  <span className="text-sm">API Server</span>
                  <Badge variant="success" dot>Online</Badge>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm">Database</span>
                  <Badge variant="success" dot>Online</Badge>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm">File Storage</span>
                  <Badge variant="warning" dot>Maintenance</Badge>
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-sm">Email Service</span>
                  <Badge variant="success" dot>Online</Badge>
                </div>
              </Card.Content>
            </Card>

            {/* Recent Users */}
            <Card>
              <Card.Header>
                <Card.Title>Recent Users</Card.Title>
              </Card.Header>
              <Card.Content className="space-y-3">
                {[
                  { name: 'Alice Johnson', email: 'alice@example.com', status: 'online' },
                  { name: 'Bob Smith', email: 'bob@example.com', status: 'offline' },
                  { name: 'Carol Davis', email: 'carol@example.com', status: 'online' },
                  { name: 'David Wilson', email: 'david@example.com', status: 'away' }
                ].map((user, index) => (
                  <div key={index} className="flex items-center space-x-3">
                    <div className="w-8 h-8 bg-primary/10 rounded-full flex items-center justify-center">
                      <User className="w-4 h-4" />
                    </div>
                    <div className="flex-1 min-w-0">
                      <p className="text-sm font-medium truncate">{user.name}</p>
                      <p className="text-xs text-muted-foreground truncate">{user.email}</p>
                    </div>
                    <Badge 
                      variant={user.status === 'online' ? 'success' : user.status === 'away' ? 'warning' : 'secondary'} 
                      size="sm" 
                      dot
                    >
                      {user.status}
                    </Badge>
                  </div>
                ))}
              </Card.Content>
            </Card>
          </div>
        </div>
      </div>
    </div>
  )
}

export const FormExample = () => {
  const [formData, setFormData] = useState({
    firstName: '',
    lastName: '',
    email: '',
    password: '',
    confirmPassword: '',
    company: '',
    role: '',
    newsletter: false
  })

  const [errors, setErrors] = useState<Record<string, string>>({})
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [submitSuccess, setSubmitSuccess] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsSubmitting(true)
    
    // Simulate validation
    const newErrors: Record<string, string> = {}
    if (!formData.firstName) newErrors.firstName = 'First name is required'
    if (!formData.lastName) newErrors.lastName = 'Last name is required'
    if (!formData.email) newErrors.email = 'Email is required'
    if (!formData.password) newErrors.password = 'Password is required'
    if (formData.password !== formData.confirmPassword) {
      newErrors.confirmPassword = 'Passwords do not match'
    }
    
    setErrors(newErrors)
    
    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 2000))
    
    if (Object.keys(newErrors).length === 0) {
      setSubmitSuccess(true)
      setFormData({
        firstName: '',
        lastName: '',
        email: '',
        password: '',
        confirmPassword: '',
        company: '',
        role: '',
        newsletter: false
      })
    }
    
    setIsSubmitting(false)
  }

  return (
    <div className="p-6 bg-background min-h-screen">
      <div className="max-w-2xl mx-auto">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold mb-4">Create Your Account</h1>
          <p className="text-lg text-muted-foreground">
            Join thousands of users already using OllamaMax
          </p>
        </div>

        {submitSuccess && (
          <Alert variant="success" title="Account Created!" className="mb-6" dismissible onDismiss={() => setSubmitSuccess(false)}>
            Your account has been created successfully. Welcome to OllamaMax!
          </Alert>
        )}

        <Card>
          <Card.Header>
            <Card.Title>Personal Information</Card.Title>
            <Card.Description>
              Please fill in your details to create your account
            </Card.Description>
          </Card.Header>

          <form onSubmit={handleSubmit}>
            <Card.Content className="space-y-6">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Input
                  label="First Name"
                  placeholder="John"
                  value={formData.firstName}
                  onChange={(e) => setFormData(prev => ({ ...prev, firstName: e.target.value }))}
                  error={errors.firstName}
                  required
                  leftIcon={<User />}
                />
                <Input
                  label="Last Name"
                  placeholder="Doe"
                  value={formData.lastName}
                  onChange={(e) => setFormData(prev => ({ ...prev, lastName: e.target.value }))}
                  error={errors.lastName}
                  required
                  leftIcon={<User />}
                />
              </div>

              <Input
                label="Email Address"
                type="email"
                placeholder="john@example.com"
                value={formData.email}
                onChange={(e) => setFormData(prev => ({ ...prev, email: e.target.value }))}
                error={errors.email}
                required
                leftIcon={<Mail />}
                helperText="We'll never share your email with anyone"
              />

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Input
                  label="Password"
                  type="password"
                  placeholder="••••••••"
                  value={formData.password}
                  onChange={(e) => setFormData(prev => ({ ...prev, password: e.target.value }))}
                  error={errors.password}
                  required
                  leftIcon={<Lock />}
                  helperText="Must be at least 8 characters"
                />
                <Input
                  label="Confirm Password"
                  type="password"
                  placeholder="••••••••"
                  value={formData.confirmPassword}
                  onChange={(e) => setFormData(prev => ({ ...prev, confirmPassword: e.target.value }))}
                  error={errors.confirmPassword}
                  required
                  leftIcon={<Lock />}
                />
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <Input
                  label="Company (Optional)"
                  placeholder="Acme Inc."
                  value={formData.company}
                  onChange={(e) => setFormData(prev => ({ ...prev, company: e.target.value }))}
                />
                <div>
                  <label className="text-sm font-medium">Role</label>
                  <select 
                    className="w-full mt-1 px-3 py-2 border rounded-md bg-background"
                    value={formData.role}
                    onChange={(e) => setFormData(prev => ({ ...prev, role: e.target.value }))}
                  >
                    <option value="">Select your role</option>
                    <option value="developer">Developer</option>
                    <option value="designer">Designer</option>
                    <option value="manager">Manager</option>
                    <option value="other">Other</option>
                  </select>
                </div>
              </div>

              <div className="flex items-center space-x-2">
                <input
                  type="checkbox"
                  id="newsletter"
                  checked={formData.newsletter}
                  onChange={(e) => setFormData(prev => ({ ...prev, newsletter: e.target.checked }))}
                  className="rounded border border-input"
                />
                <label htmlFor="newsletter" className="text-sm">
                  Subscribe to our newsletter for updates and tips
                </label>
              </div>

              <div className="p-4 bg-muted/50 rounded-lg">
                <p className="text-sm text-muted-foreground">
                  By creating an account, you agree to our{' '}
                  <a href="#" className="text-primary hover:underline">Terms of Service</a>
                  {' '}and{' '}
                  <a href="#" className="text-primary hover:underline">Privacy Policy</a>.
                </p>
              </div>
            </Card.Content>

            <Card.Footer className="flex justify-between">
              <Button variant="outline" type="button">
                Back to Login
              </Button>
              <Button 
                type="submit" 
                loading={isSubmitting}
                loadingText="Creating Account..."
                disabled={isSubmitting}
              >
                Create Account
              </Button>
            </Card.Footer>
          </form>
        </Card>

        <div className="text-center mt-6">
          <p className="text-sm text-muted-foreground">
            Already have an account?{' '}
            <a href="#" className="text-primary hover:underline">Sign in here</a>
          </p>
        </div>
      </div>
    </div>
  )
}

export const ECommerceCard = () => {
  const [favorites, setFavorites] = useState<Set<number>>(new Set())
  const [cart, setCart] = useState<Set<number>>(new Set())

  const toggleFavorite = (id: number) => {
    const newFavorites = new Set(favorites)
    if (newFavorites.has(id)) {
      newFavorites.delete(id)
    } else {
      newFavorites.add(id)
    }
    setFavorites(newFavorites)
  }

  const toggleCart = (id: number) => {
    const newCart = new Set(cart)
    if (newCart.has(id)) {
      newCart.delete(id)
    } else {
      newCart.add(id)
    }
    setCart(newCart)
  }

  const products = [
    {
      id: 1,
      name: 'Wireless Headphones',
      price: '$99.99',
      originalPrice: '$129.99',
      rating: 4.5,
      reviews: 128,
      image: 'https://images.unsplash.com/photo-1505740420928-5e560c06d30e?w=300&h=300&fit=crop',
      badges: ['Best Seller', 'Free Shipping'],
      inStock: true
    },
    {
      id: 2,
      name: 'Smart Watch',
      price: '$249.99',
      originalPrice: null,
      rating: 4.8,
      reviews: 89,
      image: 'https://images.unsplash.com/photo-1523275335684-37898b6baf30?w=300&h=300&fit=crop',
      badges: ['New Arrival'],
      inStock: true
    },
    {
      id: 3,
      name: 'Bluetooth Speaker',
      price: '$79.99',
      originalPrice: '$99.99',
      rating: 4.2,
      reviews: 245,
      image: 'https://images.unsplash.com/photo-1608043152269-423dbba4e7e1?w=300&h=300&fit=crop',
      badges: ['Sale', 'Limited Time'],
      inStock: false
    }
  ]

  return (
    <div className="p-6 bg-background">
      <div className="max-w-6xl mx-auto">
        <div className="text-center mb-8">
          <h1 className="text-3xl font-bold mb-4">Featured Products</h1>
          <p className="text-lg text-muted-foreground">
            Discover our most popular items with great reviews and competitive prices
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {products.map((product) => (
            <Card key={product.id} className="overflow-hidden hover:shadow-lg transition-shadow">
              <div className="relative">
                <img 
                  src={product.image} 
                  alt={product.name}
                  className="w-full h-48 object-cover"
                />
                
                {/* Badges */}
                <div className="absolute top-3 left-3 flex flex-wrap gap-1">
                  {product.badges.map((badge, index) => (
                    <Badge 
                      key={index}
                      variant={badge === 'Sale' ? 'destructive' : badge === 'New Arrival' ? 'success' : 'default'}
                      size="sm"
                    >
                      {badge}
                    </Badge>
                  ))}
                </div>

                {/* Favorite Button */}
                <button
                  onClick={() => toggleFavorite(product.id)}
                  className="absolute top-3 right-3 p-2 bg-background/80 backdrop-blur-sm rounded-full hover:bg-background transition-colors"
                >
                  <Heart 
                    className={`w-4 h-4 ${
                      favorites.has(product.id) 
                        ? 'fill-red-500 text-red-500' 
                        : 'text-muted-foreground'
                    }`}
                  />
                </button>

                {/* Stock Status */}
                {!product.inStock && (
                  <div className="absolute inset-0 bg-background/80 backdrop-blur-sm flex items-center justify-center">
                    <Badge variant="secondary" size="lg">Out of Stock</Badge>
                  </div>
                )}
              </div>

              <Card.Content className="p-4">
                <div className="space-y-3">
                  <div>
                    <h3 className="font-semibold text-lg">{product.name}</h3>
                    
                    <div className="flex items-center space-x-2 mt-1">
                      <div className="flex items-center">
                        {Array.from({ length: 5 }, (_, i) => (
                          <Star 
                            key={i}
                            className={`w-3 h-3 ${
                              i < Math.floor(product.rating) 
                                ? 'fill-yellow-400 text-yellow-400' 
                                : 'text-gray-300'
                            }`}
                          />
                        ))}
                      </div>
                      <span className="text-sm text-muted-foreground">
                        {product.rating} ({product.reviews} reviews)
                      </span>
                    </div>
                  </div>

                  <div className="flex items-center space-x-2">
                    <span className="text-xl font-bold">{product.price}</span>
                    {product.originalPrice && (
                      <span className="text-sm text-muted-foreground line-through">
                        {product.originalPrice}
                      </span>
                    )}
                    {product.originalPrice && (
                      <Badge variant="destructive" size="sm">
                        {Math.round((1 - parseFloat(product.price.replace('$', '')) / parseFloat(product.originalPrice.replace('$', ''))) * 100)}% OFF
                      </Badge>
                    )}
                  </div>

                  <div className="flex space-x-2">
                    <Button 
                      className="flex-1"
                      disabled={!product.inStock}
                      onClick={() => toggleCart(product.id)}
                      variant={cart.has(product.id) ? 'secondary' : 'default'}
                    >
                      {cart.has(product.id) ? 'Remove from Cart' : 'Add to Cart'}
                    </Button>
                    
                    <Button variant="outline" size="sm">
                      <Share className="w-4 h-4" />
                    </Button>
                  </div>
                </div>
              </Card.Content>
            </Card>
          ))}
        </div>

        <div className="text-center mt-8">
          <Button variant="outline" size="lg">
            View All Products
          </Button>
        </div>
      </div>
    </div>
  )
}

export const NotificationCenter = () => {
  const [notifications, setNotifications] = useState([
    {
      id: 1,
      type: 'info' as const,
      title: 'System Update Available',
      message: 'A new version of the system is ready to install. Update now to get the latest features and security improvements.',
      time: '2 minutes ago',
      read: false,
      actions: ['Update Now', 'Remind Later']
    },
    {
      id: 2,
      type: 'success' as const,
      title: 'Backup Completed',
      message: 'Your daily backup has been completed successfully. All data is safely stored.',
      time: '1 hour ago',
      read: false,
      actions: ['View Details']
    },
    {
      id: 3,
      type: 'warning' as const,
      title: 'Storage Almost Full',
      message: 'Your storage is 85% full. Consider upgrading your plan or cleaning up old files.',
      time: '3 hours ago',
      read: true,
      actions: ['Upgrade Plan', 'Manage Files']
    },
    {
      id: 4,
      type: 'destructive' as const,
      title: 'Failed Login Attempt',
      message: 'Someone tried to access your account from an unrecognized device. Please review your security settings.',
      time: '5 hours ago',
      read: true,
      actions: ['Review Security', 'Change Password']
    }
  ])

  const markAsRead = (id: number) => {
    setNotifications(prev => 
      prev.map(n => n.id === id ? { ...n, read: true } : n)
    )
  }

  const markAllAsRead = () => {
    setNotifications(prev => prev.map(n => ({ ...n, read: true })))
  }

  const dismissNotification = (id: number) => {
    setNotifications(prev => prev.filter(n => n.id !== id))
  }

  const unreadCount = notifications.filter(n => !n.read).length

  return (
    <div className="p-6 bg-background min-h-screen">
      <div className="max-w-4xl mx-auto">
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-3xl font-bold flex items-center">
              <Bell className="w-8 h-8 mr-3" />
              Notifications
              {unreadCount > 0 && (
                <Badge variant="destructive" className="ml-3">
                  {unreadCount} new
                </Badge>
              )}
            </h1>
            <p className="text-muted-foreground mt-2">
              Stay updated with the latest alerts and system status
            </p>
          </div>
          
          <div className="flex space-x-2">
            <Button variant="outline" onClick={markAllAsRead} disabled={unreadCount === 0}>
              <CheckCircle className="w-4 h-4 mr-2" />
              Mark All Read
            </Button>
            <Button variant="outline">
              <Settings className="w-4 h-4 mr-2" />
              Settings
            </Button>
          </div>
        </div>

        {notifications.length === 0 ? (
          <Card>
            <Card.Content className="pt-12 pb-12 text-center">
              <Bell className="w-16 h-16 mx-auto mb-4 text-muted-foreground opacity-50" />
              <h3 className="text-lg font-medium mb-2">No notifications</h3>
              <p className="text-muted-foreground">
                You're all caught up! New notifications will appear here.
              </p>
            </Card.Content>
          </Card>
        ) : (
          <div className="space-y-4">
            {notifications.map((notification) => (
              <Card 
                key={notification.id} 
                className={`transition-all ${!notification.read ? 'border-l-4 border-l-primary bg-primary/5' : ''}`}
              >
                <Card.Content className="p-0">
                  <Alert
                    variant={notification.type}
                    title={notification.title}
                    dismissible
                    onDismiss={() => dismissNotification(notification.id)}
                    actions={
                      <div className="flex flex-wrap gap-2 mt-3">
                        {notification.actions.map((action, index) => (
                          <Button 
                            key={index} 
                            size="sm" 
                            variant={index === 0 ? 'default' : 'outline'}
                          >
                            {action}
                          </Button>
                        ))}
                        {!notification.read && (
                          <Button 
                            size="sm" 
                            variant="ghost" 
                            onClick={() => markAsRead(notification.id)}
                          >
                            Mark as Read
                          </Button>
                        )}
                      </div>
                    }
                  >
                    <div className="space-y-2">
                      <p>{notification.message}</p>
                      <div className="flex items-center justify-between">
                        <span className="text-xs text-muted-foreground">
                          {notification.time}
                        </span>
                        {!notification.read && (
                          <Badge variant="default" size="sm">
                            New
                          </Badge>
                        )}
                      </div>
                    </div>
                  </Alert>
                </Card.Content>
              </Card>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}