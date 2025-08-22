import React, { 
  forwardRef, 
  useState, 
  useEffect, 
  useRef, 
  useCallback, 
  useMemo,
  createContext,
  useContext,
  ReactNode,
  HTMLAttributes,
  ThHTMLAttributes,
  TdHTMLAttributes
} from 'react'
import { cva, type VariantProps } from 'class-variance-authority'
import { 
  ChevronUp, 
  ChevronDown, 
  ChevronsUpDown, 
  Search, 
  X, 
  CheckSquare,
  Square,
  MinusSquare,
  Loader2,
  AlertCircle,
  FileX2
} from 'lucide-react'
import { cn } from '@/utils/cn'
import { Button } from '../Button/Button'
import { Input } from '../Input/Input'

// Table context for managing state
interface TableContextValue {
  selectedRows: Set<string>
  setSelectedRows: React.Dispatch<React.SetStateAction<Set<string>>>
  sortColumn: string | null
  sortDirection: 'asc' | 'desc' | null
  setSortColumn: (column: string | null) => void
  setSortDirection: (direction: 'asc' | 'desc' | null) => void
  selectionMode: 'none' | 'single' | 'multiple'
  isLoading?: boolean
  stickyHeader?: boolean
  density?: 'compact' | 'normal' | 'comfortable'
}

const TableContext = createContext<TableContextValue | undefined>(undefined)

const useTableContext = () => {
  const context = useContext(TableContext)
  if (!context) {
    throw new Error('Table components must be used within a Table')
  }
  return context
}

// Table variants using design tokens
const tableVariants = cva(
  'w-full caption-bottom text-sm',
  {
    variants: {
      variant: {
        default: 'border-collapse',
        bordered: 'border border-border rounded-lg overflow-hidden',
        striped: 'border-collapse'
      },
      density: {
        compact: '',
        normal: '',
        comfortable: ''
      }
    },
    defaultVariants: {
      variant: 'default',
      density: 'normal'
    }
  }
)

// Main Table component
export interface TableProps extends HTMLAttributes<HTMLTableElement>, VariantProps<typeof tableVariants> {
  /** Selection mode for rows */
  selectionMode?: 'none' | 'single' | 'multiple'
  /** Selected row IDs */
  selectedRows?: Set<string>
  /** Callback when selection changes */
  onSelectionChange?: (selectedRows: Set<string>) => void
  /** Loading state */
  isLoading?: boolean
  /** Sticky header */
  stickyHeader?: boolean
  /** Table density */
  density?: 'compact' | 'normal' | 'comfortable'
  /** Enable virtualization for large datasets */
  virtualized?: boolean
  /** Number of rows to render in virtualized mode */
  virtualRowHeight?: number
  /** Total number of rows (for virtualization) */
  totalRows?: number
  children?: ReactNode
}

export const Table = forwardRef<HTMLTableElement, TableProps>(
  ({ 
    className,
    variant,
    density = 'normal',
    selectionMode = 'none',
    selectedRows: controlledSelectedRows,
    onSelectionChange,
    isLoading = false,
    stickyHeader = false,
    virtualized = false,
    virtualRowHeight = 48,
    totalRows = 0,
    children,
    ...props 
  }, ref) => {
    const [internalSelectedRows, setInternalSelectedRows] = useState<Set<string>>(new Set())
    const [sortColumn, setSortColumn] = useState<string | null>(null)
    const [sortDirection, setSortDirection] = useState<'asc' | 'desc' | null>(null)

    const selectedRows = controlledSelectedRows ?? internalSelectedRows
    const setSelectedRows = useCallback((value: React.SetStateAction<Set<string>>) => {
      const newValue = typeof value === 'function' ? value(selectedRows) : value
      if (!controlledSelectedRows) {
        setInternalSelectedRows(newValue)
      }
      onSelectionChange?.(newValue)
    }, [selectedRows, controlledSelectedRows, onSelectionChange])

    const contextValue: TableContextValue = {
      selectedRows,
      setSelectedRows,
      sortColumn,
      sortDirection,
      setSortColumn,
      setSortDirection,
      selectionMode,
      isLoading,
      stickyHeader,
      density
    }

    return (
      <TableContext.Provider value={contextValue}>
        <div className={cn(
          'relative w-full overflow-auto',
          stickyHeader && 'max-h-[600px]',
          className
        )}>
          <table
            ref={ref}
            className={cn(tableVariants({ variant, density }))}
            {...props}
          >
            {children}
          </table>
          
          {/* Loading overlay */}
          {isLoading && (
            <div className="absolute inset-0 bg-background/50 backdrop-blur-sm flex items-center justify-center z-10">
              <div className="flex items-center gap-2 text-muted-foreground">
                <Loader2 className="h-4 w-4 animate-spin" />
                <span>Loading...</span>
              </div>
            </div>
          )}
        </div>
      </TableContext.Provider>
    )
  }
)

Table.displayName = 'Table'

// TableHeader component
export interface TableHeaderProps extends HTMLAttributes<HTMLTableSectionElement> {}

export const TableHeader = forwardRef<HTMLTableSectionElement, TableHeaderProps>(
  ({ className, ...props }, ref) => {
    const { stickyHeader } = useTableContext()
    
    return (
      <thead
        ref={ref}
        className={cn(
          'bg-secondary-50 dark:bg-secondary-900 border-b border-border',
          stickyHeader && 'sticky top-0 z-20 shadow-sm',
          className
        )}
        {...props}
      />
    )
  }
)

TableHeader.displayName = 'TableHeader'

// TableBody component
export interface TableBodyProps extends HTMLAttributes<HTMLTableSectionElement> {}

export const TableBody = forwardRef<HTMLTableSectionElement, TableBodyProps>(
  ({ className, ...props }, ref) => (
    <tbody
      ref={ref}
      className={cn('[&_tr:last-child]:border-0', className)}
      {...props}
    />
  )
)

TableBody.displayName = 'TableBody'

// TableRow component
export interface TableRowProps extends HTMLAttributes<HTMLTableRowElement> {
  /** Unique identifier for the row */
  rowId?: string
  /** Whether the row is selectable */
  selectable?: boolean
  /** Custom click handler */
  onRowClick?: () => void
}

export const TableRow = forwardRef<HTMLTableRowElement, TableRowProps>(
  ({ className, rowId, selectable = true, onRowClick, children, ...props }, ref) => {
    const { selectedRows, setSelectedRows, selectionMode } = useTableContext()
    const isSelected = rowId ? selectedRows.has(rowId) : false

    const handleRowSelection = useCallback(() => {
      if (!rowId || !selectable || selectionMode === 'none') return

      if (selectionMode === 'single') {
        setSelectedRows(new Set(isSelected ? [] : [rowId]))
      } else if (selectionMode === 'multiple') {
        setSelectedRows(prev => {
          const newSet = new Set(prev)
          if (isSelected) {
            newSet.delete(rowId)
          } else {
            newSet.add(rowId)
          }
          return newSet
        })
      }
    }, [rowId, selectable, selectionMode, isSelected, setSelectedRows])

    const handleClick = useCallback(() => {
      onRowClick?.()
      handleRowSelection()
    }, [onRowClick, handleRowSelection])

    return (
      <tr
        ref={ref}
        className={cn(
          'border-b border-border transition-colors',
          'hover:bg-accent/50',
          isSelected && 'bg-primary-50 dark:bg-primary-900/20',
          selectionMode !== 'none' && selectable && 'cursor-pointer',
          className
        )}
        onClick={handleClick}
        data-selected={isSelected}
        aria-selected={isSelected}
        {...props}
      >
        {selectionMode !== 'none' && selectable && (
          <TableCell className="w-12" onClick={(e) => e.stopPropagation()}>
            <Checkbox
              checked={isSelected}
              onCheckedChange={() => handleRowSelection()}
              aria-label={`Select row ${rowId}`}
            />
          </TableCell>
        )}
        {children}
      </tr>
    )
  }
)

TableRow.displayName = 'TableRow'

// TableHead component (header cell)
export interface TableHeadProps extends ThHTMLAttributes<HTMLTableCellElement> {
  /** Column key for sorting */
  sortKey?: string
  /** Whether the column is sortable */
  sortable?: boolean
  /** Column is resizable */
  resizable?: boolean
  /** Minimum width for resizable columns */
  minWidth?: number
  /** Maximum width for resizable columns */
  maxWidth?: number
}

export const TableHead = forwardRef<HTMLTableCellElement, TableHeadProps>(
  ({ 
    className, 
    sortKey, 
    sortable = false, 
    resizable = false,
    minWidth = 50,
    maxWidth = 500,
    children, 
    ...props 
  }, ref) => {
    const { sortColumn, sortDirection, setSortColumn, setSortDirection, density } = useTableContext()
    const [isResizing, setIsResizing] = useState(false)
    const [columnWidth, setColumnWidth] = useState<number | undefined>(undefined)
    const cellRef = useRef<HTMLTableCellElement>(null)

    const isSorted = sortKey === sortColumn
    const isAsc = isSorted && sortDirection === 'asc'
    const isDesc = isSorted && sortDirection === 'desc'

    const handleSort = useCallback(() => {
      if (!sortable || !sortKey) return

      if (!isSorted) {
        setSortColumn(sortKey)
        setSortDirection('asc')
      } else if (isAsc) {
        setSortDirection('desc')
      } else {
        setSortColumn(null)
        setSortDirection(null)
      }
    }, [sortable, sortKey, isSorted, isAsc, setSortColumn, setSortDirection])

    // Handle column resizing
    const handleMouseDown = useCallback((e: React.MouseEvent) => {
      if (!resizable) return
      e.preventDefault()
      setIsResizing(true)

      const startX = e.pageX
      const startWidth = cellRef.current?.offsetWidth || 0

      const handleMouseMove = (e: MouseEvent) => {
        const newWidth = Math.max(minWidth, Math.min(maxWidth, startWidth + e.pageX - startX))
        setColumnWidth(newWidth)
      }

      const handleMouseUp = () => {
        setIsResizing(false)
        document.removeEventListener('mousemove', handleMouseMove)
        document.removeEventListener('mouseup', handleMouseUp)
      }

      document.addEventListener('mousemove', handleMouseMove)
      document.addEventListener('mouseup', handleMouseUp)
    }, [resizable, minWidth, maxWidth])

    const paddingClass = density === 'compact' ? 'px-3 py-2' : 
                        density === 'comfortable' ? 'px-6 py-4' : 
                        'px-4 py-3'

    return (
      <th
        ref={cellRef}
        className={cn(
          paddingClass,
          'text-left align-middle font-medium text-muted-foreground',
          sortable && 'cursor-pointer select-none hover:text-foreground',
          'relative group',
          className
        )}
        style={{ width: columnWidth }}
        onClick={sortable ? handleSort : undefined}
        aria-sort={
          isSorted ? (isAsc ? 'ascending' : 'descending') : 'none'
        }
        {...props}
      >
        <div className="flex items-center justify-between gap-2">
          <span>{children}</span>
          {sortable && (
            <span className="flex-shrink-0">
              {!isSorted && <ChevronsUpDown className="h-4 w-4 opacity-50" />}
              {isAsc && <ChevronUp className="h-4 w-4" />}
              {isDesc && <ChevronDown className="h-4 w-4" />}
            </span>
          )}
        </div>
        
        {/* Resize handle */}
        {resizable && (
          <div
            className={cn(
              'absolute right-0 top-0 h-full w-1 cursor-col-resize',
              'hover:bg-primary-500 group-hover:bg-border',
              isResizing && 'bg-primary-500'
            )}
            onMouseDown={handleMouseDown}
          />
        )}
      </th>
    )
  }
)

TableHead.displayName = 'TableHead'

// TableCell component
export interface TableCellProps extends TdHTMLAttributes<HTMLTableCellElement> {}

export const TableCell = forwardRef<HTMLTableCellElement, TableCellProps>(
  ({ className, ...props }, ref) => {
    const { density } = useTableContext()
    
    const paddingClass = density === 'compact' ? 'px-3 py-2' : 
                        density === 'comfortable' ? 'px-6 py-4' : 
                        'px-4 py-3'

    return (
      <td
        ref={ref}
        className={cn(paddingClass, 'align-middle', className)}
        {...props}
      />
    )
  }
)

TableCell.displayName = 'TableCell'

// Checkbox component for row selection
interface CheckboxProps {
  checked: boolean | 'indeterminate'
  onCheckedChange: (checked: boolean) => void
  className?: string
  'aria-label'?: string
}

const Checkbox: React.FC<CheckboxProps> = ({ 
  checked, 
  onCheckedChange, 
  className,
  'aria-label': ariaLabel
}) => {
  const handleClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    onCheckedChange(!checked)
  }

  return (
    <button
      type="button"
      role="checkbox"
      aria-checked={checked === 'indeterminate' ? 'mixed' : checked}
      aria-label={ariaLabel}
      className={cn(
        'h-4 w-4 flex items-center justify-center rounded border border-primary-500',
        'focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2',
        'transition-colors',
        checked && 'bg-primary-500 text-white',
        className
      )}
      onClick={handleClick}
    >
      {checked === 'indeterminate' && <MinusSquare className="h-3 w-3" />}
      {checked === true && <CheckSquare className="h-3 w-3" />}
      {checked === false && <Square className="h-3 w-3" />}
    </button>
  )
}

// TableCaption component
export interface TableCaptionProps extends HTMLAttributes<HTMLTableCaptionElement> {}

export const TableCaption = forwardRef<HTMLTableCaptionElement, TableCaptionProps>(
  ({ className, ...props }, ref) => (
    <caption
      ref={ref}
      className={cn('mt-4 text-sm text-muted-foreground', className)}
      {...props}
    />
  )
)

TableCaption.displayName = 'TableCaption'

// TableFooter component
export interface TableFooterProps extends HTMLAttributes<HTMLTableSectionElement> {}

export const TableFooter = forwardRef<HTMLTableSectionElement, TableFooterProps>(
  ({ className, ...props }, ref) => (
    <tfoot
      ref={ref}
      className={cn(
        'bg-secondary-50 dark:bg-secondary-900 font-medium border-t border-border',
        className
      )}
      {...props}
    />
  )
)

TableFooter.displayName = 'TableFooter'

// TablePagination component
export interface TablePaginationProps {
  currentPage: number
  totalPages: number
  pageSize: number
  totalItems: number
  onPageChange: (page: number) => void
  onPageSizeChange?: (pageSize: number) => void
  pageSizeOptions?: number[]
  className?: string
}

export const TablePagination: React.FC<TablePaginationProps> = ({
  currentPage,
  totalPages,
  pageSize,
  totalItems,
  onPageChange,
  onPageSizeChange,
  pageSizeOptions = [10, 20, 30, 50, 100],
  className
}) => {
  const startItem = (currentPage - 1) * pageSize + 1
  const endItem = Math.min(currentPage * pageSize, totalItems)

  return (
    <div className={cn(
      'flex items-center justify-between px-4 py-3',
      'border-t border-border bg-background',
      className
    )}>
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <span>
          Showing {startItem} to {endItem} of {totalItems} results
        </span>
        
        {onPageSizeChange && (
          <select
            value={pageSize}
            onChange={(e) => onPageSizeChange(Number(e.target.value))}
            className="ml-4 px-2 py-1 border border-border rounded-md bg-background"
            aria-label="Items per page"
          >
            {pageSizeOptions.map(size => (
              <option key={size} value={size}>
                {size} per page
              </option>
            ))}
          </select>
        )}
      </div>

      <div className="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange(currentPage - 1)}
          disabled={currentPage === 1}
          aria-label="Previous page"
        >
          Previous
        </Button>

        <div className="flex items-center gap-1">
          {Array.from({ length: Math.min(5, totalPages) }, (_, i) => {
            let pageNum
            if (totalPages <= 5) {
              pageNum = i + 1
            } else if (currentPage <= 3) {
              pageNum = i + 1
            } else if (currentPage >= totalPages - 2) {
              pageNum = totalPages - 4 + i
            } else {
              pageNum = currentPage - 2 + i
            }

            return (
              <Button
                key={pageNum}
                variant={pageNum === currentPage ? 'primary' : 'ghost'}
                size="sm"
                onClick={() => onPageChange(pageNum)}
                className="min-w-[32px]"
                aria-label={`Go to page ${pageNum}`}
                aria-current={pageNum === currentPage ? 'page' : undefined}
              >
                {pageNum}
              </Button>
            )
          })}
        </div>

        <Button
          variant="outline"
          size="sm"
          onClick={() => onPageChange(currentPage + 1)}
          disabled={currentPage === totalPages}
          aria-label="Next page"
        >
          Next
        </Button>
      </div>
    </div>
  )
}

// TableFilter component
export interface TableFilterProps {
  value: string
  onChange: (value: string) => void
  placeholder?: string
  className?: string
}

export const TableFilter: React.FC<TableFilterProps> = ({
  value,
  onChange,
  placeholder = 'Filter...',
  className
}) => {
  return (
    <div className={cn('relative max-w-sm', className)}>
      <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground pointer-events-none" />
      <Input
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="pl-9 pr-9"
        aria-label="Filter table"
      />
      {value && (
        <button
          type="button"
          onClick={() => onChange('')}
          className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
          aria-label="Clear filter"
        >
          <X className="h-4 w-4" />
        </button>
      )}
    </div>
  )
}

// EmptyState component
export interface EmptyStateProps {
  title?: string
  description?: string
  icon?: ReactNode
  action?: ReactNode
  className?: string
}

export const EmptyState: React.FC<EmptyStateProps> = ({
  title = 'No data',
  description = 'No data to display',
  icon = <FileX2 className="h-12 w-12 text-muted-foreground" />,
  action,
  className
}) => {
  return (
    <div className={cn(
      'flex flex-col items-center justify-center py-12 px-4',
      className
    )}>
      {icon}
      <h3 className="mt-4 text-lg font-semibold">{title}</h3>
      <p className="mt-2 text-sm text-muted-foreground text-center max-w-sm">
        {description}
      </p>
      {action && <div className="mt-6">{action}</div>}
    </div>
  )
}

// Export all components
export default Table