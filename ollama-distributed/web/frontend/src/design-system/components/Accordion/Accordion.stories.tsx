import type { Meta, StoryObj } from '@storybook/react';
import { Accordion, AccordionItem, AccordionTrigger, AccordionContent } from './Accordion';
import { Info, Settings, Shield, Bell, CreditCard, User } from 'lucide-react';

const meta = {
  title: 'Design System/Components/Accordion',
  component: Accordion,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
  argTypes: {
    type: {
      control: 'select',
      options: ['single', 'multiple'],
    },
    variant: {
      control: 'select',
      options: ['default', 'bordered', 'ghost', 'filled'],
    },
    size: {
      control: 'select',
      options: ['sm', 'md', 'lg'],
    },
    iconPosition: {
      control: 'select',
      options: ['left', 'right'],
    },
    iconType: {
      control: 'select',
      options: ['chevron', 'arrow', 'plus'],
    },
    animated: {
      control: 'boolean',
    },
  },
} satisfies Meta<typeof Accordion>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <div className="w-[600px]">
      <Accordion type="single" defaultValue="item-1">
        <AccordionItem value="item-1">
          <AccordionTrigger value="item-1">What is this component?</AccordionTrigger>
          <AccordionContent value="item-1">
            This is an accessible accordion component built with React. It supports both single
            and multiple expansion modes, keyboard navigation, and ARIA attributes for screen readers.
          </AccordionContent>
        </AccordionItem>
        <AccordionItem value="item-2">
          <AccordionTrigger value="item-2">How does it work?</AccordionTrigger>
          <AccordionContent value="item-2">
            The accordion uses a context provider to manage state and share it between child components.
            It supports controlled and uncontrolled modes, animations, and various styling options.
          </AccordionContent>
        </AccordionItem>
        <AccordionItem value="item-3">
          <AccordionTrigger value="item-3">Is it accessible?</AccordionTrigger>
          <AccordionContent value="item-3">
            Yes! The accordion follows WAI-ARIA guidelines with proper roles, aria-expanded states,
            and keyboard navigation support. Users can navigate using arrow keys, Home, and End keys.
          </AccordionContent>
        </AccordionItem>
      </Accordion>
    </div>
  ),
};

export const SingleExpand: Story = {
  render: () => (
    <div className="w-[600px]">
      <Accordion type="single" defaultValue="item-1">
        <AccordionItem value="item-1">
          <AccordionTrigger value="item-1">Section 1</AccordionTrigger>
          <AccordionContent value="item-1">
            Only one section can be expanded at a time. Clicking another section will close this one.
          </AccordionContent>
        </AccordionItem>
        <AccordionItem value="item-2">
          <AccordionTrigger value="item-2">Section 2</AccordionTrigger>
          <AccordionContent value="item-2">
            This ensures users focus on one piece of content at a time.
          </AccordionContent>
        </AccordionItem>
        <AccordionItem value="item-3">
          <AccordionTrigger value="item-3">Section 3</AccordionTrigger>
          <AccordionContent value="item-3">
            Perfect for FAQs or step-by-step guides.
          </AccordionContent>
        </AccordionItem>
      </Accordion>
    </div>
  ),
};

export const MultipleExpand: Story = {
  render: () => (
    <div className="w-[600px]">
      <Accordion type="multiple" defaultValue={['item-1', 'item-3']}>
        <AccordionItem value="item-1">
          <AccordionTrigger value="item-1">Section 1</AccordionTrigger>
          <AccordionContent value="item-1">
            Multiple sections can be expanded simultaneously.
          </AccordionContent>
        </AccordionItem>
        <AccordionItem value="item-2">
          <AccordionTrigger value="item-2">Section 2</AccordionTrigger>
          <AccordionContent value="item-2">
            This is useful when users need to compare information across sections.
          </AccordionContent>
        </AccordionItem>
        <AccordionItem value="item-3">
          <AccordionTrigger value="item-3">Section 3</AccordionTrigger>
          <AccordionContent value="item-3">
            Each section operates independently of the others.
          </AccordionContent>
        </AccordionItem>
      </Accordion>
    </div>
  ),
};

export const Variants: Story = {
  render: () => (
    <div className="space-y-8">
      <div className="w-[600px]">
        <h3 className="text-sm font-medium mb-2">Default</h3>
        <Accordion type="single" variant="default">
          <AccordionItem value="item-1">
            <AccordionTrigger value="item-1">Default Style</AccordionTrigger>
            <AccordionContent value="item-1">
              Standard accordion with borders between items.
            </AccordionContent>
          </AccordionItem>
          <AccordionItem value="item-2">
            <AccordionTrigger value="item-2">Second Item</AccordionTrigger>
            <AccordionContent value="item-2">
              Clean and minimal design.
            </AccordionContent>
          </AccordionItem>
        </Accordion>
      </div>

      <div className="w-[600px]">
        <h3 className="text-sm font-medium mb-2">Bordered</h3>
        <Accordion type="single" variant="bordered">
          <AccordionItem value="item-1">
            <AccordionTrigger value="item-1">Bordered Style</AccordionTrigger>
            <AccordionContent value="item-1">
              Each item has its own border and rounded corners.
            </AccordionContent>
          </AccordionItem>
          <AccordionItem value="item-2">
            <AccordionTrigger value="item-2">Second Item</AccordionTrigger>
            <AccordionContent value="item-2">
              Separated items with spacing between them.
            </AccordionContent>
          </AccordionItem>
        </Accordion>
      </div>

      <div className="w-[600px]">
        <h3 className="text-sm font-medium mb-2">Ghost</h3>
        <Accordion type="single" variant="ghost">
          <AccordionItem value="item-1">
            <AccordionTrigger value="item-1">Ghost Style</AccordionTrigger>
            <AccordionContent value="item-1">
              Minimal style with subtle borders.
            </AccordionContent>
          </AccordionItem>
          <AccordionItem value="item-2">
            <AccordionTrigger value="item-2">Second Item</AccordionTrigger>
            <AccordionContent value="item-2">
              Less visual prominence.
            </AccordionContent>
          </AccordionItem>
        </Accordion>
      </div>

      <div className="w-[600px]">
        <h3 className="text-sm font-medium mb-2">Filled</h3>
        <Accordion type="single" variant="filled">
          <AccordionItem value="item-1">
            <AccordionTrigger value="item-1">Filled Style</AccordionTrigger>
            <AccordionContent value="item-1">
              Items have a filled background with shadow.
            </AccordionContent>
          </AccordionItem>
          <AccordionItem value="item-2">
            <AccordionTrigger value="item-2">Second Item</AccordionTrigger>
            <AccordionContent value="item-2">
              More visual depth and separation.
            </AccordionContent>
          </AccordionItem>
        </Accordion>
      </div>
    </div>
  ),
};

export const IconPositions: Story = {
  render: () => (
    <div className="space-y-8">
      <div className="w-[600px]">
        <h3 className="text-sm font-medium mb-2">Icon on Right (default)</h3>
        <Accordion type="single" iconPosition="right">
          <AccordionItem value="item-1">
            <AccordionTrigger value="item-1">Right Icon</AccordionTrigger>
            <AccordionContent value="item-1">
              The expand/collapse icon appears on the right side.
            </AccordionContent>
          </AccordionItem>
          <AccordionItem value="item-2">
            <AccordionTrigger value="item-2">Another Item</AccordionTrigger>
            <AccordionContent value="item-2">
              This is the default icon position.
            </AccordionContent>
          </AccordionItem>
        </Accordion>
      </div>

      <div className="w-[600px]">
        <h3 className="text-sm font-medium mb-2">Icon on Left</h3>
        <Accordion type="single" iconPosition="left">
          <AccordionItem value="item-1">
            <AccordionTrigger value="item-1">Left Icon</AccordionTrigger>
            <AccordionContent value="item-1">
              The expand/collapse icon appears on the left side.
            </AccordionContent>
          </AccordionItem>
          <AccordionItem value="item-2">
            <AccordionTrigger value="item-2">Another Item</AccordionTrigger>
            <AccordionContent value="item-2">
              Alternative icon positioning.
            </AccordionContent>
          </AccordionItem>
        </Accordion>
      </div>
    </div>
  ),
};

export const IconTypes: Story = {
  render: () => (
    <div className="space-y-8">
      <div className="w-[600px]">
        <h3 className="text-sm font-medium mb-2">Chevron Icon</h3>
        <Accordion type="single" iconType="chevron">
          <AccordionItem value="item-1">
            <AccordionTrigger value="item-1">Chevron Icon</AccordionTrigger>
            <AccordionContent value="item-1">Uses a chevron that rotates when expanded.</AccordionContent>
          </AccordionItem>
        </Accordion>
      </div>

      <div className="w-[600px]">
        <h3 className="text-sm font-medium mb-2">Arrow Icon</h3>
        <Accordion type="single" iconType="arrow" iconPosition="left">
          <AccordionItem value="item-1">
            <AccordionTrigger value="item-1">Arrow Icon</AccordionTrigger>
            <AccordionContent value="item-1">Uses an arrow that rotates when expanded.</AccordionContent>
          </AccordionItem>
        </Accordion>
      </div>

      <div className="w-[600px]">
        <h3 className="text-sm font-medium mb-2">Plus/Minus Icon</h3>
        <Accordion type="single" iconType="plus">
          <AccordionItem value="item-1">
            <AccordionTrigger value="item-1">Plus/Minus Icon</AccordionTrigger>
            <AccordionContent value="item-1">Shows plus when collapsed, minus when expanded.</AccordionContent>
          </AccordionItem>
        </Accordion>
      </div>
    </div>
  ),
};

export const Sizes: Story = {
  render: () => (
    <div className="space-y-8">
      <div className="w-[600px]">
        <h3 className="text-sm font-medium mb-2">Small</h3>
        <Accordion type="single" size="sm">
          <AccordionItem value="item-1">
            <AccordionTrigger value="item-1">Small Size</AccordionTrigger>
            <AccordionContent value="item-1">
              Compact accordion for limited space.
            </AccordionContent>
          </AccordionItem>
        </Accordion>
      </div>

      <div className="w-[600px]">
        <h3 className="text-sm font-medium mb-2">Medium (default)</h3>
        <Accordion type="single" size="md">
          <AccordionItem value="item-1">
            <AccordionTrigger value="item-1">Medium Size</AccordionTrigger>
            <AccordionContent value="item-1">
              Standard size for most use cases.
            </AccordionContent>
          </AccordionItem>
        </Accordion>
      </div>

      <div className="w-[600px]">
        <h3 className="text-sm font-medium mb-2">Large</h3>
        <Accordion type="single" size="lg">
          <AccordionItem value="item-1">
            <AccordionTrigger value="item-1">Large Size</AccordionTrigger>
            <AccordionContent value="item-1">
              Larger text for better visibility.
            </AccordionContent>
          </AccordionItem>
        </Accordion>
      </div>
    </div>
  ),
};

export const WithCustomIcons: Story = {
  render: () => (
    <div className="w-[600px]">
      <Accordion type="single" variant="bordered">
        <AccordionItem value="user">
          <AccordionTrigger value="user" icon={<User className="w-4 h-4" />}>
            User Settings
          </AccordionTrigger>
          <AccordionContent value="user">
            Manage your profile, preferences, and account details.
          </AccordionContent>
        </AccordionItem>
        <AccordionItem value="security">
          <AccordionTrigger value="security" icon={<Shield className="w-4 h-4" />}>
            Security & Privacy
          </AccordionTrigger>
          <AccordionContent value="security">
            Configure two-factor authentication, privacy settings, and security preferences.
          </AccordionContent>
        </AccordionItem>
        <AccordionItem value="notifications">
          <AccordionTrigger value="notifications" icon={<Bell className="w-4 h-4" />}>
            Notifications
          </AccordionTrigger>
          <AccordionContent value="notifications">
            Choose how and when you receive notifications.
          </AccordionContent>
        </AccordionItem>
        <AccordionItem value="billing">
          <AccordionTrigger value="billing" icon={<CreditCard className="w-4 h-4" />}>
            Billing & Subscription
          </AccordionTrigger>
          <AccordionContent value="billing">
            View your subscription details and manage payment methods.
          </AccordionContent>
        </AccordionItem>
      </Accordion>
    </div>
  ),
};

export const DisabledItems: Story = {
  render: () => (
    <div className="w-[600px]">
      <Accordion type="single">
        <AccordionItem value="item-1">
          <AccordionTrigger value="item-1">Available Section</AccordionTrigger>
          <AccordionContent value="item-1">
            This section can be expanded normally.
          </AccordionContent>
        </AccordionItem>
        <AccordionItem value="item-2">
          <AccordionTrigger value="item-2" disabled>
            Disabled Section
          </AccordionTrigger>
          <AccordionContent value="item-2">
            This content cannot be accessed.
          </AccordionContent>
        </AccordionItem>
        <AccordionItem value="item-3">
          <AccordionTrigger value="item-3">Another Available Section</AccordionTrigger>
          <AccordionContent value="item-3">
            This section works normally.
          </AccordionContent>
        </AccordionItem>
      </Accordion>
    </div>
  ),
};

export const NoAnimation: Story = {
  render: () => (
    <div className="w-[600px]">
      <Accordion type="single" animated={false}>
        <AccordionItem value="item-1">
          <AccordionTrigger value="item-1">No Animation</AccordionTrigger>
          <AccordionContent value="item-1">
            The accordion expands and collapses instantly without animation.
          </AccordionContent>
        </AccordionItem>
        <AccordionItem value="item-2">
          <AccordionTrigger value="item-2">Instant Toggle</AccordionTrigger>
          <AccordionContent value="item-2">
            Useful for users who prefer reduced motion.
          </AccordionContent>
        </AccordionItem>
      </Accordion>
    </div>
  ),
};

export const Controlled: Story = {
  render: function ControlledAccordion() {
    const [expandedItems, setExpandedItems] = React.useState<string[]>(['item-1']);

    return (
      <div className="w-[600px] space-y-4">
        <div className="flex gap-2">
          <button
            onClick={() => setExpandedItems(['item-1'])}
            className="px-3 py-1 bg-primary-500 text-white rounded hover:bg-primary-600"
          >
            Open Section 1
          </button>
          <button
            onClick={() => setExpandedItems(['item-2'])}
            className="px-3 py-1 bg-primary-500 text-white rounded hover:bg-primary-600"
          >
            Open Section 2
          </button>
          <button
            onClick={() => setExpandedItems(['item-3'])}
            className="px-3 py-1 bg-primary-500 text-white rounded hover:bg-primary-600"
          >
            Open Section 3
          </button>
          <button
            onClick={() => setExpandedItems([])}
            className="px-3 py-1 bg-gray-500 text-white rounded hover:bg-gray-600"
          >
            Close All
          </button>
        </div>
        <Accordion
          type="multiple"
          value={expandedItems}
          onValueChange={(value) => setExpandedItems(value as string[])}
        >
          <AccordionItem value="item-1">
            <AccordionTrigger value="item-1">Section 1</AccordionTrigger>
            <AccordionContent value="item-1">
              Content for section 1
            </AccordionContent>
          </AccordionItem>
          <AccordionItem value="item-2">
            <AccordionTrigger value="item-2">Section 2</AccordionTrigger>
            <AccordionContent value="item-2">
              Content for section 2
            </AccordionContent>
          </AccordionItem>
          <AccordionItem value="item-3">
            <AccordionTrigger value="item-3">Section 3</AccordionTrigger>
            <AccordionContent value="item-3">
              Content for section 3
            </AccordionContent>
          </AccordionItem>
        </Accordion>
        <div className="text-sm text-gray-600">
          Expanded items: <code className="bg-gray-100 px-2 py-1 rounded">{JSON.stringify(expandedItems)}</code>
        </div>
      </div>
    );
  },
};