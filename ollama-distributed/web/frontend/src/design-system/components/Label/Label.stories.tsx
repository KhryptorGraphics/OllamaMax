import type { Meta, StoryObj } from '@storybook/react'
import { Label, FieldsetLegend, FormDescription, FormError, FormSuccess, FormWarning } from './Label'
import { Input } from '../Input/Input'
import { Button } from '../Button/Button'
import { Card } from '../Card/Card'
import { Badge } from '../Badge/Badge'
import { useState } from 'react'

const meta: Meta<typeof Label> = {
  title: 'Design System/Label',
  component: Label,
  parameters: {
    docs: {
      description: {
        component: 'Label component with accessibility features and design token integration. Includes form descriptions, error messages, and validation states.'
      }
    }
  },
  argTypes: {
    variant: {
      control: 'select',
      options: ['default', 'muted', 'error', 'success', 'warning'],
      description: 'Visual variant of the label'
    },
    size: {
      control: 'select',
      options: ['sm', 'md', 'lg'],
      description: 'Size variant of the label'
    },
    required: {
      control: 'boolean',
      description: 'Whether the associated field is required'
    },
    hidden: {
      control: 'boolean',
      description: 'Whether to hide the label visually but keep it accessible'
    }
  },
  tags: ['autodocs']
}

export default meta
type Story = StoryObj<typeof Label>

// Basic labels
export const Default: Story = {
  args: {
    children: 'Email Address'
  }
}

export const Required: Story = {
  args: {
    children: 'Email Address',
    required: true
  }
}

export const Hidden: Story = {
  args: {
    children: 'Search',
    hidden: true
  },
  parameters: {
    docs: {
      description: {
        story: 'Hidden label is visually hidden but accessible to screen readers.'
      }
    }
  }
}

// Variant examples
export const Variants: Story = {
  render: () => (
    <div className="space-y-4">
      <div>
        <Label variant="default">Default Label</Label>
        <Input placeholder="Enter text..." className="mt-1" />
      </div>
      
      <div>
        <Label variant="muted">Muted Label</Label>
        <Input placeholder="Enter text..." className="mt-1" />
      </div>
      
      <div>
        <Label variant="error">Error Label</Label>
        <Input placeholder="Enter text..." className="mt-1 border-destructive" />
      </div>
      
      <div>
        <Label variant="success">Success Label</Label>
        <Input placeholder="Enter text..." className="mt-1 border-green-500" />
      </div>
      
      <div>
        <Label variant="warning">Warning Label</Label>
        <Input placeholder="Enter text..." className="mt-1 border-yellow-500" />
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Different visual variants for various states and contexts.'
      }
    }
  }
}

// Size variants
export const Sizes: Story = {
  render: () => (
    <div className="space-y-4">
      <div>
        <Label size="sm">Small Label</Label>
        <Input placeholder="Small input..." className="mt-1" />
      </div>
      
      <div>
        <Label size="md">Medium Label</Label>
        <Input placeholder="Medium input..." className="mt-1" />
      </div>
      
      <div>
        <Label size="lg">Large Label</Label>
        <Input placeholder="Large input..." className="mt-1" />
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Different size variants for various design scales.'
      }
    }
  }
}

// Form descriptions
export const WithDescriptions: Story = {
  render: () => (
    <div className="space-y-6">
      <div>
        <Label htmlFor="email-basic">Email Address</Label>
        <Input id="email-basic" placeholder="john@example.com" className="mt-1" />
        <FormDescription className="mt-1">
          We'll never share your email address with anyone else.
        </FormDescription>
      </div>
      
      <div>
        <Label htmlFor="password-basic" required>Password</Label>
        <Input 
          id="password-basic" 
          type="password" 
          placeholder="••••••••" 
          className="mt-1" 
        />
        <FormDescription className="mt-1">
          Must be at least 8 characters with mixed case letters, numbers, and symbols.
        </FormDescription>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Labels with helpful descriptions for user guidance.'
      }
    }
  }
}

// Error messages
export const WithErrors: Story = {
  render: () => (
    <div className="space-y-6">
      <div>
        <Label htmlFor="email-error" variant="error">Email Address</Label>
        <Input 
          id="email-error" 
          placeholder="john@example.com" 
          className="mt-1 border-destructive focus:ring-destructive" 
        />
        <FormError className="mt-1">
          Please enter a valid email address.
        </FormError>
      </div>
      
      <div>
        <Label htmlFor="password-error" variant="error" required>Password</Label>
        <Input 
          id="password-error" 
          type="password" 
          placeholder="••••••••" 
          className="mt-1 border-destructive focus:ring-destructive" 
        />
        <FormError className="mt-1" showIcon={false}>
          Password must be at least 8 characters long.
        </FormError>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Error states with appropriate styling and messaging.'
      }
    }
  }
}

// Success messages
export const WithSuccess: Story = {
  render: () => (
    <div className="space-y-6">
      <div>
        <Label htmlFor="email-success" variant="success">Email Address</Label>
        <Input 
          id="email-success" 
          placeholder="john@example.com" 
          value="john@example.com"
          className="mt-1 border-green-500 focus:ring-green-500" 
          readOnly
        />
        <FormSuccess className="mt-1">
          Email address verified successfully.
        </FormSuccess>
      </div>
      
      <div>
        <Label htmlFor="username-success" variant="success">Username</Label>
        <Input 
          id="username-success" 
          placeholder="username" 
          value="john_doe_2024"
          className="mt-1 border-green-500 focus:ring-green-500" 
          readOnly
        />
        <FormSuccess className="mt-1" showIcon={false}>
          Username is available!
        </FormSuccess>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Success states with confirmation messaging.'
      }
    }
  }
}

// Warning messages
export const WithWarnings: Story = {
  render: () => (
    <div className="space-y-6">
      <div>
        <Label htmlFor="password-warning" variant="warning">Password</Label>
        <Input 
          id="password-warning" 
          type="password" 
          placeholder="••••••••" 
          className="mt-1 border-yellow-500 focus:ring-yellow-500" 
        />
        <FormWarning className="mt-1">
          This password is commonly used. Consider a stronger password.
        </FormWarning>
      </div>
      
      <div>
        <Label htmlFor="domain-warning" variant="warning">Domain Name</Label>
        <Input 
          id="domain-warning" 
          placeholder="example.com" 
          className="mt-1 border-yellow-500 focus:ring-yellow-500" 
        />
        <FormWarning className="mt-1" showIcon={false}>
          This domain is already in use by another team.
        </FormWarning>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Warning states for potentially problematic input.'
      }
    }
  }
}

// Fieldset legends
export const FieldsetExamples: Story = {
  render: () => (
    <div className="space-y-8">
      <fieldset className="border border-border rounded-lg p-4">
        <FieldsetLegend required>Personal Information</FieldsetLegend>
        <div className="grid grid-cols-2 gap-4">
          <div>
            <Label htmlFor="first-name">First Name</Label>
            <Input id="first-name" placeholder="John" className="mt-1" />
          </div>
          <div>
            <Label htmlFor="last-name">Last Name</Label>
            <Input id="last-name" placeholder="Doe" className="mt-1" />
          </div>
        </div>
      </fieldset>
      
      <fieldset className="border border-border rounded-lg p-4">
        <FieldsetLegend variant="muted">Contact Preferences</FieldsetLegend>
        <div className="space-y-3">
          <div className="flex items-center space-x-2">
            <input type="checkbox" id="email-notifications" />
            <Label htmlFor="email-notifications">Email notifications</Label>
          </div>
          <div className="flex items-center space-x-2">
            <input type="checkbox" id="sms-notifications" />
            <Label htmlFor="sms-notifications">SMS notifications</Label>
          </div>
        </div>
      </fieldset>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Fieldset legends for grouping related form fields.'
      }
    }
  }
}

// Interactive validation form
export const InteractiveValidation: Story = {
  render: () => {
    const [formData, setFormData] = useState({
      email: '',
      password: '',
      confirmPassword: ''
    })
    const [errors, setErrors] = useState<Record<string, string>>({})
    const [touched, setTouched] = useState<Record<string, boolean>>({})

    const validateField = (name: string, value: string) => {
      const newErrors = { ...errors }
      
      switch (name) {
        case 'email':
          if (!value) {
            newErrors.email = 'Email is required'
          } else if (!/\S+@\S+\.\S+/.test(value)) {
            newErrors.email = 'Please enter a valid email address'
          } else {
            delete newErrors.email
          }
          break
        case 'password':
          if (!value) {
            newErrors.password = 'Password is required'
          } else if (value.length < 8) {
            newErrors.password = 'Password must be at least 8 characters'
          } else {
            delete newErrors.password
          }
          break
        case 'confirmPassword':
          if (!value) {
            newErrors.confirmPassword = 'Please confirm your password'
          } else if (value !== formData.password) {
            newErrors.confirmPassword = 'Passwords do not match'
          } else {
            delete newErrors.confirmPassword
          }
          break
      }
      
      setErrors(newErrors)
    }

    const handleChange = (name: string, value: string) => {
      setFormData(prev => ({ ...prev, [name]: value }))
      if (touched[name]) {
        validateField(name, value)
      }
    }

    const handleBlur = (name: string) => {
      setTouched(prev => ({ ...prev, [name]: true }))
      validateField(name, formData[name as keyof typeof formData])
    }

    const getFieldVariant = (name: string) => {
      if (!touched[name]) return 'default'
      return errors[name] ? 'error' : 'success'
    }

    return (
      <Card className="max-w-md">
        <Card.Header>
          <Card.Title>Create Account</Card.Title>
          <Card.Description>
            Fill out the form below to create your account.
          </Card.Description>
        </Card.Header>
        <Card.Content className="space-y-4">
          <div>
            <Label 
              htmlFor="signup-email" 
              variant={getFieldVariant('email')}
              required
            >
              Email Address
            </Label>
            <Input 
              id="signup-email"
              type="email"
              placeholder="john@example.com"
              value={formData.email}
              onChange={(e) => handleChange('email', e.target.value)}
              onBlur={() => handleBlur('email')}
              className={`mt-1 ${
                errors.email ? 'border-destructive focus:ring-destructive' :
                touched.email && !errors.email ? 'border-green-500 focus:ring-green-500' : ''
              }`}
            />
            {errors.email && touched.email && (
              <FormError className="mt-1">{errors.email}</FormError>
            )}
            {!errors.email && touched.email && formData.email && (
              <FormSuccess className="mt-1">Email looks good!</FormSuccess>
            )}
          </div>
          
          <div>
            <Label 
              htmlFor="signup-password" 
              variant={getFieldVariant('password')}
              required
            >
              Password
            </Label>
            <Input 
              id="signup-password"
              type="password"
              placeholder="••••••••"
              value={formData.password}
              onChange={(e) => handleChange('password', e.target.value)}
              onBlur={() => handleBlur('password')}
              className={`mt-1 ${
                errors.password ? 'border-destructive focus:ring-destructive' :
                touched.password && !errors.password ? 'border-green-500 focus:ring-green-500' : ''
              }`}
            />
            <FormDescription className="mt-1">
              Must be at least 8 characters long.
            </FormDescription>
            {errors.password && touched.password && (
              <FormError className="mt-1">{errors.password}</FormError>
            )}
          </div>
          
          <div>
            <Label 
              htmlFor="signup-confirm" 
              variant={getFieldVariant('confirmPassword')}
              required
            >
              Confirm Password
            </Label>
            <Input 
              id="signup-confirm"
              type="password"
              placeholder="••••••••"
              value={formData.confirmPassword}
              onChange={(e) => handleChange('confirmPassword', e.target.value)}
              onBlur={() => handleBlur('confirmPassword')}
              className={`mt-1 ${
                errors.confirmPassword ? 'border-destructive focus:ring-destructive' :
                touched.confirmPassword && !errors.confirmPassword ? 'border-green-500 focus:ring-green-500' : ''
              }`}
            />
            {errors.confirmPassword && touched.confirmPassword && (
              <FormError className="mt-1">{errors.confirmPassword}</FormError>
            )}
            {!errors.confirmPassword && touched.confirmPassword && formData.confirmPassword && (
              <FormSuccess className="mt-1">Passwords match!</FormSuccess>
            )}
          </div>
        </Card.Content>
        <Card.Footer>
          <Button 
            className="w-full" 
            disabled={Object.keys(errors).length > 0 || !formData.email || !formData.password}
          >
            Create Account
          </Button>
        </Card.Footer>
      </Card>
    )
  },
  parameters: {
    docs: {
      description: {
        story: 'Interactive form with real-time validation and dynamic label states.'
      }
    }
  }
}

// Complex form layout
export const ComplexForm: Story = {
  render: () => (
    <div className="max-w-2xl space-y-8">
      <fieldset className="border border-border rounded-lg p-6">
        <FieldsetLegend required size="lg">
          Project Details
        </FieldsetLegend>
        
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="md:col-span-2">
            <Label htmlFor="project-name" required>Project Name</Label>
            <Input 
              id="project-name" 
              placeholder="My Awesome Project" 
              className="mt-1" 
            />
            <FormDescription className="mt-1">
              Choose a descriptive name for your project.
            </FormDescription>
          </div>
          
          <div>
            <Label htmlFor="project-type">Project Type</Label>
            <select 
              id="project-type" 
              className="mt-1 w-full px-3 py-2 border border-border rounded-md bg-background"
            >
              <option value="">Select type...</option>
              <option value="web">Web Application</option>
              <option value="mobile">Mobile App</option>
              <option value="desktop">Desktop App</option>
            </select>
          </div>
          
          <div>
            <Label htmlFor="project-priority">Priority</Label>
            <select 
              id="project-priority" 
              className="mt-1 w-full px-3 py-2 border border-border rounded-md bg-background"
            >
              <option value="low">Low</option>
              <option value="medium">Medium</option>
              <option value="high">High</option>
              <option value="critical">Critical</option>
            </select>
          </div>
          
          <div className="md:col-span-2">
            <Label htmlFor="project-description">Description</Label>
            <textarea 
              id="project-description"
              placeholder="Describe your project..."
              rows={4}
              className="mt-1 w-full px-3 py-2 border border-border rounded-md bg-background resize-none"
            />
            <FormDescription className="mt-1">
              Provide a detailed description of your project goals and requirements.
            </FormDescription>
          </div>
        </div>
      </fieldset>
      
      <fieldset className="border border-border rounded-lg p-6">
        <FieldsetLegend variant="muted">
          Team & Collaboration
        </FieldsetLegend>
        
        <div className="space-y-4">
          <div>
            <Label htmlFor="team-lead">Team Lead</Label>
            <Input 
              id="team-lead" 
              placeholder="john@company.com" 
              className="mt-1" 
            />
          </div>
          
          <div>
            <Label>Team Members</Label>
            <div className="mt-2 space-y-2">
              <div className="flex items-center space-x-2">
                <input type="checkbox" id="member1" />
                <Label htmlFor="member1" size="sm">Sarah Wilson (Designer)</Label>
                <Badge variant="secondary" size="sm">Active</Badge>
              </div>
              <div className="flex items-center space-x-2">
                <input type="checkbox" id="member2" />
                <Label htmlFor="member2" size="sm">Mike Johnson (Developer)</Label>
                <Badge variant="success" size="sm">Available</Badge>
              </div>
              <div className="flex items-center space-x-2">
                <input type="checkbox" id="member3" />
                <Label htmlFor="member3" size="sm">Emma Taylor (QA)</Label>
                <Badge variant="warning" size="sm">Busy</Badge>
              </div>
            </div>
          </div>
        </div>
      </fieldset>
      
      <div className="flex space-x-4">
        <Button variant="outline" className="flex-1">
          Save as Draft
        </Button>
        <Button className="flex-1">
          Create Project
        </Button>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Complex form layout with fieldsets, various input types, and organized sections.'
      }
    }
  }
}

// Accessibility features
export const AccessibilityFeatures: Story = {
  render: () => (
    <div className="space-y-8">
      <div>
        <h3 className="text-sm font-medium mb-4">Proper Form Association</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Labels are properly associated with form controls using htmlFor attribute.
        </p>
        <div className="space-y-4">
          <div>
            <Label htmlFor="accessible-email">Email Address</Label>
            <Input 
              id="accessible-email" 
              type="email" 
              placeholder="john@example.com" 
              className="mt-1"
              aria-describedby="email-description"
            />
            <FormDescription id="email-description" className="mt-1">
              This description is linked to the input via aria-describedby.
            </FormDescription>
          </div>
        </div>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-4">Error Announcements</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Error messages use proper ARIA attributes for screen reader announcements.
        </p>
        <div>
          <Label htmlFor="error-demo" variant="error" required>
            Required Field
          </Label>
          <Input 
            id="error-demo" 
            placeholder="Enter value..." 
            className="mt-1 border-destructive"
            aria-invalid="true"
            aria-describedby="error-message"
          />
          <FormError id="error-message" className="mt-1">
            This field is required and cannot be empty.
          </FormError>
        </div>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-4">High Contrast Support</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Labels maintain proper contrast ratios in high contrast mode.
        </p>
        <div className="high-contrast p-4 rounded-lg border space-y-3">
          <div>
            <Label htmlFor="hc-input">High Contrast Label</Label>
            <Input id="hc-input" placeholder="Text input..." className="mt-1" />
          </div>
          <div>
            <Label htmlFor="hc-error" variant="error">Error Label</Label>
            <Input id="hc-error" placeholder="Error input..." className="mt-1" />
            <FormError className="mt-1">Error message in high contrast</FormError>
          </div>
        </div>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-4">Screen Reader Support</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Hidden labels remain accessible to screen readers.
        </p>
        <div className="space-y-3">
          <div className="flex items-center space-x-2">
            <Label htmlFor="search-visible">Visible Search Label:</Label>
            <Input id="search-visible" placeholder="Search..." className="flex-1" />
          </div>
          <div className="flex items-center space-x-2">
            <Label htmlFor="search-hidden" hidden>Hidden Search Label</Label>
            <span className="text-sm">Hidden label:</span>
            <Input id="search-hidden" placeholder="Search with hidden label..." className="flex-1" />
          </div>
        </div>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Demonstrates accessibility features including proper form association, error announcements, and screen reader support.'
      }
    }
  }
}

// Playground
export const Playground: Story = {
  args: {
    children: 'Label Text',
    variant: 'default',
    size: 'md',
    required: false,
    hidden: false
  },
  parameters: {
    docs: {
      description: {
        story: 'Interactive playground to test different label configurations.'
      }
    }
  }
}