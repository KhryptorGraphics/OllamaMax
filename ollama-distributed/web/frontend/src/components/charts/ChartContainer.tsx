/**
 * ChartContainer Component
 * Responsive chart wrapper with loading states, error handling, export functionality,
 * and theme integration for dark/light mode
 */

import React, { useRef, useEffect, useState, useCallback } from 'react'
import { Download, RefreshCw, Maximize2, Settings } from 'lucide-react'
import { colorUtils, exportUtils, responsiveUtils, type ChartTheme } from '@/utils/chartUtils'
import { Button } from '@/design-system/components/Button/Button'
import { Spinner } from '@/design-system/components/Spinner/Spinner'
import { Alert } from '@/design-system/components/Alert/Alert'
import { cn } from '@/utils/cn'

export interface ChartContainerProps {
  /** Chart title */
  title?: string
  
  /** Chart description */
  description?: string
  
  /** Loading state */
  loading?: boolean
  
  /** Error state */
  error?: string | Error | null
  
  /** Chart data for export */
  data?: any[]
  
  /** Theme mode */
  theme?: 'light' | 'dark'
  
  /** Auto-refresh interval in milliseconds */
  autoRefresh?: number
  
  /** Refresh callback */
  onRefresh?: () => void
  
  /** Export filename prefix */
  exportFilename?: string
  
  /** Enable export functionality */
  enableExport?: boolean
  
  /** Enable fullscreen mode */
  enableFullscreen?: boolean
  
  /** Container height */
  height?: number | string
  
  /** Additional CSS classes */
  className?: string
  
  /** Children (chart components) */
  children: React.ReactNode
  
  /** Chart configuration options */
  chartConfig?: {
    responsive?: boolean
    maintainAspectRatio?: boolean
    animations?: boolean
  }
}

export interface ChartContainerRef {
  exportAsPNG: () => Promise<void>
  exportAsSVG: () => void
  exportAsCSV: () => void
  exportAsPDF: () => Promise<void>
  refresh: () => void
  toggleFullscreen: () => void
}

const ChartContainer = React.forwardRef<ChartContainerRef, ChartContainerProps>(({
  title,
  description,
  loading = false,
  error = null,
  data = [],
  theme = 'light',
  autoRefresh,
  onRefresh,
  exportFilename = 'chart',
  enableExport = true,
  enableFullscreen = true,
  height = 400,
  className,
  children,
  chartConfig = {
    responsive: true,
    maintainAspectRatio: true,
    animations: true
  }
}, ref) => {
  const containerRef = useRef<HTMLDivElement>(null)
  const chartRef = useRef<HTMLDivElement>(null)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [showExportMenu, setShowExportMenu] = useState(false)
  const [dimensions, setDimensions] = useState({ width: 0, height: 0 })
  const [chartTheme, setChartTheme] = useState<ChartTheme>(() => 
    colorUtils.generateChartTheme(theme)
  )

  // Update theme when prop changes
  useEffect(() => {
    setChartTheme(colorUtils.generateChartTheme(theme))
  }, [theme])

  // Handle responsive dimensions
  useEffect(() => {
    if (!chartConfig.responsive || !containerRef.current) return

    const updateDimensions = () => {
      if (containerRef.current) {
        const newDimensions = responsiveUtils.getResponsiveDimensions(containerRef.current)
        setDimensions(newDimensions)
      }
    }

    const resizeObserver = new ResizeObserver(updateDimensions)
    resizeObserver.observe(containerRef.current)
    updateDimensions()

    return () => resizeObserver.disconnect()
  }, [chartConfig.responsive])

  // Auto-refresh functionality
  useEffect(() => {
    if (!autoRefresh || !onRefresh) return

    const interval = setInterval(onRefresh, autoRefresh)
    return () => clearInterval(interval)
  }, [autoRefresh, onRefresh])

  // Export functions
  const exportAsPNG = useCallback(async () => {
    if (!chartRef.current) return
    try {
      await exportUtils.exportAsPNG(chartRef.current, `${exportFilename}.png`)
    } catch (error) {
      console.error('Failed to export as PNG:', error)
    }
  }, [exportFilename])

  const exportAsSVG = useCallback(() => {
    if (!chartRef.current) return
    try {
      exportUtils.exportAsSVG(chartRef.current, `${exportFilename}.svg`)
    } catch (error) {
      console.error('Failed to export as SVG:', error)
    }
  }, [exportFilename])

  const exportAsCSV = useCallback(() => {
    try {
      exportUtils.exportAsCSV(data, `${exportFilename}.csv`)
    } catch (error) {
      console.error('Failed to export as CSV:', error)
    }
  }, [data, exportFilename])

  const exportAsPDF = useCallback(async () => {
    if (!chartRef.current) return
    try {
      await exportUtils.exportAsPDF(chartRef.current, `${exportFilename}.pdf`)
    } catch (error) {
      console.error('Failed to export as PDF:', error)
    }
  }, [exportFilename])

  // Refresh function
  const refresh = useCallback(() => {
    onRefresh?.()
  }, [onRefresh])

  // Fullscreen functionality
  const toggleFullscreen = useCallback(() => {
    if (!enableFullscreen || !containerRef.current) return

    if (!isFullscreen) {
      if (containerRef.current.requestFullscreen) {
        containerRef.current.requestFullscreen()
        setIsFullscreen(true)
      }
    } else {
      if (document.exitFullscreen) {
        document.exitFullscreen()
        setIsFullscreen(false)
      }
    }
  }, [enableFullscreen, isFullscreen])

  // Handle fullscreen change events
  useEffect(() => {
    const handleFullscreenChange = () => {
      setIsFullscreen(!!document.fullscreenElement)
    }

    document.addEventListener('fullscreenchange', handleFullscreenChange)
    return () => document.removeEventListener('fullscreenchange', handleFullscreenChange)
  }, [])

  // Expose ref methods
  React.useImperativeHandle(ref, () => ({
    exportAsPNG,
    exportAsSVG,
    exportAsCSV,
    exportAsPDF,
    refresh,
    toggleFullscreen
  }), [exportAsPNG, exportAsSVG, exportAsCSV, exportAsPDF, refresh, toggleFullscreen])

  const errorMessage = error instanceof Error ? error.message : error

  return (
    <div
      ref={containerRef}
      className={cn(
        'relative rounded-lg border bg-card p-4 transition-all duration-300',
        'border-border bg-background',
        isFullscreen && 'fixed inset-0 z-50 rounded-none border-0 p-8',
        className
      )}
      style={{
        backgroundColor: chartTheme.colors.background,
        borderColor: chartTheme.colors.grid
      }}
    >
      {/* Header */}
      {(title || description || enableExport || enableFullscreen || onRefresh) && (
        <div className="mb-4 flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
          <div className="flex-1">
            {title && (
              <h3 
                className="text-lg font-semibold"
                style={{ color: chartTheme.colors.text }}
              >
                {title}
              </h3>
            )}
            {description && (
              <p 
                className="text-sm mt-1"
                style={{ color: chartTheme.colors.axis }}
              >
                {description}
              </p>
            )}
          </div>

          <div className="flex items-center gap-2">
            {/* Refresh button */}
            {onRefresh && (
              <Button
                variant="ghost"
                size="sm"
                onClick={refresh}
                disabled={loading}
                aria-label="Refresh chart"
              >
                <RefreshCw className={cn('h-4 w-4', loading && 'animate-spin')} />
              </Button>
            )}

            {/* Export menu */}
            {enableExport && (
              <div className="relative">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => setShowExportMenu(!showExportMenu)}
                  aria-label="Export chart"
                >
                  <Download className="h-4 w-4" />
                </Button>

                {showExportMenu && (
                  <>
                    {/* Backdrop */}
                    <div
                      className="fixed inset-0 z-10"
                      onClick={() => setShowExportMenu(false)}
                    />
                    
                    {/* Export menu */}
                    <div className="absolute right-0 top-full z-20 mt-2 w-48 rounded-md border bg-popover p-1 shadow-lg">
                      <button
                        className="flex w-full items-center px-3 py-2 text-sm hover:bg-accent rounded-sm"
                        onClick={() => {
                          exportAsPNG()
                          setShowExportMenu(false)
                        }}
                      >
                        Export as PNG
                      </button>
                      <button
                        className="flex w-full items-center px-3 py-2 text-sm hover:bg-accent rounded-sm"
                        onClick={() => {
                          exportAsSVG()
                          setShowExportMenu(false)
                        }}
                      >
                        Export as SVG
                      </button>
                      <button
                        className="flex w-full items-center px-3 py-2 text-sm hover:bg-accent rounded-sm"
                        onClick={() => {
                          exportAsPDF()
                          setShowExportMenu(false)
                        }}
                      >
                        Export as PDF
                      </button>
                      {data.length > 0 && (
                        <button
                          className="flex w-full items-center px-3 py-2 text-sm hover:bg-accent rounded-sm"
                          onClick={() => {
                            exportAsCSV()
                            setShowExportMenu(false)
                          }}
                        >
                          Export data as CSV
                        </button>
                      )}
                    </div>
                  </>
                )}
              </div>
            )}

            {/* Fullscreen button */}
            {enableFullscreen && (
              <Button
                variant="ghost"
                size="sm"
                onClick={toggleFullscreen}
                aria-label={isFullscreen ? "Exit fullscreen" : "Enter fullscreen"}
              >
                <Maximize2 className="h-4 w-4" />
              </Button>
            )}
          </div>
        </div>
      )}

      {/* Chart content */}
      <div
        ref={chartRef}
        className="relative"
        style={{
          height: typeof height === 'number' ? `${height}px` : height,
          minHeight: isFullscreen ? 'calc(100vh - 200px)' : undefined
        }}
      >
        {/* Loading state */}
        {loading && (
          <div className="absolute inset-0 flex items-center justify-center bg-background/80 backdrop-blur-sm">
            <div className="flex flex-col items-center gap-3">
              <Spinner size="lg" />
              <p className="text-sm text-muted-foreground">Loading chart data...</p>
            </div>
          </div>
        )}

        {/* Error state */}
        {errorMessage && !loading && (
          <div className="absolute inset-0 flex items-center justify-center p-4">
            <Alert variant="error" className="max-w-md">
              <div>
                <h4 className="font-medium">Failed to load chart</h4>
                <p className="mt-1 text-sm">{errorMessage}</p>
                {onRefresh && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={refresh}
                    className="mt-3"
                  >
                    <RefreshCw className="mr-2 h-4 w-4" />
                    Try again
                  </Button>
                )}
              </div>
            </Alert>
          </div>
        )}

        {/* Chart content */}
        {!loading && !errorMessage && (
          <div
            className="h-full w-full"
            style={{
              color: chartTheme.colors.text
            }}
          >
            {React.Children.map(children, child => {
              if (React.isValidElement(child)) {
                return React.cloneElement(child, {
                  theme: chartTheme,
                  dimensions: chartConfig.responsive ? dimensions : undefined,
                  animations: chartConfig.animations,
                  ...child.props
                } as any)
              }
              return child
            })}
          </div>
        )}
      </div>

      {/* Auto-refresh indicator */}
      {autoRefresh && !loading && (
        <div className="absolute bottom-2 right-2">
          <div className="flex items-center gap-2 rounded-full bg-muted px-3 py-1 text-xs text-muted-foreground">
            <div className="h-2 w-2 animate-pulse rounded-full bg-green-500" />
            Auto-refresh: {Math.floor(autoRefresh / 1000)}s
          </div>
        </div>
      )}
    </div>
  )
})

ChartContainer.displayName = 'ChartContainer'

export { ChartContainer }
export type { ChartContainerProps, ChartContainerRef }