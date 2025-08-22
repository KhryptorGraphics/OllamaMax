import type { Meta, StoryObj } from '@storybook/react'
import { useState, useMemo } from 'react'
import { 
  Table, 
  TableHeader, 
  TableBody, 
  TableRow, 
  TableHead, 
  TableCell,
  TableFooter,
  TableCaption,
  TablePagination,
  TableFilter,
  EmptyState
} from './Table'
import { Button } from '../Button/Button'
import { Badge } from '../Badge/Badge'
import { Avatar } from '../Avatar/Avatar'
import { 
  MoreHorizontal, 
  Download, 
  Plus,
  Edit,
  Trash2,
  Eye,
  CheckCircle,
  XCircle,
  Clock,
  AlertCircle
} from 'lucide-react'

const meta = {
  title: 'Design System/Table',
  component: Table,
  parameters: {
    layout: 'padded',
    docs: {
      description: {
        component: 'A comprehensive table component with sorting, filtering, pagination, selection, and more.'
      }
    }
  },
  tags: ['autodocs'],
} satisfies Meta<typeof Table>

export default meta
type Story = StoryObj<typeof meta>

// Sample data for stories
const sampleData = [
  {
    id: '1',
    name: 'John Doe',
    email: 'john.doe@example.com',
    role: 'Admin',
    status: 'active',
    lastActive: '2024-01-20',
    avatar: 'https://i.pravatar.cc/150?u=john'
  },
  {
    id: '2',
    name: 'Jane Smith',
    email: 'jane.smith@example.com',
    role: 'Editor',
    status: 'active',
    lastActive: '2024-01-19',
    avatar: 'https://i.pravatar.cc/150?u=jane'
  },
  {
    id: '3',
    name: 'Bob Johnson',
    email: 'bob.johnson@example.com',
    role: 'Viewer',
    status: 'inactive',
    lastActive: '2024-01-15',
    avatar: 'https://i.pravatar.cc/150?u=bob'
  },
  {
    id: '4',
    name: 'Alice Brown',
    email: 'alice.brown@example.com',
    role: 'Editor',
    status: 'pending',
    lastActive: '2024-01-18',
    avatar: 'https://i.pravatar.cc/150?u=alice'
  },
  {
    id: '5',
    name: 'Charlie Wilson',
    email: 'charlie.wilson@example.com',
    role: 'Admin',
    status: 'active',
    lastActive: '2024-01-20',
    avatar: 'https://i.pravatar.cc/150?u=charlie'
  }
]

// Generate large dataset for virtualization demo
const largeDataset = Array.from({ length: 1000 }, (_, i) => ({
  id: `${i + 1}`,
  name: `User ${i + 1}`,
  email: `user${i + 1}@example.com`,
  role: ['Admin', 'Editor', 'Viewer'][Math.floor(Math.random() * 3)],
  status: ['active', 'inactive', 'pending'][Math.floor(Math.random() * 3)],
  lastActive: new Date(Date.now() - Math.random() * 10000000000).toISOString().split('T')[0],
  avatar: `https://i.pravatar.cc/150?u=user${i + 1}`
}))

// Status badge component
const StatusBadge = ({ status }: { status: string }) => {
  const variants: Record<string, 'success' | 'destructive' | 'warning' | 'default'> = {
    active: 'success',
    inactive: 'destructive',
    pending: 'warning'
  }

  const icons = {
    active: <CheckCircle className="h-3 w-3" />,
    inactive: <XCircle className="h-3 w-3" />,
    pending: <Clock className="h-3 w-3" />
  }

  return (
    <Badge variant={variants[status] || 'default'}>
      <span className="flex items-center gap-1">
        {icons[status as keyof typeof icons]}
        {status}
      </span>
    </Badge>
  )
}

// Basic table
export const Basic: Story = {
  render: () => (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Email</TableHead>
          <TableHead>Role</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Last Active</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {sampleData.map((user) => (
          <TableRow key={user.id}>
            <TableCell className="font-medium">{user.name}</TableCell>
            <TableCell>{user.email}</TableCell>
            <TableCell>{user.role}</TableCell>
            <TableCell>
              <StatusBadge status={user.status} />
            </TableCell>
            <TableCell>{user.lastActive}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}

// Table with sorting
export const Sortable: Story = {
  render: () => {
    const [data, setData] = useState(sampleData)

    return (
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead sortKey="name" sortable>Name</TableHead>
            <TableHead sortKey="email" sortable>Email</TableHead>
            <TableHead sortKey="role" sortable>Role</TableHead>
            <TableHead sortKey="status" sortable>Status</TableHead>
            <TableHead sortKey="lastActive" sortable>Last Active</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((user) => (
            <TableRow key={user.id}>
              <TableCell className="font-medium">{user.name}</TableCell>
              <TableCell>{user.email}</TableCell>
              <TableCell>{user.role}</TableCell>
              <TableCell>
                <StatusBadge status={user.status} />
              </TableCell>
              <TableCell>{user.lastActive}</TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    )
  }
}

// Table with row selection
export const WithSelection: Story = {
  render: () => {
    const [selectedRows, setSelectedRows] = useState<Set<string>>(new Set())

    return (
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div className="text-sm text-muted-foreground">
            {selectedRows.size} of {sampleData.length} row(s) selected
          </div>
          {selectedRows.size > 0 && (
            <div className="flex gap-2">
              <Button variant="outline" size="sm">
                Export Selected
              </Button>
              <Button variant="destructive" size="sm">
                Delete Selected
              </Button>
            </div>
          )}
        </div>

        <Table 
          selectionMode="multiple"
          selectedRows={selectedRows}
          onSelectionChange={setSelectedRows}
        >
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Email</TableHead>
              <TableHead>Role</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {sampleData.map((user) => (
              <TableRow key={user.id} rowId={user.id}>
                <TableCell className="font-medium">
                  <div className="flex items-center gap-3">
                    <Avatar src={user.avatar} alt={user.name} size="sm" />
                    {user.name}
                  </div>
                </TableCell>
                <TableCell>{user.email}</TableCell>
                <TableCell>{user.role}</TableCell>
                <TableCell>
                  <StatusBadge status={user.status} />
                </TableCell>
                <TableCell>
                  <Button variant="ghost" size="icon">
                    <MoreHorizontal className="h-4 w-4" />
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    )
  }
}

// Table with filtering and pagination
export const WithFilteringAndPagination: Story = {
  render: () => {
    const [filter, setFilter] = useState('')
    const [currentPage, setCurrentPage] = useState(1)
    const [pageSize, setPageSize] = useState(10)

    const filteredData = useMemo(() => {
      return largeDataset.filter(user => 
        user.name.toLowerCase().includes(filter.toLowerCase()) ||
        user.email.toLowerCase().includes(filter.toLowerCase()) ||
        user.role.toLowerCase().includes(filter.toLowerCase())
      )
    }, [filter])

    const paginatedData = useMemo(() => {
      const start = (currentPage - 1) * pageSize
      const end = start + pageSize
      return filteredData.slice(start, end)
    }, [filteredData, currentPage, pageSize])

    const totalPages = Math.ceil(filteredData.length / pageSize)

    return (
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <TableFilter 
            value={filter}
            onChange={setFilter}
            placeholder="Search users..."
          />
          <Button>
            <Plus className="h-4 w-4 mr-2" />
            Add User
          </Button>
        </div>

        <Table>
          <TableHeader>
            <TableRow>
              <TableHead sortKey="name" sortable>Name</TableHead>
              <TableHead sortKey="email" sortable>Email</TableHead>
              <TableHead sortKey="role" sortable>Role</TableHead>
              <TableHead sortKey="status" sortable>Status</TableHead>
              <TableHead>Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {paginatedData.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5}>
                  <EmptyState
                    title="No users found"
                    description="Try adjusting your search filters"
                  />
                </TableCell>
              </TableRow>
            ) : (
              paginatedData.map((user) => (
                <TableRow key={user.id}>
                  <TableCell className="font-medium">{user.name}</TableCell>
                  <TableCell>{user.email}</TableCell>
                  <TableCell>{user.role}</TableCell>
                  <TableCell>
                    <StatusBadge status={user.status} />
                  </TableCell>
                  <TableCell>
                    <div className="flex gap-1">
                      <Button variant="ghost" size="icon" title="View">
                        <Eye className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" title="Edit">
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" title="Delete">
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>

        <TablePagination
          currentPage={currentPage}
          totalPages={totalPages}
          pageSize={pageSize}
          totalItems={filteredData.length}
          onPageChange={setCurrentPage}
          onPageSizeChange={(newSize) => {
            setPageSize(newSize)
            setCurrentPage(1)
          }}
        />
      </div>
    )
  }
}

// Table with sticky header
export const StickyHeader: Story = {
  render: () => (
    <div className="h-[400px]">
      <Table stickyHeader>
        <TableHeader>
          <TableRow>
            <TableHead>ID</TableHead>
            <TableHead>Name</TableHead>
            <TableHead>Email</TableHead>
            <TableHead>Role</TableHead>
            <TableHead>Status</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {largeDataset.slice(0, 50).map((user) => (
            <TableRow key={user.id}>
              <TableCell>{user.id}</TableCell>
              <TableCell className="font-medium">{user.name}</TableCell>
              <TableCell>{user.email}</TableCell>
              <TableCell>{user.role}</TableCell>
              <TableCell>
                <StatusBadge status={user.status} />
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}

// Table with resizable columns
export const ResizableColumns: Story = {
  render: () => (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead resizable minWidth={100} maxWidth={300}>Name</TableHead>
          <TableHead resizable minWidth={150} maxWidth={400}>Email</TableHead>
          <TableHead resizable minWidth={80} maxWidth={200}>Role</TableHead>
          <TableHead resizable minWidth={80} maxWidth={200}>Status</TableHead>
          <TableHead>Last Active</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {sampleData.map((user) => (
          <TableRow key={user.id}>
            <TableCell className="font-medium">{user.name}</TableCell>
            <TableCell>{user.email}</TableCell>
            <TableCell>{user.role}</TableCell>
            <TableCell>
              <StatusBadge status={user.status} />
            </TableCell>
            <TableCell>{user.lastActive}</TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}

// Table with different densities
export const Densities: Story = {
  render: () => {
    const [density, setDensity] = useState<'compact' | 'normal' | 'comfortable'>('normal')

    return (
      <div className="space-y-4">
        <div className="flex gap-2">
          <Button 
            variant={density === 'compact' ? 'primary' : 'outline'}
            size="sm"
            onClick={() => setDensity('compact')}
          >
            Compact
          </Button>
          <Button 
            variant={density === 'normal' ? 'primary' : 'outline'}
            size="sm"
            onClick={() => setDensity('normal')}
          >
            Normal
          </Button>
          <Button 
            variant={density === 'comfortable' ? 'primary' : 'outline'}
            size="sm"
            onClick={() => setDensity('comfortable')}
          >
            Comfortable
          </Button>
        </div>

        <Table density={density}>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Email</TableHead>
              <TableHead>Role</TableHead>
              <TableHead>Status</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {sampleData.map((user) => (
              <TableRow key={user.id}>
                <TableCell className="font-medium">{user.name}</TableCell>
                <TableCell>{user.email}</TableCell>
                <TableCell>{user.role}</TableCell>
                <TableCell>
                  <StatusBadge status={user.status} />
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    )
  }
}

// Table with loading state
export const LoadingState: Story = {
  render: () => (
    <Table isLoading>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Email</TableHead>
          <TableHead>Role</TableHead>
          <TableHead>Status</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {sampleData.map((user) => (
          <TableRow key={user.id}>
            <TableCell className="font-medium">{user.name}</TableCell>
            <TableCell>{user.email}</TableCell>
            <TableCell>{user.role}</TableCell>
            <TableCell>
              <StatusBadge status={user.status} />
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}

// Table with empty state
export const EmptyStateExample: Story = {
  render: () => (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Email</TableHead>
          <TableHead>Role</TableHead>
          <TableHead>Status</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        <TableRow>
          <TableCell colSpan={4}>
            <EmptyState
              title="No users found"
              description="Get started by adding your first user"
              action={
                <Button>
                  <Plus className="h-4 w-4 mr-2" />
                  Add User
                </Button>
              }
            />
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>
  )
}

// Responsive mobile table
export const MobileResponsive: Story = {
  render: () => (
    <div className="max-w-sm mx-auto">
      <Table>
        <TableBody>
          {sampleData.map((user) => (
            <TableRow key={user.id} className="flex flex-col sm:table-row">
              <TableCell className="flex justify-between sm:table-cell">
                <span className="font-medium sm:hidden">Name:</span>
                <span className="font-medium">{user.name}</span>
              </TableCell>
              <TableCell className="flex justify-between sm:table-cell">
                <span className="font-medium sm:hidden">Email:</span>
                {user.email}
              </TableCell>
              <TableCell className="flex justify-between sm:table-cell">
                <span className="font-medium sm:hidden">Role:</span>
                {user.role}
              </TableCell>
              <TableCell className="flex justify-between sm:table-cell">
                <span className="font-medium sm:hidden">Status:</span>
                <StatusBadge status={user.status} />
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}

// Table with footer
export const WithFooter: Story = {
  render: () => {
    const totals = {
      admin: sampleData.filter(u => u.role === 'Admin').length,
      editor: sampleData.filter(u => u.role === 'Editor').length,
      viewer: sampleData.filter(u => u.role === 'Viewer').length,
      active: sampleData.filter(u => u.status === 'active').length
    }

    return (
      <Table>
        <TableCaption>User management table with role distribution</TableCaption>
        <TableHeader>
          <TableRow>
            <TableHead>Name</TableHead>
            <TableHead>Email</TableHead>
            <TableHead>Role</TableHead>
            <TableHead>Status</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {sampleData.map((user) => (
            <TableRow key={user.id}>
              <TableCell className="font-medium">{user.name}</TableCell>
              <TableCell>{user.email}</TableCell>
              <TableCell>{user.role}</TableCell>
              <TableCell>
                <StatusBadge status={user.status} />
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
        <TableFooter>
          <TableRow>
            <TableCell colSpan={2}>Total Users: {sampleData.length}</TableCell>
            <TableCell>
              Admin: {totals.admin}, Editor: {totals.editor}, Viewer: {totals.viewer}
            </TableCell>
            <TableCell>Active: {totals.active}</TableCell>
          </TableRow>
        </TableFooter>
      </Table>
    )
  }
}

// Complex table with all features
export const CompleteExample: Story = {
  render: () => {
    const [selectedRows, setSelectedRows] = useState<Set<string>>(new Set())
    const [filter, setFilter] = useState('')
    const [currentPage, setCurrentPage] = useState(1)
    const [pageSize, setPageSize] = useState(5)

    const filteredData = useMemo(() => {
      return sampleData.filter(user => 
        user.name.toLowerCase().includes(filter.toLowerCase()) ||
        user.email.toLowerCase().includes(filter.toLowerCase())
      )
    }, [filter])

    const paginatedData = useMemo(() => {
      const start = (currentPage - 1) * pageSize
      const end = start + pageSize
      return filteredData.slice(start, end)
    }, [filteredData, currentPage, pageSize])

    const totalPages = Math.ceil(filteredData.length / pageSize)

    return (
      <div className="space-y-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <TableFilter 
              value={filter}
              onChange={setFilter}
              placeholder="Search users..."
            />
            {selectedRows.size > 0 && (
              <div className="text-sm text-muted-foreground">
                {selectedRows.size} selected
              </div>
            )}
          </div>
          <div className="flex gap-2">
            {selectedRows.size > 0 && (
              <>
                <Button variant="outline" size="sm">
                  <Download className="h-4 w-4 mr-2" />
                  Export
                </Button>
                <Button variant="destructive" size="sm">
                  <Trash2 className="h-4 w-4 mr-2" />
                  Delete
                </Button>
              </>
            )}
            <Button>
              <Plus className="h-4 w-4 mr-2" />
              Add User
            </Button>
          </div>
        </div>

        <Table 
          selectionMode="multiple"
          selectedRows={selectedRows}
          onSelectionChange={setSelectedRows}
          stickyHeader
          density="normal"
        >
          <TableHeader>
            <TableRow>
              <TableHead sortKey="name" sortable resizable>User</TableHead>
              <TableHead sortKey="email" sortable resizable>Email</TableHead>
              <TableHead sortKey="role" sortable>Role</TableHead>
              <TableHead sortKey="status" sortable>Status</TableHead>
              <TableHead sortKey="lastActive" sortable>Last Active</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {paginatedData.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7}>
                  <EmptyState
                    title="No users found"
                    description="Try adjusting your search filters or add a new user"
                    action={
                      <Button>
                        <Plus className="h-4 w-4 mr-2" />
                        Add User
                      </Button>
                    }
                  />
                </TableCell>
              </TableRow>
            ) : (
              paginatedData.map((user) => (
                <TableRow key={user.id} rowId={user.id}>
                  <TableCell>
                    <div className="flex items-center gap-3">
                      <Avatar src={user.avatar} alt={user.name} size="sm" />
                      <div>
                        <div className="font-medium">{user.name}</div>
                        <div className="text-xs text-muted-foreground">ID: {user.id}</div>
                      </div>
                    </div>
                  </TableCell>
                  <TableCell>{user.email}</TableCell>
                  <TableCell>
                    <Badge variant="outline">{user.role}</Badge>
                  </TableCell>
                  <TableCell>
                    <StatusBadge status={user.status} />
                  </TableCell>
                  <TableCell>{user.lastActive}</TableCell>
                  <TableCell>
                    <div className="flex justify-end gap-1">
                      <Button variant="ghost" size="icon" title="View">
                        <Eye className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" title="Edit">
                        <Edit className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" title="Delete">
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>

        <TablePagination
          currentPage={currentPage}
          totalPages={totalPages}
          pageSize={pageSize}
          totalItems={filteredData.length}
          onPageChange={setCurrentPage}
          onPageSizeChange={(newSize) => {
            setPageSize(newSize)
            setCurrentPage(1)
          }}
          pageSizeOptions={[5, 10, 20, 50]}
        />
      </div>
    )
  }
}