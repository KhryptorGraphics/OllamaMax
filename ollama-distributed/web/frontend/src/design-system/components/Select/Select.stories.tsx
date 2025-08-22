import type { Meta, StoryObj } from '@storybook/react';
import { Select } from './Select';
import { User, Mail, Phone, Home, Settings, Star } from 'lucide-react';

const meta = {
  title: 'Design System/Components/Select',
  component: Select,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
  argTypes: {
    size: {
      control: 'select',
      options: ['sm', 'md', 'lg'],
    },
    multiple: {
      control: 'boolean',
    },
    searchable: {
      control: 'boolean',
    },
    clearable: {
      control: 'boolean',
    },
    disabled: {
      control: 'boolean',
    },
    error: {
      control: 'boolean',
    },
    loading: {
      control: 'boolean',
    },
  },
} satisfies Meta<typeof Select>;

export default meta;
type Story = StoryObj<typeof meta>;

const basicOptions = [
  { value: 'option1', label: 'Option 1' },
  { value: 'option2', label: 'Option 2' },
  { value: 'option3', label: 'Option 3' },
  { value: 'option4', label: 'Option 4', disabled: true },
  { value: 'option5', label: 'Option 5' },
];

const iconOptions = [
  { value: 'user', label: 'User Profile', icon: <User className="w-4 h-4" /> },
  { value: 'mail', label: 'Email Settings', icon: <Mail className="w-4 h-4" /> },
  { value: 'phone', label: 'Phone Numbers', icon: <Phone className="w-4 h-4" /> },
  { value: 'home', label: 'Home Address', icon: <Home className="w-4 h-4" /> },
  { value: 'settings', label: 'General Settings', icon: <Settings className="w-4 h-4" /> },
];

const groupedOptions = [
  { value: 'john', label: 'John Doe', group: 'Users' },
  { value: 'jane', label: 'Jane Smith', group: 'Users' },
  { value: 'bob', label: 'Bob Johnson', group: 'Users' },
  { value: 'admin', label: 'Admin Role', group: 'Roles' },
  { value: 'editor', label: 'Editor Role', group: 'Roles' },
  { value: 'viewer', label: 'Viewer Role', group: 'Roles' },
  { value: 'project1', label: 'Project Alpha', group: 'Projects' },
  { value: 'project2', label: 'Project Beta', group: 'Projects' },
];

const descriptionOptions = [
  {
    value: 'basic',
    label: 'Basic Plan',
    description: 'Perfect for individuals and small teams',
  },
  {
    value: 'pro',
    label: 'Pro Plan',
    description: 'Advanced features for growing businesses',
  },
  {
    value: 'enterprise',
    label: 'Enterprise Plan',
    description: 'Custom solutions for large organizations',
    disabled: true,
  },
];

export const Default: Story = {
  args: {
    options: basicOptions,
    placeholder: 'Select an option',
  },
};

export const SingleSelect: Story = {
  args: {
    options: basicOptions,
    placeholder: 'Choose one option',
    clearable: true,
  },
};

export const MultiSelect: Story = {
  args: {
    options: basicOptions,
    placeholder: 'Select multiple options',
    multiple: true,
    clearable: true,
  },
};

export const Searchable: Story = {
  args: {
    options: groupedOptions,
    placeholder: 'Search and select',
    searchable: true,
    clearable: true,
  },
};

export const SearchableMultiple: Story = {
  args: {
    options: groupedOptions,
    placeholder: 'Search and select multiple',
    multiple: true,
    searchable: true,
    clearable: true,
  },
};

export const WithIcons: Story = {
  args: {
    options: iconOptions,
    placeholder: 'Select with icons',
    searchable: true,
  },
};

export const GroupedOptions: Story = {
  args: {
    options: groupedOptions,
    placeholder: 'Select from groups',
    searchable: true,
  },
};

export const WithDescriptions: Story = {
  args: {
    options: descriptionOptions,
    placeholder: 'Select a plan',
  },
};

export const Sizes: Story = {
  render: () => (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Small</label>
        <Select options={basicOptions} size="sm" placeholder="Small select" />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Medium</label>
        <Select options={basicOptions} size="md" placeholder="Medium select" />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Large</label>
        <Select options={basicOptions} size="lg" placeholder="Large select" />
      </div>
    </div>
  ),
};

export const States: Story = {
  render: () => (
    <div className="space-y-4">
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Normal</label>
        <Select options={basicOptions} placeholder="Normal state" />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Disabled</label>
        <Select options={basicOptions} placeholder="Disabled state" disabled />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Error</label>
        <Select options={basicOptions} placeholder="Error state" error />
      </div>
      <div>
        <label className="block text-sm font-medium text-gray-700 mb-1">Loading</label>
        <Select options={basicOptions} placeholder="Loading state" loading />
      </div>
    </div>
  ),
};

export const CustomRenderOption: Story = {
  args: {
    options: descriptionOptions,
    placeholder: 'Custom rendered options',
    renderOption: (option, isSelected) => (
      <div className={`p-3 cursor-pointer ${isSelected ? 'bg-primary-50' : 'hover:bg-gray-50'}`}>
        <div className="flex items-start justify-between">
          <div>
            <div className="font-medium flex items-center">
              {isSelected && <Star className="w-4 h-4 mr-2 text-yellow-500" />}
              {option.label}
            </div>
            {option.description && (
              <div className="text-sm text-gray-500 mt-1">{option.description}</div>
            )}
          </div>
          {option.disabled && (
            <span className="text-xs bg-gray-100 text-gray-600 px-2 py-1 rounded">
              Coming Soon
            </span>
          )}
        </div>
      </div>
    ),
  },
};

export const ControlledComponent: Story = {
  render: function ControlledSelectStory() {
    const [value, setValue] = React.useState<string>('option2');
    
    return (
      <div className="space-y-4">
        <Select
          options={basicOptions}
          value={value}
          onChange={(newValue) => setValue(newValue as string)}
          placeholder="Controlled select"
          clearable
        />
        <div className="text-sm text-gray-600">
          Selected value: <code className="bg-gray-100 px-2 py-1 rounded">{value || 'none'}</code>
        </div>
      </div>
    );
  },
};

export const ControlledMultiple: Story = {
  render: function ControlledMultiSelectStory() {
    const [values, setValues] = React.useState<string[]>(['option1', 'option3']);
    
    return (
      <div className="space-y-4">
        <Select
          options={basicOptions}
          value={values}
          onChange={(newValues) => setValues(newValues as string[])}
          placeholder="Controlled multi-select"
          multiple
          clearable
        />
        <div className="text-sm text-gray-600">
          Selected values: <code className="bg-gray-100 px-2 py-1 rounded">{JSON.stringify(values)}</code>
        </div>
      </div>
    );
  },
};