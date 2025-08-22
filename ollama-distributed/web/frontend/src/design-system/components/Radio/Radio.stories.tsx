import type { Meta, StoryObj } from '@storybook/react'
import { useState } from 'react'
import { Radio, RadioGroup } from './Radio'
import { CreditCard, Wallet, Banknote, Building2 } from 'lucide-react'

const meta = {
  title: 'Design System/Forms/Radio',
  component: Radio,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: `
The Radio component provides accessible radio buttons with group management for mutually exclusive selections.

## Features
- **Mutual Exclusion**: Only one option can be selected in a group
- **Variants**: Default, error, success, and warning styles
- **Sizes**: Small, medium, and large sizes
- **Custom Colors**: Ability to customize selected state color
- **Group Layouts**: Horizontal, vertical, and grid layouts
- **Keyboard Navigation**: Arrow keys for navigation within groups
- **Accessibility**: Full ARIA support and proper grouping

## Accessibility
- Proper \`radiogroup\` role and labeling
- Arrow key navigation (Up/Down, Left/Right)
- Home/End key support for first/last item
- \`aria-invalid\` for error states
- \`aria-describedby\` for descriptions

## Usage
\`\`\`tsx
import { Radio, RadioGroup } from '@/design-system/components/Radio'

// Basic radio button
<Radio name="option" label="Option 1" />

// Radio group
<RadioGroup
  label="Select an option"
  options={[
    { value: 'option1', label: 'Option 1' },
    { value: 'option2', label: 'Option 2' }
  ]}
  value={selected}
  onChange={setSelected}
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
    checkedColor: {
      control: 'color'
    }
  }
} satisfies Meta<typeof Radio>

export default meta
type Story = StoryObj<typeof meta>

// Basic radio button
export const Default: Story = {
  args: {
    label: 'Radio option',
    name: 'example'
  }
}

// Different sizes
export const Sizes: Story = {
  render: () => {
    const [selected, setSelected] = useState('md')
    
    return (
      <div className="flex flex-col gap-4">
        <Radio 
          size="sm" 
          label="Small radio" 
          name="size"
          value="sm"
          checked={selected === 'sm'}
          onChange={() => setSelected('sm')}
        />
        <Radio 
          size="md" 
          label="Medium radio (default)" 
          name="size"
          value="md"
          checked={selected === 'md'}
          onChange={() => setSelected('md')}
        />
        <Radio 
          size="lg" 
          label="Large radio" 
          name="size"
          value="lg"
          checked={selected === 'lg'}
          onChange={() => setSelected('lg')}
        />
      </div>
    )
  }
}

// Different variants
export const Variants: Story = {
  render: () => (
    <div className="flex flex-col gap-4">
      <Radio variant="default" label="Default variant" name="variant1" defaultChecked />
      <Radio variant="error" label="Error variant" name="variant2" defaultChecked />
      <Radio variant="success" label="Success variant" name="variant3" defaultChecked />
      <Radio variant="warning" label="Warning variant" name="variant4" defaultChecked />
    </div>
  )
}

// With descriptions
export const WithDescriptions: Story = {
  render: () => {
    const [selected, setSelected] = useState('standard')
    
    return (
      <div className="flex flex-col gap-4">
        <Radio 
          label="Standard shipping"
          description="5-7 business days"
          name="shipping"
          value="standard"
          checked={selected === 'standard'}
          onChange={() => setSelected('standard')}
        />
        <Radio 
          label="Express shipping"
          description="2-3 business days"
          name="shipping"
          value="express"
          checked={selected === 'express'}
          onChange={() => setSelected('express')}
        />
        <Radio 
          label="Next day delivery"
          description="Delivered by tomorrow"
          name="shipping"
          value="nextday"
          checked={selected === 'nextday'}
          onChange={() => setSelected('nextday')}
        />
      </div>
    )
  }
}

// Custom colors
export const CustomColors: Story = {
  render: () => {
    const [selected, setSelected] = useState('purple')
    
    return (
      <div className="flex flex-col gap-4">
        <Radio 
          label="Purple theme"
          checkedColor="#8B5CF6"
          name="color"
          value="purple"
          checked={selected === 'purple'}
          onChange={() => setSelected('purple')}
        />
        <Radio 
          label="Pink theme"
          checkedColor="#EC4899"
          name="color"
          value="pink"
          checked={selected === 'pink'}
          onChange={() => setSelected('pink')}
        />
        <Radio 
          label="Teal theme"
          checkedColor="#14B8A6"
          name="color"
          value="teal"
          checked={selected === 'teal'}
          onChange={() => setSelected('teal')}
        />
        <Radio 
          label="Orange theme"
          checkedColor="#F97316"
          name="color"
          value="orange"
          checked={selected === 'orange'}
          onChange={() => setSelected('orange')}
        />
      </div>
    )
  }
}

// Disabled states
export const DisabledStates: Story = {
  render: () => (
    <div className="flex flex-col gap-4">
      <Radio label="Unchecked disabled" name="disabled" disabled />
      <Radio label="Checked disabled" name="disabled2" defaultChecked disabled />
      <div className="text-sm text-muted-foreground">
        Disabled radio buttons cannot be selected
      </div>
    </div>
  )
}

// Radio group - Basic
export const GroupBasic: Story = {
  render: () => {
    const [selected, setSelected] = useState('option2')

    return (
      <RadioGroup
        label="Select an option"
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

// Radio group - Layouts
export const GroupLayouts: Story = {
  render: () => {
    const options = [
      { value: 'plan1', label: 'Basic' },
      { value: 'plan2', label: 'Professional' },
      { value: 'plan3', label: 'Business' },
      { value: 'plan4', label: 'Enterprise' }
    ]

    return (
      <div className="space-y-8">
        <RadioGroup
          label="Horizontal layout"
          layout="horizontal"
          options={options}
          defaultValue="plan1"
        />
        
        <RadioGroup
          label="Vertical layout (default)"
          layout="vertical"
          options={options}
          defaultValue="plan2"
        />
        
        <RadioGroup
          label="Grid layout"
          layout="grid"
          options={options}
          defaultValue="plan3"
        />
      </div>
    )
  }
}

// Radio group - With descriptions
export const GroupWithDescriptions: Story = {
  render: () => {
    const [plan, setPlan] = useState('pro')

    return (
      <RadioGroup
        label="Choose your plan"
        value={plan}
        onChange={setPlan}
        options={[
          { 
            value: 'free',
            label: 'Free',
            description: 'For individuals just getting started'
          },
          { 
            value: 'pro',
            label: 'Pro',
            description: 'For professionals and small teams'
          },
          { 
            value: 'business',
            label: 'Business',
            description: 'For growing companies'
          },
          { 
            value: 'enterprise',
            label: 'Enterprise',
            description: 'For large organizations with custom needs'
          }
        ]}
      />
    )
  }
}

// Radio group - Different sizes
export const GroupSizes: Story = {
  render: () => (
    <div className="space-y-6">
      <RadioGroup
        label="Small radio buttons"
        size="sm"
        options={[
          { value: 'opt1', label: 'Option 1' },
          { value: 'opt2', label: 'Option 2' },
          { value: 'opt3', label: 'Option 3' }
        ]}
        defaultValue="opt1"
      />
      
      <RadioGroup
        label="Medium radio buttons (default)"
        size="md"
        options={[
          { value: 'opt1', label: 'Option 1' },
          { value: 'opt2', label: 'Option 2' },
          { value: 'opt3', label: 'Option 3' }
        ]}
        defaultValue="opt2"
      />
      
      <RadioGroup
        label="Large radio buttons"
        size="lg"
        options={[
          { value: 'opt1', label: 'Option 1' },
          { value: 'opt2', label: 'Option 2' },
          { value: 'opt3', label: 'Option 3' }
        ]}
        defaultValue="opt3"
      />
    </div>
  )
}

// Radio group - Validation states
export const GroupValidation: Story = {
  render: () => (
    <div className="space-y-6">
      <RadioGroup
        label="Required selection"
        required
        error="Please select an option"
        options={[
          { value: 'option1', label: 'Option 1' },
          { value: 'option2', label: 'Option 2' },
          { value: 'option3', label: 'Option 3' }
        ]}
      />
      
      <RadioGroup
        label="Some options disabled"
        options={[
          { value: 'available1', label: 'Available option 1' },
          { value: 'unavailable', label: 'Unavailable option', disabled: true },
          { value: 'available2', label: 'Available option 2' },
          { value: 'soldout', label: 'Sold out', disabled: true }
        ]}
        defaultValue="available1"
      />
    </div>
  )
}

// Keyboard navigation demo
export const KeyboardNavigation: Story = {
  render: () => {
    const [selected, setSelected] = useState('option2')

    return (
      <div className="space-y-4">
        <div className="text-sm text-muted-foreground">
          Use arrow keys (↑↓←→) to navigate, Home/End for first/last item
        </div>
        <RadioGroup
          label="Navigate with keyboard"
          value={selected}
          onChange={setSelected}
          options={[
            { value: 'option1', label: 'First option' },
            { value: 'option2', label: 'Second option' },
            { value: 'option3', label: 'Third option' },
            { value: 'option4', label: 'Fourth option' },
            { value: 'option5', label: 'Fifth option' }
          ]}
        />
      </div>
    )
  }
}

// Real-world example - Payment method selection
export const RealWorldPayment: Story = {
  render: () => {
    const [paymentMethod, setPaymentMethod] = useState('card')
    const [cardType, setCardType] = useState('personal')

    return (
      <div className="w-full max-w-md space-y-6 p-6 bg-background border rounded-lg">
        <h3 className="text-lg font-semibold">Payment Method</h3>
        
        <RadioGroup
          label="Select payment method"
          value={paymentMethod}
          onChange={setPaymentMethod}
          options={[
            { 
              value: 'card',
              label: 'Credit or Debit Card',
              description: 'Visa, Mastercard, American Express'
            },
            { 
              value: 'paypal',
              label: 'PayPal',
              description: 'Pay with your PayPal account'
            },
            { 
              value: 'bank',
              label: 'Bank Transfer',
              description: 'Direct transfer from your bank account'
            },
            { 
              value: 'crypto',
              label: 'Cryptocurrency',
              description: 'Bitcoin, Ethereum, and more',
              disabled: true
            }
          ]}
        />
        
        {paymentMethod === 'card' && (
          <RadioGroup
            label="Card type"
            value={cardType}
            onChange={setCardType}
            layout="horizontal"
            options={[
              { value: 'personal', label: 'Personal' },
              { value: 'business', label: 'Business' }
            ]}
          />
        )}
        
        <div className="pt-4 border-t">
          <div className="flex items-center justify-between text-sm">
            <span className="text-muted-foreground">Processing fee</span>
            <span className="font-medium">
              {paymentMethod === 'card' ? '2.9% + $0.30' : 
               paymentMethod === 'paypal' ? '3.49% + $0.49' :
               paymentMethod === 'bank' ? '$0.00' : 'N/A'}
            </span>
          </div>
        </div>
        
        <button className="w-full px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90">
          Continue to Payment
        </button>
      </div>
    )
  }
}

// Real-world example - Survey form
export const RealWorldSurvey: Story = {
  render: () => {
    const [experience, setExperience] = useState('')
    const [satisfaction, setSatisfaction] = useState('')
    const [recommend, setRecommend] = useState('')

    return (
      <div className="w-full max-w-md space-y-6 p-6 bg-background border rounded-lg">
        <div>
          <h3 className="text-lg font-semibold">Customer Feedback</h3>
          <p className="text-sm text-muted-foreground mt-1">
            Help us improve by sharing your experience
          </p>
        </div>
        
        <RadioGroup
          label="How would you rate your overall experience?"
          value={experience}
          onChange={setExperience}
          required
          options={[
            { value: 'excellent', label: 'Excellent' },
            { value: 'good', label: 'Good' },
            { value: 'average', label: 'Average' },
            { value: 'poor', label: 'Poor' },
            { value: 'terrible', label: 'Terrible' }
          ]}
        />
        
        <RadioGroup
          label="How satisfied are you with our service?"
          value={satisfaction}
          onChange={setSatisfaction}
          required
          layout="horizontal"
          options={[
            { value: 'very-satisfied', label: 'Very satisfied' },
            { value: 'satisfied', label: 'Satisfied' },
            { value: 'neutral', label: 'Neutral' },
            { value: 'dissatisfied', label: 'Dissatisfied' },
            { value: 'very-dissatisfied', label: 'Very dissatisfied' }
          ]}
        />
        
        <RadioGroup
          label="Would you recommend us to others?"
          value={recommend}
          onChange={setRecommend}
          required
          options={[
            { value: 'definitely', label: 'Definitely' },
            { value: 'probably', label: 'Probably' },
            { value: 'not-sure', label: 'Not sure' },
            { value: 'probably-not', label: 'Probably not' },
            { value: 'definitely-not', label: 'Definitely not' }
          ]}
        />
        
        <div className="pt-4 flex justify-end gap-2">
          <button className="px-4 py-2 text-sm border rounded-md hover:bg-muted">
            Skip
          </button>
          <button 
            className="px-4 py-2 text-sm bg-primary text-primary-foreground rounded-md hover:bg-primary/90 disabled:opacity-50"
            disabled={!experience || !satisfaction || !recommend}
          >
            Submit Feedback
          </button>
        </div>
      </div>
    )
  }
}