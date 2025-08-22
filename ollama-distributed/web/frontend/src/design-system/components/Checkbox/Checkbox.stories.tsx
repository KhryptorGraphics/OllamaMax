import type { Meta, StoryObj } from '@storybook/react'
import { useState } from 'react'
import { Checkbox, CheckboxGroup } from './Checkbox'
import { Heart, Star, Settings, Shield, Bell, Lock } from 'lucide-react'

const meta = {
  title: 'Design System/Forms/Checkbox',
  component: Checkbox,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: `
The Checkbox component provides a flexible and accessible way to create checkboxes with various states and styles.

## Features
- **States**: Checked, unchecked, and indeterminate states
- **Variants**: Default, error, success, and warning styles
- **Sizes**: Small, medium, and large sizes
- **Custom Icons**: Support for custom check and indeterminate icons
- **Custom Colors**: Ability to customize checked state color
- **Accessibility**: Full ARIA support and keyboard navigation
- **Group Layouts**: Horizontal, vertical, and grid layouts for checkbox groups
- **Select All**: Automatic select all functionality in groups

## Accessibility
- Proper labeling with \`for\` attribute connection
- \`aria-invalid\` for error states
- \`aria-describedby\` for descriptions and error messages
- Keyboard navigation support
- Screen reader friendly

## Usage
\`\`\`tsx
import { Checkbox, CheckboxGroup } from '@/design-system/components/Checkbox'

// Basic checkbox
<Checkbox label="Accept terms" />

// With custom icon
<Checkbox 
  label="Favorite" 
  checkedIcon={<Heart className="h-3 w-3" />}
/>

// Checkbox group
<CheckboxGroup
  label="Select features"
  options={[
    { value: 'feature1', label: 'Feature 1' },
    { value: 'feature2', label: 'Feature 2' }
  ]}
/>
\`\`\`
        `
      }
    }
  },
  tags: ['autodocs'],
  argTypes: {
    variant: {
      control: 'select',
      options: ['default', 'error', 'success', 'warning']
    },
    size: {
      control: 'select',
      options: ['sm', 'md', 'lg']
    },
    label: {
      control: 'text'
    },
    description: {
      control: 'text'
    },
    error: {
      control: 'text'
    },
    disabled: {
      control: 'boolean'
    },
    indeterminate: {
      control: 'boolean'
    },
    checkedColor: {
      control: 'color'
    }
  }
} satisfies Meta<typeof Checkbox>

export default meta
type Story = StoryObj<typeof meta>

// Basic checkbox
export const Default: Story = {
  args: {
    label: 'Accept terms and conditions'
  }
}

// Different sizes
export const Sizes: Story = {
  render: () => (
    <div className="flex flex-col gap-4">
      <Checkbox size="sm" label="Small checkbox" />
      <Checkbox size="md" label="Medium checkbox (default)" defaultChecked />
      <Checkbox size="lg" label="Large checkbox" />
    </div>
  )
}

// Different variants
export const Variants: Story = {
  render: () => (
    <div className="flex flex-col gap-4">
      <Checkbox variant="default" label="Default variant" defaultChecked />
      <Checkbox variant="error" label="Error variant" defaultChecked />
      <Checkbox variant="success" label="Success variant" defaultChecked />
      <Checkbox variant="warning" label="Warning variant" defaultChecked />
    </div>
  )
}

// With descriptions
export const WithDescriptions: Story = {
  render: () => (
    <div className="flex flex-col gap-4">
      <Checkbox 
        label="Enable notifications"
        description="Receive email updates about your account"
      />
      <Checkbox 
        label="Marketing emails"
        description="Get updates about new features and promotions"
        defaultChecked
      />
    </div>
  )
}

// Validation states
export const ValidationStates: Story = {
  render: () => (
    <div className="flex flex-col gap-4">
      <Checkbox 
        label="Terms of Service"
        error="You must accept the terms to continue"
      />
      <Checkbox 
        variant="success"
        label="Email verified"
        description="Your email has been successfully verified"
        defaultChecked
      />
      <Checkbox 
        variant="warning"
        label="Optional marketing"
        description="This setting is recommended but not required"
      />
    </div>
  )
}

// Indeterminate state
export const IndeterminateState: Story = {
  render: () => {
    const [parentChecked, setParentChecked] = useState(false)
    const [parentIndeterminate, setParentIndeterminate] = useState(true)
    const [childStates, setChildStates] = useState([true, false, true])

    const handleParentChange = (e: React.ChangeEvent<HTMLInputElement>) => {
      const checked = e.target.checked
      setParentChecked(checked)
      setParentIndeterminate(false)
      setChildStates([checked, checked, checked])
    }

    const handleChildChange = (index: number, checked: boolean) => {
      const newStates = [...childStates]
      newStates[index] = checked
      setChildStates(newStates)

      const allChecked = newStates.every(state => state)
      const someChecked = newStates.some(state => state)

      setParentChecked(allChecked)
      setParentIndeterminate(someChecked && !allChecked)
    }

    return (
      <div className="space-y-2">
        <Checkbox
          label="Select all features"
          checked={parentChecked}
          indeterminate={parentIndeterminate}
          onChange={handleParentChange}
        />
        <div className="ml-6 space-y-2">
          <Checkbox
            label="Feature 1"
            checked={childStates[0]}
            onChange={(e) => handleChildChange(0, e.target.checked)}
          />
          <Checkbox
            label="Feature 2"
            checked={childStates[1]}
            onChange={(e) => handleChildChange(1, e.target.checked)}
          />
          <Checkbox
            label="Feature 3"
            checked={childStates[2]}
            onChange={(e) => handleChildChange(2, e.target.checked)}
          />
        </div>
      </div>
    )
  }
}

// Custom icons
export const CustomIcons: Story = {
  render: () => (
    <div className="flex flex-col gap-4">
      <Checkbox 
        label="Favorite"
        checkedIcon={<Heart className="h-3 w-3 fill-current" />}
        defaultChecked
      />
      <Checkbox 
        label="Starred"
        checkedIcon={<Star className="h-3 w-3 fill-current" />}
      />
      <Checkbox 
        label="Settings"
        checkedIcon={<Settings className="h-3 w-3" />}
        indeterminateIcon={<Settings className="h-3 w-3 animate-spin" />}
        indeterminate
      />
    </div>
  )
}

// Custom colors
export const CustomColors: Story = {
  render: () => (
    <div className="flex flex-col gap-4">
      <Checkbox 
        label="Purple theme"
        checkedColor="#8B5CF6"
        defaultChecked
      />
      <Checkbox 
        label="Pink theme"
        checkedColor="#EC4899"
        defaultChecked
      />
      <Checkbox 
        label="Teal theme"
        checkedColor="#14B8A6"
        defaultChecked
      />
      <Checkbox 
        label="Orange theme"
        checkedColor="#F97316"
        defaultChecked
      />
    </div>
  )
}

// Disabled states
export const DisabledStates: Story = {
  render: () => (
    <div className="flex flex-col gap-4">
      <Checkbox label="Unchecked disabled" disabled />
      <Checkbox label="Checked disabled" defaultChecked disabled />
      <Checkbox label="Indeterminate disabled" indeterminate disabled />
    </div>
  )
}

// Checkbox group - Basic
export const GroupBasic: Story = {
  render: () => {
    const [selected, setSelected] = useState<string[]>(['option1', 'option3'])

    return (
      <CheckboxGroup
        label="Select your preferences"
        options={[
          { value: 'option1', label: 'Option 1' },
          { value: 'option2', label: 'Option 2' },
          { value: 'option3', label: 'Option 3' },
          { value: 'option4', label: 'Option 4' }
        ]}
        value={selected}
        onChange={setSelected}
      />
    )
  }
}

// Checkbox group - Layouts
export const GroupLayouts: Story = {
  render: () => {
    const options = [
      { value: 'feature1', label: 'Feature 1' },
      { value: 'feature2', label: 'Feature 2' },
      { value: 'feature3', label: 'Feature 3' },
      { value: 'feature4', label: 'Feature 4' }
    ]

    return (
      <div className="space-y-8">
        <CheckboxGroup
          label="Horizontal layout"
          layout="horizontal"
          options={options}
          defaultValue={['feature1']}
        />
        
        <CheckboxGroup
          label="Vertical layout (default)"
          layout="vertical"
          options={options}
          defaultValue={['feature2']}
        />
        
        <CheckboxGroup
          label="Grid layout"
          layout="grid"
          options={options}
          defaultValue={['feature3']}
        />
      </div>
    )
  }
}

// Checkbox group - With descriptions
export const GroupWithDescriptions: Story = {
  render: () => (
    <CheckboxGroup
      label="Notification preferences"
      options={[
        { 
          value: 'email',
          label: 'Email notifications',
          description: 'Receive updates via email'
        },
        { 
          value: 'sms',
          label: 'SMS notifications',
          description: 'Receive text messages for urgent updates'
        },
        { 
          value: 'push',
          label: 'Push notifications',
          description: 'Get instant notifications on your device'
        },
        { 
          value: 'in-app',
          label: 'In-app notifications',
          description: 'See notifications within the application'
        }
      ]}
      defaultValue={['email', 'in-app']}
    />
  )
}

// Checkbox group - Validation
export const GroupValidation: Story = {
  render: () => (
    <div className="space-y-6">
      <CheckboxGroup
        label="Required selection"
        required
        error="Please select at least one option"
        options={[
          { value: 'option1', label: 'Option 1' },
          { value: 'option2', label: 'Option 2' },
          { value: 'option3', label: 'Option 3' }
        ]}
      />
      
      <CheckboxGroup
        label="Some options disabled"
        options={[
          { value: 'free', label: 'Free tier', disabled: true },
          { value: 'pro', label: 'Pro tier' },
          { value: 'enterprise', label: 'Enterprise tier' },
          { value: 'custom', label: 'Custom tier', disabled: true }
        ]}
        defaultValue={['pro']}
      />
    </div>
  )
}

// Real-world example - Settings panel
export const RealWorldSettings: Story = {
  render: () => {
    const [privacy, setPrivacy] = useState(['profile-visible', 'show-activity'])
    const [notifications, setNotifications] = useState(['email-updates'])
    const [security, setSecurity] = useState(['two-factor'])

    return (
      <div className="w-full max-w-md space-y-6 p-6 bg-background border rounded-lg">
        <h3 className="text-lg font-semibold">Account Settings</h3>
        
        <CheckboxGroup
          label="Privacy"
          value={privacy}
          onChange={setPrivacy}
          options={[
            { 
              value: 'profile-visible',
              label: 'Make profile visible',
              description: 'Others can see your profile information'
            },
            { 
              value: 'show-activity',
              label: 'Show activity status',
              description: 'Display when you were last active'
            },
            { 
              value: 'searchable',
              label: 'Appear in search results',
              description: 'Allow others to find you via search'
            }
          ]}
        />
        
        <CheckboxGroup
          label="Notifications"
          value={notifications}
          onChange={setNotifications}
          options={[
            { 
              value: 'email-updates',
              label: 'Email updates',
              description: 'Receive important updates via email'
            },
            { 
              value: 'marketing',
              label: 'Marketing communications',
              description: 'Get news about features and offers'
            },
            { 
              value: 'tips',
              label: 'Tips and tutorials',
              description: 'Learn how to get the most out of our service'
            }
          ]}
        />
        
        <CheckboxGroup
          label="Security"
          value={security}
          onChange={setSecurity}
          options={[
            { 
              value: 'two-factor',
              label: 'Two-factor authentication',
              description: 'Add an extra layer of security'
            },
            { 
              value: 'login-alerts',
              label: 'Login alerts',
              description: 'Get notified of new login attempts'
            },
            { 
              value: 'api-access',
              label: 'API access',
              description: 'Enable third-party integrations',
              disabled: true
            }
          ]}
        />
        
        <div className="pt-4 flex justify-end gap-2">
          <button className="px-4 py-2 text-sm border rounded-md hover:bg-muted">
            Cancel
          </button>
          <button className="px-4 py-2 text-sm bg-primary text-primary-foreground rounded-md hover:bg-primary/90">
            Save Changes
          </button>
        </div>
      </div>
    )
  }
}