import type { Meta, StoryObj } from '@storybook/react'
import { Input, Textarea } from './Input'
import { Search, Mail, Lock, User, Phone, Calendar } from 'lucide-react'
import { useState } from 'react'

const meta: Meta<typeof Input> = {
  title: 'Design System/Input',
  component: Input,
  parameters: {
    docs: {
      description: {
        component: 'A comprehensive input component with validation states, icons, and accessibility features. Includes password visibility toggle and various input types support.'
      }
    }
  },
  argTypes: {
    variant: {
      control: 'select',
      options: ['default', 'error', 'success', 'warning'],
      description: 'Visual state of the input'
    },
    size: {
      control: 'select',
      options: ['sm', 'md', 'lg'],
      description: 'Size of the input'
    },
    type: {
      control: 'select',
      options: ['text', 'email', 'password', 'number', 'tel', 'url', 'search'],
      description: 'HTML input type'
    },
    disabled: {
      control: 'boolean',
      description: 'Disable the input'
    },
    required: {
      control: 'boolean',
      description: 'Mark the field as required'
    },
    label: {
      control: 'text',
      description: 'Label text for the input'
    },
    placeholder: {
      control: 'text',
      description: 'Placeholder text'
    },
    helperText: {
      control: 'text',
      description: 'Helper text displayed below the input'
    },
    error: {
      control: 'text',
      description: 'Error message'
    },
    success: {
      control: 'text',
      description: 'Success message'
    },
    warning: {
      control: 'text',
      description: 'Warning message'
    }
  },
  tags: ['autodocs']
}

export default meta
type Story = StoryObj<typeof Input>

// Default story
export const Default: Story = {
  args: {
    placeholder: 'Enter text...'
  }
}

// With label
export const WithLabel: Story = {
  args: {
    label: 'Email Address',
    type: 'email',
    placeholder: 'you@example.com'
  }
}

// Required field
export const Required: Story = {
  args: {
    label: 'Full Name',
    placeholder: 'Enter your full name',
    required: true,
    helperText: 'This field is required'
  }
}

// Size variations
export const Sizes: Story = {
  render: () => (
    <div className="space-y-4 max-w-md">
      <Input 
        label="Small Input" 
        size="sm" 
        placeholder="Small size input" 
      />
      <Input 
        label="Medium Input" 
        size="md" 
        placeholder="Medium size input" 
      />
      <Input 
        label="Large Input" 
        size="lg" 
        placeholder="Large size input" 
      />
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Available input sizes: small, medium, and large.'
      }
    }
  }
}

// Validation states
export const ValidationStates: Story = {
  render: () => (
    <div className="space-y-4 max-w-md">
      <Input 
        label="Default State" 
        placeholder="Normal input" 
        helperText="This is a normal input field"
      />
      <Input 
        label="Success State" 
        placeholder="Valid input" 
        success="Great! This looks correct."
      />
      <Input 
        label="Warning State" 
        placeholder="Warning input" 
        warning="Please double-check this information."
      />
      <Input 
        label="Error State" 
        placeholder="Invalid input" 
        error="This field is required."
      />
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Different validation states with appropriate styling and messaging.'
      }
    }
  }
}

// With icons
export const WithIcons: Story = {
  render: () => (
    <div className="space-y-4 max-w-md">
      <Input 
        label="Search" 
        placeholder="Search for something..." 
        leftIcon={<Search />}
        type="search"
      />
      <Input 
        label="Email" 
        placeholder="you@example.com" 
        leftIcon={<Mail />}
        type="email"
      />
      <Input 
        label="Username" 
        placeholder="Enter username" 
        leftIcon={<User />}
        rightIcon={<Calendar />}
      />
      <Input 
        label="Phone Number" 
        placeholder="+1 (555) 123-4567" 
        leftIcon={<Phone />}
        type="tel"
      />
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Inputs with left and right icons for better visual context.'
      }
    }
  }
}

// Password input
export const PasswordInput: Story = {
  render: () => (
    <div className="space-y-4 max-w-md">
      <Input 
        label="Password" 
        type="password" 
        placeholder="Enter your password"
        leftIcon={<Lock />}
        helperText="Password must be at least 8 characters"
      />
      <Input 
        label="Confirm Password" 
        type="password" 
        placeholder="Confirm your password"
        leftIcon={<Lock />}
        success="Passwords match!"
      />
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Password inputs with automatic toggle visibility button.'
      }
    }
  }
}

// Disabled state
export const Disabled: Story = {
  render: () => (
    <div className="space-y-4 max-w-md">
      <Input 
        label="Disabled Input" 
        placeholder="This input is disabled" 
        disabled
        helperText="This field cannot be edited"
      />
      <Input 
        label="Disabled with Value" 
        value="Read-only value" 
        disabled
        leftIcon={<User />}
      />
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Disabled inputs with appropriate visual feedback.'
      }
    }
  }
}

// Form example
export const FormExample: Story = {
  render: () => {
    const [formData, setFormData] = useState({
      firstName: '',
      lastName: '',
      email: '',
      password: '',
      confirmPassword: ''
    })

    const [errors, setErrors] = useState<Record<string, string>>({})

    const handleSubmit = (e: React.FormEvent) => {
      e.preventDefault()
      
      const newErrors: Record<string, string> = {}
      
      if (!formData.firstName) newErrors.firstName = 'First name is required'
      if (!formData.lastName) newErrors.lastName = 'Last name is required'
      if (!formData.email) newErrors.email = 'Email is required'
      if (!formData.password) newErrors.password = 'Password is required'
      if (formData.password !== formData.confirmPassword) {
        newErrors.confirmPassword = 'Passwords do not match'
      }
      
      setErrors(newErrors)
      
      if (Object.keys(newErrors).length === 0) {
        alert('Form submitted successfully!')
      }
    }

    return (
      <form onSubmit={handleSubmit} className="space-y-4 max-w-md">
        <div className="grid grid-cols-2 gap-4">
          <Input
            label="First Name"
            placeholder="John"
            value={formData.firstName}
            onChange={(e) => setFormData(prev => ({ ...prev, firstName: e.target.value }))}
            error={errors.firstName}
            required
          />
          <Input
            label="Last Name"
            placeholder="Doe"
            value={formData.lastName}
            onChange={(e) => setFormData(prev => ({ ...prev, lastName: e.target.value }))}
            error={errors.lastName}
            required
          />
        </div>
        
        <Input
          label="Email Address"
          type="email"
          placeholder="john@example.com"
          leftIcon={<Mail />}
          value={formData.email}
          onChange={(e) => setFormData(prev => ({ ...prev, email: e.target.value }))}
          error={errors.email}
          required
        />
        
        <Input
          label="Password"
          type="password"
          placeholder="Enter password"
          leftIcon={<Lock />}
          value={formData.password}
          onChange={(e) => setFormData(prev => ({ ...prev, password: e.target.value }))}
          error={errors.password}
          required
          helperText="Must be at least 8 characters"
        />
        
        <Input
          label="Confirm Password"
          type="password"
          placeholder="Confirm password"
          leftIcon={<Lock />}
          value={formData.confirmPassword}
          onChange={(e) => setFormData(prev => ({ ...prev, confirmPassword: e.target.value }))}
          error={errors.confirmPassword}
          required
        />
        
        <button 
          type="submit"
          className="w-full mt-6 px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 transition-colors"
        >
          Create Account
        </button>
      </form>
    )
  },
  parameters: {
    docs: {
      description: {
        story: 'Complete form example with validation and error handling.'
      }
    }
  }
}

// Textarea examples
export const TextareaExample: Story = {
  render: () => (
    <div className="space-y-4 max-w-md">
      <Textarea
        label="Message"
        placeholder="Enter your message..."
        helperText="Maximum 500 characters"
      />
      
      <Textarea
        label="Description"
        placeholder="Describe the issue..."
        size="lg"
        error="Description is required"
      />
      
      <Textarea
        label="Comments"
        placeholder="Add your comments..."
        success="Comments saved successfully!"
        autoResize
      />
      
      <Textarea
        label="Notes"
        placeholder="Additional notes..."
        disabled
        value="This is a read-only note that cannot be edited by the user."
      />
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Textarea component with similar styling and validation features as Input.'
      }
    }
  }
}

// Accessibility demonstration
export const AccessibilityDemo: Story = {
  render: () => (
    <div className="space-y-6 max-w-md">
      <div>
        <h3 className="text-sm font-medium mb-3">Keyboard Navigation</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Use Tab to navigate between fields, and Shift+Tab to go backwards.
        </p>
        <div className="space-y-3">
          <Input label="First Field" placeholder="Tab here first" />
          <Input label="Second Field" placeholder="Then tab here" />
          <Input label="Third Field" placeholder="Finally here" />
        </div>
      </div>
      
      <div>
        <h3 className="text-sm font-medium mb-3">ARIA Attributes</h3>
        <p className="text-sm text-muted-foreground mb-4">
          Inputs have proper ARIA labels, descriptions, and invalid states.
        </p>
        <div className="space-y-3">
          <Input 
            label="Required Field" 
            placeholder="This field is required"
            required
            helperText="Screen readers will announce this as required"
          />
          <Input 
            label="Invalid Field" 
            placeholder="This field has an error"
            error="Screen readers will announce this error"
          />
        </div>
      </div>
    </div>
  ),
  parameters: {
    docs: {
      description: {
        story: 'Accessibility features including proper ARIA attributes and keyboard navigation.'
      }
    }
  }
}

// Interactive playground
export const Playground: Story = {
  args: {
    label: 'Interactive Input',
    placeholder: 'Try different configurations...',
    variant: 'default',
    size: 'md',
    type: 'text',
    disabled: false,
    required: false,
    helperText: 'This is helper text'
  },
  parameters: {
    docs: {
      description: {
        story: 'Interactive playground to test different input configurations.'
      }
    }
  }
}