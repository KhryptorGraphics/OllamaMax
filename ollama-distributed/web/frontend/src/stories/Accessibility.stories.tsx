import { Meta } from '@storybook/react'
import { Button } from '../design-system/components/Button/Button'
import { Input } from '../design-system/components/Input/Input'
import { Alert } from '../design-system/components/Alert/Alert'
import { Card } from '../design-system/components/Card/Card'
import { Badge } from '../design-system/components/Badge/Badge'
import { useState } from 'react'

export default {
  title: 'Design System/Accessibility',
  parameters: {
    docs: {
      page: () => (
        <div className="p-6">
          <h1 className="text-3xl font-bold mb-6">Accessibility Guidelines</h1>
          <p className="text-lg text-muted-foreground mb-8">
            Our design system is built with accessibility as a core principle. All components 
            meet WCAG 2.1 AA standards and provide comprehensive support for assistive technologies.
          </p>

          <div className="space-y-12">
            {/* Keyboard Navigation */}
            <section>
              <h2 className="text-2xl font-semibold mb-6">Keyboard Navigation</h2>
              <p className="text-muted-foreground mb-6">
                All interactive elements support keyboard navigation using standard patterns.
              </p>
              
              <div className="space-y-6">
                <div className="p-6 border rounded-lg">
                  <h3 className="text-lg font-medium mb-4">Navigation Keys</h3>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <kbd className="px-2 py-1 bg-muted rounded text-sm">Tab</kbd>
                      <p className="text-sm text-muted-foreground">Move to next focusable element</p>
                    </div>
                    <div className="space-y-2">
                      <kbd className="px-2 py-1 bg-muted rounded text-sm">Shift + Tab</kbd>
                      <p className="text-sm text-muted-foreground">Move to previous focusable element</p>
                    </div>
                    <div className="space-y-2">
                      <kbd className="px-2 py-1 bg-muted rounded text-sm">Enter</kbd>
                      <p className="text-sm text-muted-foreground">Activate buttons and links</p>
                    </div>
                    <div className="space-y-2">
                      <kbd className="px-2 py-1 bg-muted rounded text-sm">Space</kbd>
                      <p className="text-sm text-muted-foreground">Activate buttons and toggle states</p>
                    </div>
                    <div className="space-y-2">
                      <kbd className="px-2 py-1 bg-muted rounded text-sm">Escape</kbd>
                      <p className="text-sm text-muted-foreground">Close modals and dropdowns</p>
                    </div>
                    <div className="space-y-2">
                      <kbd className="px-2 py-1 bg-muted rounded text-sm">Arrow Keys</kbd>
                      <p className="text-sm text-muted-foreground">Navigate within components</p>
                    </div>
                  </div>
                </div>
                
                <div className="p-6 border rounded-lg">
                  <h3 className="text-lg font-medium mb-4">Focus Management</h3>
                  <ul className="space-y-2 text-muted-foreground">
                    <li>• Focus indicators are clearly visible with high contrast</li>
                    <li>• Focus order follows logical reading sequence</li>
                    <li>• Focus is trapped within modal dialogs</li>
                    <li>• Focus returns to trigger element when dialogs close</li>
                    <li>• Skip links allow navigation to main content</li>
                  </ul>
                </div>
              </div>
            </section>

            {/* Screen Reader Support */}
            <section>
              <h2 className="text-2xl font-semibold mb-6">Screen Reader Support</h2>
              
              <div className="space-y-6">
                <div className="p-6 border rounded-lg">
                  <h3 className="text-lg font-medium mb-4">ARIA Attributes</h3>
                  <div className="space-y-4">
                    <div>
                      <h4 className="font-medium mb-2">aria-label</h4>
                      <p className="text-sm text-muted-foreground mb-2">Provides accessible names for elements</p>
                      <code className="text-xs bg-muted px-2 py-1 rounded">
                        &lt;button aria-label="Close dialog"&gt;×&lt;/button&gt;
                      </code>
                    </div>
                    
                    <div>
                      <h4 className="font-medium mb-2">aria-describedby</h4>
                      <p className="text-sm text-muted-foreground mb-2">Links elements to their descriptions</p>
                      <code className="text-xs bg-muted px-2 py-1 rounded">
                        &lt;input aria-describedby="password-help" /&gt;
                      </code>
                    </div>
                    
                    <div>
                      <h4 className="font-medium mb-2">aria-invalid</h4>
                      <p className="text-sm text-muted-foreground mb-2">Indicates form validation state</p>
                      <code className="text-xs bg-muted px-2 py-1 rounded">
                        &lt;input aria-invalid="true" /&gt;
                      </code>
                    </div>
                    
                    <div>
                      <h4 className="font-medium mb-2">role</h4>
                      <p className="text-sm text-muted-foreground mb-2">Defines element purpose and behavior</p>
                      <code className="text-xs bg-muted px-2 py-1 rounded">
                        &lt;div role="alert"&gt;Error message&lt;/div&gt;
                      </code>
                    </div>
                  </div>
                </div>

                <div className="p-6 border rounded-lg">
                  <h3 className="text-lg font-medium mb-4">Live Regions</h3>
                  <ul className="space-y-2 text-muted-foreground">
                    <li>• Status updates are announced automatically</li>
                    <li>• Error messages use role="alert" for immediate announcement</li>
                    <li>• Progress updates use aria-live="polite"</li>
                    <li>• Loading states are communicated clearly</li>
                  </ul>
                </div>
              </div>
            </section>

            {/* Color and Contrast */}
            <section>
              <h2 className="text-2xl font-semibold mb-6">Color and Contrast</h2>
              
              <div className="space-y-6">
                <div className="p-6 border rounded-lg">
                  <h3 className="text-lg font-medium mb-4">Contrast Requirements</h3>
                  <div className="space-y-3">
                    <div className="flex items-center justify-between p-3 border rounded">
                      <span>Normal text (14px+)</span>
                      <span className="font-mono text-sm">4.5:1 minimum</span>
                    </div>
                    <div className="flex items-center justify-between p-3 border rounded">
                      <span>Large text (18px+ or 14px+ bold)</span>
                      <span className="font-mono text-sm">3:1 minimum</span>
                    </div>
                    <div className="flex items-center justify-between p-3 border rounded">
                      <span>UI components and graphics</span>
                      <span className="font-mono text-sm">3:1 minimum</span>
                    </div>
                  </div>
                </div>

                <div className="p-6 border rounded-lg">
                  <h3 className="text-lg font-medium mb-4">Color Usage</h3>
                  <ul className="space-y-2 text-muted-foreground">
                    <li>• Color is never the only way to convey information</li>
                    <li>• Error states include icons and text labels</li>
                    <li>• Status indicators use multiple visual cues</li>
                    <li>• Links are distinguishable by more than color</li>
                    <li>• Dark mode maintains contrast requirements</li>
                  </ul>
                </div>
              </div>
            </section>

            {/* Form Accessibility */}
            <section>
              <h2 className="text-2xl font-semibold mb-6">Form Accessibility</h2>
              
              <div className="space-y-6">
                <div className="p-6 border rounded-lg">
                  <h3 className="text-lg font-medium mb-4">Form Structure</h3>
                  <ul className="space-y-2 text-muted-foreground">
                    <li>• All form controls have associated labels</li>
                    <li>• Required fields are clearly marked</li>
                    <li>• Error messages are linked to form controls</li>
                    <li>• Fieldsets group related form controls</li>
                    <li>• Form submission provides clear feedback</li>
                  </ul>
                </div>

                <div className="p-6 border rounded-lg">
                  <h3 className="text-lg font-medium mb-4">Error Handling</h3>
                  <ul className="space-y-2 text-muted-foreground">
                    <li>• Error messages are specific and actionable</li>
                    <li>• Errors are announced to screen readers</li>
                    <li>• Invalid fields maintain focus for correction</li>
                    <li>• Success states provide positive confirmation</li>
                    <li>• Progressive enhancement prevents data loss</li>
                  </ul>
                </div>
              </div>
            </section>

            {/* Motion and Animation */}
            <section>
              <h2 className="text-2xl font-semibold mb-6">Motion and Animation</h2>
              
              <div className="space-y-6">
                <div className="p-6 border rounded-lg">
                  <h3 className="text-lg font-medium mb-4">Reduced Motion</h3>
                  <ul className="space-y-2 text-muted-foreground">
                    <li>• Respects prefers-reduced-motion setting</li>
                    <li>• Essential animations remain functional</li>
                    <li>• Decorative animations are disabled</li>
                    <li>• Parallax effects are reduced or removed</li>
                    <li>• Auto-playing content can be paused</li>
                  </ul>
                </div>

                <div className="p-6 border rounded-lg">
                  <h3 className="text-lg font-medium mb-4">Animation Guidelines</h3>
                  <ul className="space-y-2 text-muted-foreground">
                    <li>• Animations have clear purpose and meaning</li>
                    <li>• Duration is appropriate for the interaction</li>
                    <li>• Motion follows natural physics</li>
                    <li>• Flashing content is avoided</li>
                    <li>• Loading animations provide clear feedback</li>
                  </ul>
                </div>
              </div>
            </section>

            {/* Testing Checklist */}
            <section>
              <h2 className="text-2xl font-semibold mb-6">Accessibility Testing Checklist</h2>
              
              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="p-6 border rounded-lg">
                  <h3 className="text-lg font-medium mb-4">Manual Testing</h3>
                  <ul className="space-y-2 text-sm">
                    <li className="flex items-start">
                      <input type="checkbox" className="mt-1 mr-2" />
                      Navigate using only keyboard
                    </li>
                    <li className="flex items-start">
                      <input type="checkbox" className="mt-1 mr-2" />
                      Test with screen reader
                    </li>
                    <li className="flex items-start">
                      <input type="checkbox" className="mt-1 mr-2" />
                      Verify color contrast
                    </li>
                    <li className="flex items-start">
                      <input type="checkbox" className="mt-1 mr-2" />
                      Check focus indicators
                    </li>
                    <li className="flex items-start">
                      <input type="checkbox" className="mt-1 mr-2" />
                      Test at 200% zoom
                    </li>
                    <li className="flex items-start">
                      <input type="checkbox" className="mt-1 mr-2" />
                      Verify form validation
                    </li>
                  </ul>
                </div>

                <div className="p-6 border rounded-lg">
                  <h3 className="text-lg font-medium mb-4">Automated Testing</h3>
                  <ul className="space-y-2 text-sm">
                    <li className="flex items-start">
                      <input type="checkbox" className="mt-1 mr-2" />
                      Run axe-core accessibility tests
                    </li>
                    <li className="flex items-start">
                      <input type="checkbox" className="mt-1 mr-2" />
                      Validate HTML semantics
                    </li>
                    <li className="flex items-start">
                      <input type="checkbox" className="mt-1 mr-2" />
                      Check ARIA attributes
                    </li>
                    <li className="flex items-start">
                      <input type="checkbox" className="mt-1 mr-2" />
                      Verify heading structure
                    </li>
                    <li className="flex items-start">
                      <input type="checkbox" className="mt-1 mr-2" />
                      Test color contrast ratios
                    </li>
                    <li className="flex items-start">
                      <input type="checkbox" className="mt-1 mr-2" />
                      Validate form labels
                    </li>
                  </ul>
                </div>
              </div>
            </section>
          </div>
        </div>
      )
    }
  }
} as Meta

export const KeyboardNavigation = () => {
  return (
    <div className="space-y-6 p-6">
      <h2 className="text-xl font-semibold">Keyboard Navigation Demo</h2>
      <p className="text-muted-foreground">
        Use Tab to navigate between elements, Enter or Space to activate.
      </p>
      
      <div className="space-y-4">
        <div className="flex gap-2">
          <Button>First Button</Button>
          <Button variant="secondary">Second Button</Button>
          <Button variant="outline">Third Button</Button>
        </div>
        
        <div className="max-w-md">
          <Input label="Email" type="email" placeholder="test@example.com" />
        </div>
        
        <div className="flex gap-2">
          <Badge interactive>Interactive Badge</Badge>
          <Badge removable onRemove={() => alert('Badge removed')}>
            Removable Badge
          </Badge>
        </div>
      </div>
    </div>
  )
}

export const ScreenReaderDemo = () => {
  const [showAlert, setShowAlert] = useState(false)
  
  return (
    <div className="space-y-6 p-6">
      <h2 className="text-xl font-semibold">Screen Reader Support Demo</h2>
      <p className="text-muted-foreground">
        These elements are properly announced by screen readers.
      </p>
      
      <div className="space-y-4">
        <Button onClick={() => setShowAlert(true)}>
          Trigger Alert (Screen Reader Test)
        </Button>
        
        {showAlert && (
          <Alert 
            variant="info" 
            title="Screen Reader Alert"
            dismissible
            onDismiss={() => setShowAlert(false)}
          >
            This alert will be announced immediately by screen readers using role="alert".
          </Alert>
        )}
        
        <Card>
          <Card.Header>
            <Card.Title level={3}>Accessible Card</Card.Title>
            <Card.Description>
              This card uses proper heading hierarchy and semantic structure.
            </Card.Description>
          </Card.Header>
          <Card.Content>
            <Input 
              label="Username" 
              helperText="This help text is linked to the input for screen readers"
              aria-describedby="username-help"
            />
          </Card.Content>
        </Card>
      </div>
    </div>
  )
}

export const FormAccessibility = () => {
  const [formData, setFormData] = useState({
    name: '',
    email: '',
    message: ''
  })
  const [errors, setErrors] = useState<Record<string, string>>({})
  
  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    
    const newErrors: Record<string, string> = {}
    if (!formData.name) newErrors.name = 'Name is required'
    if (!formData.email) newErrors.email = 'Email is required'
    if (!formData.message) newErrors.message = 'Message is required'
    
    setErrors(newErrors)
  }
  
  return (
    <div className="space-y-6 p-6">
      <h2 className="text-xl font-semibold">Accessible Form Demo</h2>
      <p className="text-muted-foreground">
        This form demonstrates proper labeling, error handling, and validation feedback.
      </p>
      
      <form onSubmit={handleSubmit} className="space-y-4 max-w-md">
        <Input
          label="Full Name"
          value={formData.name}
          onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
          error={errors.name}
          required
          aria-describedby="name-help"
          helperText="Enter your first and last name"
        />
        
        <Input
          label="Email Address"
          type="email"
          value={formData.email}
          onChange={(e) => setFormData(prev => ({ ...prev, email: e.target.value }))}
          error={errors.email}
          required
          aria-describedby="email-help"
          helperText="We'll never share your email"
        />
        
        <div className="space-y-2">
          <label htmlFor="message" className="text-sm font-medium">
            Message *
          </label>
          <textarea
            id="message"
            className="w-full p-3 border rounded-md"
            rows={4}
            value={formData.message}
            onChange={(e) => setFormData(prev => ({ ...prev, message: e.target.value }))}
            aria-invalid={!!errors.message}
            aria-describedby={errors.message ? "message-error" : "message-help"}
            required
          />
          {errors.message ? (
            <p id="message-error" className="text-sm text-destructive" role="alert">
              {errors.message}
            </p>
          ) : (
            <p id="message-help" className="text-sm text-muted-foreground">
              Tell us how we can help you
            </p>
          )}
        </div>
        
        <Button type="submit">Submit Form</Button>
      </form>
    </div>
  )
}

export const ColorContrastDemo = () => {
  return (
    <div className="space-y-6 p-6">
      <h2 className="text-xl font-semibold">Color Contrast Examples</h2>
      <p className="text-muted-foreground">
        All color combinations meet WCAG 2.1 AA contrast requirements.
      </p>
      
      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        <div className="space-y-4">
          <h3 className="font-medium">Text on Backgrounds</h3>
          <div className="space-y-2">
            <div className="p-4 bg-primary text-primary-foreground rounded">
              Primary background with foreground text (4.5:1)
            </div>
            <div className="p-4 bg-secondary text-secondary-foreground rounded">
              Secondary background with foreground text (4.5:1)
            </div>
            <div className="p-4 bg-muted text-muted-foreground rounded">
              Muted background with foreground text (4.5:1)
            </div>
          </div>
        </div>
        
        <div className="space-y-4">
          <h3 className="font-medium">Status Colors</h3>
          <div className="space-y-2">
            <Alert variant="success" title="Success">
              Success message with proper contrast
            </Alert>
            <Alert variant="warning" title="Warning">
              Warning message with proper contrast
            </Alert>
            <Alert variant="destructive" title="Error">
              Error message with proper contrast
            </Alert>
          </div>
        </div>
      </div>
    </div>
  )
}