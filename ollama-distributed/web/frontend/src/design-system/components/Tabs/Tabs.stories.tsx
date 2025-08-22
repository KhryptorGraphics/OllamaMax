import type { Meta, StoryObj } from '@storybook/react';
import { Tabs, TabsList, TabsTrigger, TabsContent } from './Tabs';
import { User, Settings, Bell, Shield, CreditCard, Activity } from 'lucide-react';

const meta = {
  title: 'Design System/Components/Tabs',
  component: Tabs,
  parameters: {
    layout: 'centered',
  },
  tags: ['autodocs'],
  argTypes: {
    orientation: {
      control: 'select',
      options: ['horizontal', 'vertical'],
    },
    variant: {
      control: 'select',
      options: ['line', 'pill', 'card'],
    },
    size: {
      control: 'select',
      options: ['sm', 'md', 'lg'],
    },
  },
} satisfies Meta<typeof Tabs>;

export default meta;
type Story = StoryObj<typeof meta>;

export const Default: Story = {
  render: () => (
    <div className="w-[600px]">
      <Tabs defaultValue="account">
        <TabsList>
          <TabsTrigger value="account">Account</TabsTrigger>
          <TabsTrigger value="security">Security</TabsTrigger>
          <TabsTrigger value="notifications">Notifications</TabsTrigger>
          <TabsTrigger value="billing">Billing</TabsTrigger>
        </TabsList>
        <TabsContent value="account">
          <div className="p-4 bg-white rounded-lg border">
            <h3 className="text-lg font-semibold mb-2">Account Settings</h3>
            <p className="text-gray-600">Manage your account settings and preferences.</p>
          </div>
        </TabsContent>
        <TabsContent value="security">
          <div className="p-4 bg-white rounded-lg border">
            <h3 className="text-lg font-semibold mb-2">Security Settings</h3>
            <p className="text-gray-600">Configure your security preferences and two-factor authentication.</p>
          </div>
        </TabsContent>
        <TabsContent value="notifications">
          <div className="p-4 bg-white rounded-lg border">
            <h3 className="text-lg font-semibold mb-2">Notification Preferences</h3>
            <p className="text-gray-600">Choose how and when you want to receive notifications.</p>
          </div>
        </TabsContent>
        <TabsContent value="billing">
          <div className="p-4 bg-white rounded-lg border">
            <h3 className="text-lg font-semibold mb-2">Billing Information</h3>
            <p className="text-gray-600">View and manage your billing details and subscription.</p>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  ),
};

export const LineVariant: Story = {
  render: () => (
    <div className="w-[600px]">
      <Tabs defaultValue="tab1" variant="line">
        <TabsList>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
          <TabsTrigger value="tab2">Tab 2</TabsTrigger>
          <TabsTrigger value="tab3">Tab 3</TabsTrigger>
          <TabsTrigger value="tab4" disabled>Disabled</TabsTrigger>
        </TabsList>
        <TabsContent value="tab1">
          <div className="p-4">Content for Tab 1</div>
        </TabsContent>
        <TabsContent value="tab2">
          <div className="p-4">Content for Tab 2</div>
        </TabsContent>
        <TabsContent value="tab3">
          <div className="p-4">Content for Tab 3</div>
        </TabsContent>
      </Tabs>
    </div>
  ),
};

export const PillVariant: Story = {
  render: () => (
    <div className="w-[600px]">
      <Tabs defaultValue="tab1" variant="pill">
        <TabsList>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
          <TabsTrigger value="tab2">Tab 2</TabsTrigger>
          <TabsTrigger value="tab3">Tab 3</TabsTrigger>
          <TabsTrigger value="tab4" disabled>Disabled</TabsTrigger>
        </TabsList>
        <TabsContent value="tab1">
          <div className="p-4">Content for Tab 1</div>
        </TabsContent>
        <TabsContent value="tab2">
          <div className="p-4">Content for Tab 2</div>
        </TabsContent>
        <TabsContent value="tab3">
          <div className="p-4">Content for Tab 3</div>
        </TabsContent>
      </Tabs>
    </div>
  ),
};

export const CardVariant: Story = {
  render: () => (
    <div className="w-[600px]">
      <Tabs defaultValue="tab1" variant="card">
        <TabsList>
          <TabsTrigger value="tab1">Tab 1</TabsTrigger>
          <TabsTrigger value="tab2">Tab 2</TabsTrigger>
          <TabsTrigger value="tab3">Tab 3</TabsTrigger>
          <TabsTrigger value="tab4" disabled>Disabled</TabsTrigger>
        </TabsList>
        <TabsContent value="tab1">
          <div className="p-4 -mt-4 bg-white border border-gray-200 rounded-b-lg rounded-tr-lg">
            Content for Tab 1
          </div>
        </TabsContent>
        <TabsContent value="tab2">
          <div className="p-4 -mt-4 bg-white border border-gray-200 rounded-b-lg rounded-tr-lg">
            Content for Tab 2
          </div>
        </TabsContent>
        <TabsContent value="tab3">
          <div className="p-4 -mt-4 bg-white border border-gray-200 rounded-b-lg rounded-tr-lg">
            Content for Tab 3
          </div>
        </TabsContent>
      </Tabs>
    </div>
  ),
};

export const WithIcons: Story = {
  render: () => (
    <div className="w-[600px]">
      <Tabs defaultValue="account" variant="line">
        <TabsList>
          <TabsTrigger value="account" icon={<User className="w-4 h-4" />}>
            Account
          </TabsTrigger>
          <TabsTrigger value="security" icon={<Shield className="w-4 h-4" />}>
            Security
          </TabsTrigger>
          <TabsTrigger value="notifications" icon={<Bell className="w-4 h-4" />}>
            Notifications
          </TabsTrigger>
          <TabsTrigger value="billing" icon={<CreditCard className="w-4 h-4" />}>
            Billing
          </TabsTrigger>
        </TabsList>
        <TabsContent value="account">
          <div className="p-4">Account settings content</div>
        </TabsContent>
        <TabsContent value="security">
          <div className="p-4">Security settings content</div>
        </TabsContent>
        <TabsContent value="notifications">
          <div className="p-4">Notification settings content</div>
        </TabsContent>
        <TabsContent value="billing">
          <div className="p-4">Billing settings content</div>
        </TabsContent>
      </Tabs>
    </div>
  ),
};

export const VerticalOrientation: Story = {
  render: () => (
    <div className="w-[600px] h-[400px]">
      <Tabs defaultValue="account" orientation="vertical">
        <TabsList>
          <TabsTrigger value="account" icon={<User className="w-4 h-4" />}>
            Account
          </TabsTrigger>
          <TabsTrigger value="security" icon={<Shield className="w-4 h-4" />}>
            Security
          </TabsTrigger>
          <TabsTrigger value="notifications" icon={<Bell className="w-4 h-4" />}>
            Notifications
          </TabsTrigger>
          <TabsTrigger value="billing" icon={<CreditCard className="w-4 h-4" />}>
            Billing
          </TabsTrigger>
          <TabsTrigger value="activity" icon={<Activity className="w-4 h-4" />}>
            Activity
          </TabsTrigger>
        </TabsList>
        <TabsContent value="account">
          <div className="p-4 bg-white rounded-lg border h-full">
            <h3 className="text-lg font-semibold mb-2">Account Settings</h3>
            <p className="text-gray-600">Manage your account settings and preferences.</p>
          </div>
        </TabsContent>
        <TabsContent value="security">
          <div className="p-4 bg-white rounded-lg border h-full">
            <h3 className="text-lg font-semibold mb-2">Security Settings</h3>
            <p className="text-gray-600">Configure your security preferences.</p>
          </div>
        </TabsContent>
        <TabsContent value="notifications">
          <div className="p-4 bg-white rounded-lg border h-full">
            <h3 className="text-lg font-semibold mb-2">Notification Preferences</h3>
            <p className="text-gray-600">Choose how you want to receive notifications.</p>
          </div>
        </TabsContent>
        <TabsContent value="billing">
          <div className="p-4 bg-white rounded-lg border h-full">
            <h3 className="text-lg font-semibold mb-2">Billing Information</h3>
            <p className="text-gray-600">Manage your billing details.</p>
          </div>
        </TabsContent>
        <TabsContent value="activity">
          <div className="p-4 bg-white rounded-lg border h-full">
            <h3 className="text-lg font-semibold mb-2">Activity Log</h3>
            <p className="text-gray-600">View your recent activity.</p>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  ),
};

export const Sizes: Story = {
  render: () => (
    <div className="space-y-8">
      <div className="w-[600px]">
        <h3 className="text-sm font-medium mb-2">Small</h3>
        <Tabs defaultValue="tab1" size="sm">
          <TabsList>
            <TabsTrigger value="tab1">Tab 1</TabsTrigger>
            <TabsTrigger value="tab2">Tab 2</TabsTrigger>
            <TabsTrigger value="tab3">Tab 3</TabsTrigger>
          </TabsList>
          <TabsContent value="tab1">Small size content</TabsContent>
          <TabsContent value="tab2">Small size content</TabsContent>
          <TabsContent value="tab3">Small size content</TabsContent>
        </Tabs>
      </div>
      <div className="w-[600px]">
        <h3 className="text-sm font-medium mb-2">Medium</h3>
        <Tabs defaultValue="tab1" size="md">
          <TabsList>
            <TabsTrigger value="tab1">Tab 1</TabsTrigger>
            <TabsTrigger value="tab2">Tab 2</TabsTrigger>
            <TabsTrigger value="tab3">Tab 3</TabsTrigger>
          </TabsList>
          <TabsContent value="tab1">Medium size content</TabsContent>
          <TabsContent value="tab2">Medium size content</TabsContent>
          <TabsContent value="tab3">Medium size content</TabsContent>
        </Tabs>
      </div>
      <div className="w-[600px]">
        <h3 className="text-sm font-medium mb-2">Large</h3>
        <Tabs defaultValue="tab1" size="lg">
          <TabsList>
            <TabsTrigger value="tab1">Tab 1</TabsTrigger>
            <TabsTrigger value="tab2">Tab 2</TabsTrigger>
            <TabsTrigger value="tab3">Tab 3</TabsTrigger>
          </TabsList>
          <TabsContent value="tab1">Large size content</TabsContent>
          <TabsContent value="tab2">Large size content</TabsContent>
          <TabsContent value="tab3">Large size content</TabsContent>
        </Tabs>
      </div>
    </div>
  ),
};

export const LazyLoading: Story = {
  render: () => (
    <div className="w-[600px]">
      <Tabs defaultValue="tab1">
        <TabsList>
          <TabsTrigger value="tab1">Immediate</TabsTrigger>
          <TabsTrigger value="tab2">Lazy Load 1</TabsTrigger>
          <TabsTrigger value="tab3">Lazy Load 2</TabsTrigger>
        </TabsList>
        <TabsContent value="tab1">
          <div className="p-4 bg-blue-50 rounded">
            This content is rendered immediately
          </div>
        </TabsContent>
        <TabsContent value="tab2" lazy>
          <div className="p-4 bg-green-50 rounded">
            This content is lazy loaded - only rendered when first activated
          </div>
        </TabsContent>
        <TabsContent value="tab3" lazy>
          <div className="p-4 bg-yellow-50 rounded">
            This content is also lazy loaded
          </div>
        </TabsContent>
      </Tabs>
    </div>
  ),
};

export const Controlled: Story = {
  render: function ControlledTabs() {
    const [activeTab, setActiveTab] = React.useState('tab2');

    return (
      <div className="w-[600px] space-y-4">
        <div className="flex gap-2">
          <button
            onClick={() => setActiveTab('tab1')}
            className="px-3 py-1 bg-primary-500 text-white rounded hover:bg-primary-600"
          >
            Go to Tab 1
          </button>
          <button
            onClick={() => setActiveTab('tab2')}
            className="px-3 py-1 bg-primary-500 text-white rounded hover:bg-primary-600"
          >
            Go to Tab 2
          </button>
          <button
            onClick={() => setActiveTab('tab3')}
            className="px-3 py-1 bg-primary-500 text-white rounded hover:bg-primary-600"
          >
            Go to Tab 3
          </button>
        </div>
        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList>
            <TabsTrigger value="tab1">Tab 1</TabsTrigger>
            <TabsTrigger value="tab2">Tab 2</TabsTrigger>
            <TabsTrigger value="tab3">Tab 3</TabsTrigger>
          </TabsList>
          <TabsContent value="tab1">
            <div className="p-4">Content for Tab 1</div>
          </TabsContent>
          <TabsContent value="tab2">
            <div className="p-4">Content for Tab 2</div>
          </TabsContent>
          <TabsContent value="tab3">
            <div className="p-4">Content for Tab 3</div>
          </TabsContent>
        </Tabs>
        <div className="text-sm text-gray-600">
          Active tab: <code className="bg-gray-100 px-2 py-1 rounded">{activeTab}</code>
        </div>
      </div>
    );
  },
};