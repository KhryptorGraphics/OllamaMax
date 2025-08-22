import React from 'react';
import styled from 'styled-components';
import { useThemeContext } from '../hooks';
import { Sun, Moon, Monitor } from 'lucide-react';

interface ThemeToggleProps {
  variant?: 'icon' | 'dropdown';
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

const ToggleButton = styled.button<{ size: string }>`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: ${({ theme, size }) => {
    switch (size) {
      case 'sm': return `${theme.spacing[1]} ${theme.spacing[2]}`;
      case 'lg': return `${theme.spacing[3]} ${theme.spacing[4]}`;
      default: return `${theme.spacing[2]} ${theme.spacing[3]}`;
    }
  }};
  border: 1px solid ${({ theme }) => theme.colors.border.default};
  border-radius: ${({ theme }) => theme.radii.md};
  background-color: ${({ theme }) => theme.colors.surface.elevated};
  color: ${({ theme }) => theme.colors.text.secondary};
  font-size: ${({ theme, size }) => {
    switch (size) {
      case 'sm': return theme.typography.fontSize.sm[0];
      case 'lg': return theme.typography.fontSize.lg[0];
      default: return theme.typography.fontSize.base[0];
    }
  }};
  cursor: pointer;
  transition: all ${({ theme }) => theme.animation.duration.fast} ${({ theme }) => theme.animation.easing['ease-out']};
  
  &:hover {
    background-color: ${({ theme }) => theme.colors.surface.muted};
    color: ${({ theme }) => theme.colors.text.primary};
    border-color: ${({ theme }) => theme.colors.border.strong};
  }
  
  &:focus {
    outline: 2px solid ${({ theme }) => theme.colors.border.focus};
    outline-offset: 2px;
  }
  
  &:active {
    transform: scale(0.98);
  }
  
  svg {
    width: ${({ size }) => {
      switch (size) {
        case 'sm': return '16px';
        case 'lg': return '24px';
        default: return '20px';
      }
    }};
    height: ${({ size }) => {
      switch (size) {
        case 'sm': return '16px';
        case 'lg': return '24px';
        default: return '20px';
      }
    }};
  }
`;

const DropdownContainer = styled.div`
  position: relative;
  display: inline-block;
`;

const DropdownMenu = styled.div<{ isOpen: boolean }>`
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: ${({ theme }) => theme.spacing[1]};
  min-width: 150px;
  background-color: ${({ theme }) => theme.colors.surface.elevated};
  border: 1px solid ${({ theme }) => theme.colors.border.default};
  border-radius: ${({ theme }) => theme.radii.lg};
  box-shadow: ${({ theme }) => theme.shadows.lg};
  opacity: ${({ isOpen }) => (isOpen ? 1 : 0)};
  visibility: ${({ isOpen }) => (isOpen ? 'visible' : 'hidden')};
  transform: ${({ isOpen }) => isOpen ? 'translateY(0)' : 'translateY(-8px)'};
  transition: all ${({ theme }) => theme.animation.duration.fast} ${({ theme }) => theme.animation.easing['ease-out']};
  z-index: ${({ theme }) => theme.zIndex.dropdown};
`;

const DropdownItem = styled.button<{ isActive: boolean }>`
  display: flex;
  align-items: center;
  width: 100%;
  padding: ${({ theme }) => `${theme.spacing[2]} ${theme.spacing[3]}`};
  text-align: left;
  font-size: ${({ theme }) => theme.typography.fontSize.sm[0]};
  color: ${({ theme }) => theme.colors.text.secondary};
  background-color: ${({ theme, isActive }) => 
    isActive ? theme.colors.surface.muted : 'transparent'
  };
  border: none;
  cursor: pointer;
  transition: all ${({ theme }) => theme.animation.duration.fast} ${({ theme }) => theme.animation.easing['ease-out']};
  
  &:first-child {
    border-top-left-radius: ${({ theme }) => theme.radii.lg};
    border-top-right-radius: ${({ theme }) => theme.radii.lg};
  }
  
  &:last-child {
    border-bottom-left-radius: ${({ theme }) => theme.radii.lg};
    border-bottom-right-radius: ${({ theme }) => theme.radii.lg};
  }
  
  &:hover {
    background-color: ${({ theme }) => theme.colors.surface.muted};
    color: ${({ theme }) => theme.colors.text.primary};
  }
  
  &:focus {
    outline: 2px solid ${({ theme }) => theme.colors.border.focus};
    outline-offset: -2px;
  }
  
  svg {
    width: 16px;
    height: 16px;
    margin-right: ${({ theme }) => theme.spacing[2]};
  }
`;

const IconThemeToggle: React.FC<ThemeToggleProps> = ({ size = 'md', className }) => {
  const { mode, toggleTheme } = useThemeContext();
  
  const getIcon = () => {
    switch (mode) {
      case 'dark':
        return <Moon />;
      case 'light':
        return <Sun />;
      case 'system':
        return <Monitor />;
      default:
        return <Sun />;
    }
  };
  
  return (
    <ToggleButton
      onClick={toggleTheme}
      size={size}
      className={className}
      aria-label={`Switch to ${mode === 'dark' ? 'light' : 'dark'} theme`}
      title={`Current theme: ${mode}`}
    >
      {getIcon()}
    </ToggleButton>
  );
};

const DropdownThemeToggle: React.FC<ThemeToggleProps> = ({ size = 'md', className }) => {
  const { mode, setTheme } = useThemeContext();
  const [isOpen, setIsOpen] = React.useState(false);
  
  const options = [
    { value: 'light' as const, label: 'Light', icon: Sun },
    { value: 'dark' as const, label: 'Dark', icon: Moon },
    { value: 'system' as const, label: 'System', icon: Monitor }
  ];
  
  const currentOption = options.find(option => option.value === mode);
  
  const handleSelect = (value: typeof mode) => {
    setTheme(value);
    setIsOpen(false);
  };
  
  // Close dropdown when clicking outside
  React.useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      const target = event.target as Element;
      if (!target.closest('[data-dropdown]')) {
        setIsOpen(false);
      }
    };
    
    if (isOpen) {
      document.addEventListener('click', handleClickOutside);
      return () => document.removeEventListener('click', handleClickOutside);
    }
  }, [isOpen]);
  
  return (
    <DropdownContainer data-dropdown className={className}>
      <ToggleButton
        onClick={() => setIsOpen(!isOpen)}
        size={size}
        aria-expanded={isOpen}
        aria-haspopup="true"
        aria-label="Theme selector"
      >
        {currentOption && <currentOption.icon />}
        <span style={{ marginLeft: '8px' }}>{currentOption?.label}</span>
      </ToggleButton>
      
      <DropdownMenu isOpen={isOpen}>
        {options.map((option) => (
          <DropdownItem
            key={option.value}
            onClick={() => handleSelect(option.value)}
            isActive={mode === option.value}
            aria-pressed={mode === option.value}
          >
            <option.icon />
            {option.label}
          </DropdownItem>
        ))}
      </DropdownMenu>
    </DropdownContainer>
  );
};

export const ThemeToggle: React.FC<ThemeToggleProps> = ({ 
  variant = 'icon',
  ...props 
}) => {
  if (variant === 'dropdown') {
    return <DropdownThemeToggle {...props} />;
  }
  
  return <IconThemeToggle {...props} />;
};