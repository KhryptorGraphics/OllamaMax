/**
 * @fileoverview Accessibility Settings Panel
 * Provides user interface for configuring accessibility preferences
 */

import React, { useState } from 'react'
import { 
  Settings, 
  Eye, 
  Type, 
  Contrast, 
  Volume2, 
  Keyboard, 
  Focus,
  Palette,
  MousePointer2,
  Zap
} from 'lucide-react'
import { Button } from '@/design-system/components/Button/Button'
import { Card } from '@/design-system/components/Card/Card'
import { useAccessibilityContext } from './AccessibilityProvider'

interface AccessibilityPanelProps {
  /** Whether the panel is open */
  isOpen?: boolean
  /** Callback when panel open state changes */
  onOpenChange?: (open: boolean) => void
  /** Additional CSS classes */
  className?: string
}

/**
 * Accessibility Settings Panel Component
 * 
 * Provides a comprehensive interface for users to configure accessibility
 * preferences including motion, contrast, text size, and interaction modes.
 */
export const AccessibilityPanel: React.FC<AccessibilityPanelProps> = ({
  isOpen = false,
  onOpenChange,
  className
}) => {
  const { settings, updateSettings, announce } = useAccessibilityContext()
  const [isPanelExpanded, setIsPanelExpanded] = useState(isOpen)

  const handleTogglePanel = () => {
    const newState = !isPanelExpanded
    setIsPanelExpanded(newState)
    onOpenChange?.(newState)
    
    announce(
      newState ? 'Accessibility panel opened' : 'Accessibility panel closed',
      'polite'
    )
  }

  const handleSettingChange = (key: keyof typeof settings, value: any) => {
    updateSettings({ [key]: value })
    
    // Announce the change
    const settingNames = {
      reducedMotion: 'Reduced motion',
      highContrast: 'High contrast',
      largeText: 'Large text',
      darkMode: 'Dark mode',
      screenReaderOptimized: 'Screen reader optimization',
      keyboardEnhanced: 'Enhanced keyboard navigation',
      focusTrapping: 'Focus trapping'
    }
    
    const settingName = settingNames[key as keyof typeof settingNames] || key
    const state = value ? 'enabled' : 'disabled'
    announce(`${settingName} ${state}`, 'polite')
  }

  const handleAnnouncementChange = (key: keyof typeof settings.announcements, value: boolean) => {
    updateSettings({
      announcements: {
        ...settings.announcements,
        [key]: value
      }
    })
    
    const announcementNames = {
      navigation: 'Navigation announcements',
      loading: 'Loading announcements',
      errors: 'Error announcements',
      success: 'Success announcements'
    }
    
    const name = announcementNames[key] || key
    const state = value ? 'enabled' : 'disabled'
    announce(`${name} ${state}`, 'polite')
  }

  return (
    <div className={`accessibility-panel ${className || ''}`}>
      {/* Toggle Button */}
      <Button
        variant="outline"
        size="md"
        onClick={handleTogglePanel}
        aria-expanded={isPanelExpanded}
        aria-controls="accessibility-settings-panel"
        aria-label="Accessibility settings"
        className="accessibility-panel-toggle"
      >
        <Settings className="w-4 h-4 mr-2" aria-hidden="true" />
        Accessibility
      </Button>

      {/* Panel Content */}
      {isPanelExpanded && (
        <Card
          id="accessibility-settings-panel"
          className="accessibility-panel-content mt-2 p-4 w-80 shadow-lg"
          role="dialog"
          aria-labelledby="accessibility-panel-title"
          aria-describedby="accessibility-panel-description"
        >
          <div className="space-y-6">
            {/* Header */}
            <div>
              <h2 id="accessibility-panel-title" className="text-lg font-semibold flex items-center">
                <Settings className="w-5 h-5 mr-2" aria-hidden="true" />
                Accessibility Settings
              </h2>
              <p id="accessibility-panel-description" className="text-sm text-muted-foreground mt-1">
                Customize your experience for better accessibility
              </p>
            </div>

            {/* Visual Settings */}
            <div className="space-y-3">
              <h3 className="text-sm font-medium flex items-center">
                <Eye className="w-4 h-4 mr-2" aria-hidden="true" />
                Visual
              </h3>
              
              <div className="space-y-2 ml-6">
                <ToggleSetting
                  id="high-contrast"
                  icon={<Contrast className="w-4 h-4" />}
                  label="High contrast"
                  description="Increase color contrast for better visibility"
                  checked={settings.highContrast}
                  onChange={(checked) => handleSettingChange('highContrast', checked)}
                />
                
                <ToggleSetting
                  id="large-text"
                  icon={<Type className="w-4 h-4" />}
                  label="Large text"
                  description="Increase text size throughout the application"
                  checked={settings.largeText}
                  onChange={(checked) => handleSettingChange('largeText', checked)}
                />
                
                <ToggleSetting
                  id="dark-mode"
                  icon={<Palette className="w-4 h-4" />}
                  label="Dark mode"
                  description="Use dark theme to reduce eye strain"
                  checked={settings.darkMode}
                  onChange={(checked) => handleSettingChange('darkMode', checked)}
                />
              </div>
            </div>

            {/* Motion Settings */}
            <div className="space-y-3">
              <h3 className="text-sm font-medium flex items-center">
                <Zap className="w-4 h-4 mr-2" aria-hidden="true" />
                Motion
              </h3>
              
              <div className="space-y-2 ml-6">
                <ToggleSetting
                  id="reduced-motion"
                  icon={<MousePointer2 className="w-4 h-4" />}
                  label="Reduced motion"
                  description="Minimize animations and transitions"
                  checked={settings.reducedMotion}
                  onChange={(checked) => handleSettingChange('reducedMotion', checked)}
                />
              </div>
            </div>

            {/* Navigation Settings */}
            <div className="space-y-3">
              <h3 className="text-sm font-medium flex items-center">
                <Keyboard className="w-4 h-4 mr-2" aria-hidden="true" />
                Navigation
              </h3>
              
              <div className="space-y-2 ml-6">
                <ToggleSetting
                  id="keyboard-enhanced"
                  icon={<Keyboard className="w-4 h-4" />}
                  label="Enhanced keyboard navigation"
                  description="Improve keyboard navigation with additional shortcuts"
                  checked={settings.keyboardEnhanced}
                  onChange={(checked) => handleSettingChange('keyboardEnhanced', checked)}
                />
                
                <ToggleSetting
                  id="focus-trapping"
                  icon={<Focus className="w-4 h-4" />}
                  label="Focus trapping"
                  description="Keep focus within modal dialogs and popups"
                  checked={settings.focusTrapping}
                  onChange={(checked) => handleSettingChange('focusTrapping', checked)}
                />
              </div>
            </div>

            {/* Screen Reader Settings */}
            <div className="space-y-3">
              <h3 className="text-sm font-medium flex items-center">
                <Volume2 className="w-4 h-4 mr-2" aria-hidden="true" />
                Screen Reader
              </h3>
              
              <div className="space-y-2 ml-6">
                <ToggleSetting
                  id="screen-reader-optimized"
                  icon={<Volume2 className="w-4 h-4" />}
                  label="Screen reader optimization"
                  description="Optimize interface for screen reader users"
                  checked={settings.screenReaderOptimized}
                  onChange={(checked) => handleSettingChange('screenReaderOptimized', checked)}
                />
              </div>
            </div>

            {/* Announcement Settings */}
            <div className="space-y-3">
              <h3 className="text-sm font-medium">Announcements</h3>
              
              <div className="space-y-2 ml-6">
                <ToggleSetting
                  id="navigation-announcements"
                  label="Navigation changes"
                  description="Announce page and section changes"
                  checked={settings.announcements.navigation}
                  onChange={(checked) => handleAnnouncementChange('navigation', checked)}
                />
                
                <ToggleSetting
                  id="loading-announcements"
                  label="Loading states"
                  description="Announce loading and completion states"
                  checked={settings.announcements.loading}
                  onChange={(checked) => handleAnnouncementChange('loading', checked)}
                />
                
                <ToggleSetting
                  id="error-announcements"
                  label="Error messages"
                  description="Announce error messages immediately"
                  checked={settings.announcements.errors}
                  onChange={(checked) => handleAnnouncementChange('errors', checked)}
                />
                
                <ToggleSetting
                  id="success-announcements"
                  label="Success messages"
                  description="Announce successful operations"
                  checked={settings.announcements.success}
                  onChange={(checked) => handleAnnouncementChange('success', checked)}
                />
              </div>
            </div>

            {/* Reset Button */}
            <div className="pt-4 border-t">
              <Button
                variant="outline"
                size="sm"
                onClick={() => {
                  // Reset to defaults
                  updateSettings({
                    reducedMotion: false,
                    highContrast: false,
                    largeText: false,
                    darkMode: false,
                    screenReaderOptimized: false,
                    keyboardEnhanced: false,
                    focusTrapping: true,
                    announcements: {
                      navigation: true,
                      loading: true,
                      errors: true,
                      success: true
                    }
                  })
                  announce('Accessibility settings reset to defaults', 'polite')
                }}
                className="w-full"
              >
                Reset to defaults
              </Button>
            </div>
          </div>
        </Card>
      )}
    </div>
  )
}

/**
 * Toggle Setting Component
 */
interface ToggleSettingProps {
  id: string
  icon?: React.ReactNode
  label: string
  description?: string
  checked: boolean
  onChange: (checked: boolean) => void
}

const ToggleSetting: React.FC<ToggleSettingProps> = ({
  id,
  icon,
  label,
  description,
  checked,
  onChange
}) => {
  return (
    <div className="flex items-start space-x-3">
      <div className="flex-shrink-0 pt-0.5">
        {icon && (
          <span className="text-muted-foreground" aria-hidden="true">
            {icon}
          </span>
        )}
      </div>
      
      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between">
          <label
            htmlFor={id}
            className="text-sm font-medium cursor-pointer"
          >
            {label}
          </label>
          
          <input
            type="checkbox"
            id={id}
            checked={checked}
            onChange={(e) => onChange(e.target.checked)}
            className="w-4 h-4 text-primary-600 bg-background border-gray-300 rounded focus:ring-primary-500 focus:ring-2"
            aria-describedby={description ? `${id}-description` : undefined}
          />
        </div>
        
        {description && (
          <p
            id={`${id}-description`}
            className="text-xs text-muted-foreground mt-1"
          >
            {description}
          </p>
        )}
      </div>
    </div>
  )
}

export default AccessibilityPanel