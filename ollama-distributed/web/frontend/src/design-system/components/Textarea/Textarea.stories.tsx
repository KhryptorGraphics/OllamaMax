import type { Meta, StoryObj } from '@storybook/react'
import { useState } from 'react'
import { Textarea } from './Textarea'

const meta = {
  title: 'Design System/Forms/Textarea',
  component: Textarea,
  parameters: {
    layout: 'centered',
    docs: {
      description: {
        component: `
The Textarea component provides a multi-line text input with advanced features like auto-resize, character counting, and validation states.

## Features
- **Multi-line Input**: Support for multiple lines of text
- **Auto-resize**: Automatically adjusts height based on content
- **Character Counter**: Shows current/max character count
- **Validation States**: Error, success, and warning states
- **Size Variants**: Small, medium, and large sizes
- **Resize Control**: Control resize behavior (none, vertical, horizontal, both)
- **Read-only & Disabled**: Support for non-editable states
- **Accessibility**: Full ARIA support and keyboard navigation

## Accessibility
- Proper labeling and description association
- \`aria-invalid\` for error states
- \`aria-describedby\` for helper text and errors
- Character count announcements for screen readers
- Keyboard navigation support

## Usage
\`\`\`tsx
import { Textarea } from '@/design-system/components/Textarea'

// Basic textarea
<Textarea label="Description" placeholder="Enter description..." />

// With character limit
<Textarea 
  label="Bio"
  maxLength={200}
  showCounter
/>

// Auto-resize
<Textarea 
  label="Comments"
  autoResize
  minRows={3}
  maxRows={10}
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
    resize: {
      control: 'select',
      options: ['none', 'vertical', 'horizontal', 'both']
    },
    label: {
      control: 'text'
    },
    helperText: {
      control: 'text'
    },
    error: {
      control: 'text'
    },
    success: {
      control: 'text'
    },
    warning: {
      control: 'text'
    },
    maxLength: {
      control: 'number'
    },
    rows: {
      control: 'number'
    },
    disabled: {
      control: 'boolean'
    },
    readOnly: {
      control: 'boolean'
    },
    autoResize: {
      control: 'boolean'
    },
    showCounter: {
      control: 'boolean'
    }
  }
} satisfies Meta<typeof Textarea>

export default meta
type Story = StoryObj<typeof meta>

// Basic textarea
export const Default: Story = {
  args: {
    label: 'Description',
    placeholder: 'Enter your description here...'
  }
}

// Different sizes
export const Sizes: Story = {
  render: () => (
    <div className="flex flex-col gap-4 w-96">
      <Textarea 
        size="sm" 
        label="Small textarea" 
        placeholder="Small size..."
        rows={3}
      />
      <Textarea 
        size="md" 
        label="Medium textarea (default)" 
        placeholder="Medium size..."
        rows={3}
      />
      <Textarea 
        size="lg" 
        label="Large textarea" 
        placeholder="Large size..."
        rows={3}
      />
    </div>
  )
}

// Different variants
export const Variants: Story = {
  render: () => (
    <div className="flex flex-col gap-4 w-96">
      <Textarea 
        variant="default" 
        label="Default variant" 
        defaultValue="Default styling"
        rows={3}
      />
      <Textarea 
        variant="error" 
        label="Error variant" 
        defaultValue="Error styling"
        error="This field has an error"
        rows={3}
      />
      <Textarea 
        variant="success" 
        label="Success variant" 
        defaultValue="Success styling"
        success="Field validated successfully"
        rows={3}
      />
      <Textarea 
        variant="warning" 
        label="Warning variant" 
        defaultValue="Warning styling"
        warning="Please review this field"
        rows={3}
      />
    </div>
  )
}

// With helper text
export const WithHelperText: Story = {
  render: () => (
    <div className="w-96">
      <Textarea 
        label="Bio"
        helperText="Tell us about yourself. This will be displayed on your public profile."
        placeholder="Enter your bio..."
        rows={4}
      />
    </div>
  )
}

// Character counter
export const CharacterCounter: Story = {
  render: () => {
    const [value1, setValue1] = useState('')
    const [value2, setValue2] = useState('This is some initial text that demonstrates the character counter.')
    const [value3, setValue3] = useState('This text exceeds the maximum character limit and will show an error state when typing more characters.')

    return (
      <div className="flex flex-col gap-4 w-96">
        <Textarea 
          label="Short message"
          placeholder="Type your message..."
          maxLength={100}
          showCounter
          value={value1}
          onChange={(e) => setValue1(e.target.value)}
          rows={3}
        />
        
        <Textarea 
          label="Description"
          helperText="Provide a detailed description"
          maxLength={200}
          showCounter
          value={value2}
          onChange={(e) => setValue2(e.target.value)}
          rows={4}
        />
        
        <Textarea 
          label="Limited input"
          maxLength={80}
          showCounter
          value={value3}
          onChange={(e) => setValue3(e.target.value)}
          rows={3}
        />
      </div>
    )
  }
}

// Auto-resize functionality
export const AutoResize: Story = {
  render: () => {
    const [value, setValue] = useState('Start typing to see the textarea automatically resize. The height will adjust based on the content.\n\nTry adding more lines to see it grow!')

    return (
      <div className="flex flex-col gap-4 w-96">
        <Textarea 
          label="Auto-resize textarea"
          helperText="This textarea will grow as you type"
          autoResize
          minRows={3}
          maxRows={10}
          value={value}
          onChange={(e) => setValue(e.target.value)}
          placeholder="Start typing..."
        />
        
        <Textarea 
          label="Auto-resize with character limit"
          autoResize
          minRows={2}
          maxRows={8}
          maxLength={500}
          showCounter
          placeholder="Type here..."
        />
      </div>
    )
  }
}

// Resize control
export const ResizeControl: Story = {
  render: () => (
    <div className="flex flex-col gap-4 w-96">
      <Textarea 
        label="No resize"
        resize="none"
        defaultValue="This textarea cannot be resized"
        rows={3}
      />
      
      <Textarea 
        label="Vertical resize only (default)"
        resize="vertical"
        defaultValue="This textarea can only be resized vertically"
        rows={3}
      />
      
      <Textarea 
        label="Horizontal resize only"
        resize="horizontal"
        defaultValue="This textarea can only be resized horizontally"
        rows={3}
      />
      
      <Textarea 
        label="Both directions"
        resize="both"
        defaultValue="This textarea can be resized in both directions"
        rows={3}
      />
    </div>
  )
}

// Validation states
export const ValidationStates: Story = {
  render: () => (
    <div className="flex flex-col gap-4 w-96">
      <Textarea 
        label="Required field"
        required
        error="This field is required"
        placeholder="Enter required text..."
        rows={3}
      />
      
      <Textarea 
        label="Validated field"
        success="Content looks good!"
        defaultValue="This content has been validated"
        rows={3}
      />
      
      <Textarea 
        label="Warning field"
        warning="This content may contain sensitive information"
        defaultValue="Some content here"
        rows={3}
      />
      
      <Textarea 
        label="With validation icon"
        error="Please fix this error"
        showValidationIcon
        defaultValue="Error content"
        rows={3}
      />
    </div>
  )
}

// Disabled and readonly states
export const DisabledReadonly: Story = {
  render: () => (
    <div className="flex flex-col gap-4 w-96">
      <Textarea 
        label="Disabled textarea"
        disabled
        defaultValue="This textarea is disabled and cannot be edited"
        rows={3}
      />
      
      <Textarea 
        label="Read-only textarea"
        readOnly
        defaultValue="This textarea is read-only. You can select and copy the text but cannot edit it."
        rows={3}
      />
      
      <Textarea 
        label="Disabled with helper text"
        disabled
        helperText="This feature is currently unavailable"
        placeholder="Coming soon..."
        rows={3}
      />
    </div>
  )
}

// Different row configurations
export const RowConfigurations: Story = {
  render: () => (
    <div className="flex flex-col gap-4 w-96">
      <Textarea 
        label="2 rows"
        rows={2}
        placeholder="Compact textarea..."
      />
      
      <Textarea 
        label="4 rows (default)"
        rows={4}
        placeholder="Standard textarea..."
      />
      
      <Textarea 
        label="8 rows"
        rows={8}
        placeholder="Large textarea..."
      />
    </div>
  )
}

// Real-world example - Comment form
export const RealWorldComment: Story = {
  render: () => {
    const [comment, setComment] = useState('')
    const [isSubmitting, setIsSubmitting] = useState(false)
    const maxLength = 500

    const handleSubmit = () => {
      setIsSubmitting(true)
      setTimeout(() => {
        setIsSubmitting(false)
        setComment('')
      }, 1500)
    }

    return (
      <div className="w-full max-w-lg space-y-4 p-6 bg-background border rounded-lg">
        <h3 className="text-lg font-semibold">Leave a comment</h3>
        
        <Textarea
          label="Your comment"
          placeholder="Share your thoughts..."
          value={comment}
          onChange={(e) => setComment(e.target.value)}
          maxLength={maxLength}
          showCounter
          autoResize
          minRows={3}
          maxRows={8}
          required
          helperText="Be respectful and constructive"
          disabled={isSubmitting}
        />
        
        <div className="flex items-center justify-between">
          <div className="text-xs text-muted-foreground">
            Markdown formatting supported
          </div>
          <div className="flex gap-2">
            <button 
              className="px-4 py-2 text-sm border rounded-md hover:bg-muted"
              onClick={() => setComment('')}
              disabled={!comment || isSubmitting}
            >
              Clear
            </button>
            <button 
              className="px-4 py-2 text-sm bg-primary text-primary-foreground rounded-md hover:bg-primary/90 disabled:opacity-50"
              onClick={handleSubmit}
              disabled={!comment || isSubmitting}
            >
              {isSubmitting ? 'Posting...' : 'Post Comment'}
            </button>
          </div>
        </div>
      </div>
    )
  }
}

// Real-world example - Feedback form
export const RealWorldFeedback: Story = {
  render: () => {
    const [category, setCategory] = useState('bug')
    const [title, setTitle] = useState('')
    const [description, setDescription] = useState('')
    const [steps, setSteps] = useState('')

    return (
      <div className="w-full max-w-2xl space-y-6 p-6 bg-background border rounded-lg">
        <div>
          <h3 className="text-lg font-semibold">Submit Feedback</h3>
          <p className="text-sm text-muted-foreground mt-1">
            Help us improve by reporting issues or suggesting features
          </p>
        </div>
        
        <div className="space-y-4">
          <div>
            <label className="text-sm font-medium">Category</label>
            <select 
              className="w-full mt-2 px-3 py-2 border rounded-md"
              value={category}
              onChange={(e) => setCategory(e.target.value)}
            >
              <option value="bug">Bug Report</option>
              <option value="feature">Feature Request</option>
              <option value="improvement">Improvement</option>
              <option value="other">Other</option>
            </select>
          </div>
          
          <div>
            <label className="text-sm font-medium">Title</label>
            <input 
              className="w-full mt-2 px-3 py-2 border rounded-md"
              placeholder="Brief summary of your feedback"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
            />
          </div>
          
          <Textarea
            label="Description"
            placeholder="Describe your feedback in detail..."
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            required
            maxLength={1000}
            showCounter
            autoResize
            minRows={4}
            maxRows={10}
            helperText="Include as much detail as possible"
          />
          
          {category === 'bug' && (
            <Textarea
              label="Steps to reproduce"
              placeholder="1. Go to...\n2. Click on...\n3. See error..."
              value={steps}
              onChange={(e) => setSteps(e.target.value)}
              autoResize
              minRows={3}
              maxRows={8}
              helperText="Help us reproduce the issue"
            />
          )}
        </div>
        
        <div className="pt-4 flex justify-end gap-2">
          <button className="px-4 py-2 text-sm border rounded-md hover:bg-muted">
            Cancel
          </button>
          <button 
            className="px-4 py-2 text-sm bg-primary text-primary-foreground rounded-md hover:bg-primary/90 disabled:opacity-50"
            disabled={!title || !description}
          >
            Submit Feedback
          </button>
        </div>
      </div>
    )
  }
}

// Real-world example - Article editor
export const RealWorldEditor: Story = {
  render: () => {
    const [title, setTitle] = useState('Getting Started with React')
    const [excerpt, setExcerpt] = useState('Learn the fundamentals of React, including components, state, and props.')
    const [content, setContent] = useState(`## Introduction

React is a powerful JavaScript library for building user interfaces. In this article, we'll explore the core concepts that make React such a popular choice for modern web development.

## Key Concepts

### Components
Components are the building blocks of React applications. They can be either functional or class-based.

### State and Props
State represents the internal data of a component, while props are used to pass data between components.

### Hooks
Hooks allow you to use state and other React features in functional components.`)

    return (
      <div className="w-full max-w-4xl space-y-6 p-6 bg-background border rounded-lg">
        <div className="flex items-center justify-between">
          <h3 className="text-xl font-semibold">Edit Article</h3>
          <div className="flex gap-2">
            <button className="px-3 py-1.5 text-sm border rounded-md hover:bg-muted">
              Save Draft
            </button>
            <button className="px-3 py-1.5 text-sm bg-primary text-primary-foreground rounded-md hover:bg-primary/90">
              Publish
            </button>
          </div>
        </div>
        
        <div className="space-y-4">
          <div>
            <label className="text-sm font-medium">Title</label>
            <input 
              className="w-full mt-2 px-3 py-2 border rounded-md text-lg font-medium"
              value={title}
              onChange={(e) => setTitle(e.target.value)}
            />
          </div>
          
          <Textarea
            label="Excerpt"
            placeholder="Brief description of your article..."
            value={excerpt}
            onChange={(e) => setExcerpt(e.target.value)}
            maxLength={200}
            showCounter
            rows={2}
            helperText="This will appear in article previews"
          />
          
          <Textarea
            label="Content"
            placeholder="Write your article content here..."
            value={content}
            onChange={(e) => setContent(e.target.value)}
            autoResize
            minRows={10}
            maxRows={30}
            helperText="Markdown formatting is supported"
            className="font-mono"
          />
        </div>
        
        <div className="pt-4 border-t flex items-center justify-between text-sm">
          <div className="text-muted-foreground">
            Last saved: 2 minutes ago
          </div>
          <div className="text-muted-foreground">
            Word count: {content.split(/\s+/).length}
          </div>
        </div>
      </div>
    )
  }
}