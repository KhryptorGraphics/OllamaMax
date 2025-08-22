# Models Management Components

This directory contains components for managing AI models in the distributed Ollama system.

## Components

### ModelCard
- **Purpose**: Grid view display of individual models
- **Features**: 
  - Model information display (name, size, family, status)
  - Sync progress indicators
  - Quick action buttons (view, copy, delete)
  - Click to view details

### ModelDetailPanel  
- **Purpose**: Slide-out panel with comprehensive model details
- **Features**:
  - Tabbed interface (Overview, Distribution, Usage, Versions)
  - Real-time sync status
  - Node distribution visualization
  - Usage statistics and analytics
  - Version history with rollback capability
  - Full model actions (download, copy, sync, delete)

### ModelPullDialog
- **Purpose**: Modal for downloading new models from registry
- **Features**:
  - Model name and tag input
  - Popular models suggestions with tags
  - Real-time download progress
  - Search functionality for models
  - Error handling and validation

### SimpleSelect
- **Purpose**: Basic select dropdown component
- **Features**:
  - Customizable options
  - Value change callbacks
  - Styled to match design system

## Usage

```tsx
import { ModelCard, ModelDetailPanel, ModelPullDialog } from '@/components/models'

// In your component
<ModelCard 
  model={modelData}
  onAction={handleAction}
  onSelect={handleSelect}
/>

<ModelDetailPanel
  model={selectedModel}
  onClose={handleClose}
  onAction={handleAction}
/>

<ModelPullDialog
  isOpen={showDialog}
  onClose={handleClose}
  onSuccess={handleSuccess}
/>
```

## Integration

These components integrate with:
- **Design System**: Uses design system components (Card, Button, Badge, etc.)
- **API Client**: ModelsAPI for data operations
- **WebSocket**: Real-time updates for sync status
- **Table Component**: Works with the Sprint B Table component for list view

## Features

- **Responsive Design**: Works on desktop and mobile
- **Real-time Updates**: WebSocket integration for live status updates
- **Accessibility**: WCAG AA compliant with proper ARIA labels
- **Performance**: Optimized rendering with proper React patterns
- **Error Handling**: Comprehensive error states and recovery
- **Offline Support**: Graceful degradation when offline