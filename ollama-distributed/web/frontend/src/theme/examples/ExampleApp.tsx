import React from 'react';
import styled from 'styled-components';
import { 
  ThemeProvider, 
  ThemeToggle, 
  useTheme, 
  useBreakpoint, 
  responsive, 
  space, 
  color, 
  shadow, 
  radius 
} from '../index';

// Example styled components using the theme system
const AppContainer = styled.div`
  min-height: 100vh;
  background-color: ${color('background.primary')};
  color: ${color('text.primary')};
  padding: ${space(4)};
  transition: all ${({ theme }) => theme.animation.duration.normal} ${({ theme }) => theme.animation.easing['ease-out']};
`;

const Header = styled.header`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: ${space(8)};
  padding: ${space(4)} 0;
  border-bottom: 1px solid ${color('border.default')};
`;

const Title = styled.h1`
  font-size: ${({ theme }) => theme.typography.fontSize['3xl'][0]};
  line-height: ${({ theme }) => theme.typography.fontSize['3xl'][1].lineHeight};
  font-weight: ${({ theme }) => theme.typography.fontWeight.bold};
  color: ${color('text.primary')};
  margin: 0;
  
  ${responsive.sm(`
    font-size: ${({ theme }: any) => theme.typography.fontSize['4xl'][0]};
  `)}
`;

const Card = styled.div`
  background-color: ${color('surface.elevated')};
  border: 1px solid ${color('border.default')};
  border-radius: ${radius('lg')};
  padding: ${space(6)};
  margin-bottom: ${space(4)};
  box-shadow: ${shadow('sm')};
  transition: all ${({ theme }) => theme.animation.duration.fast} ${({ theme }) => theme.animation.easing['ease-out']};
  
  &:hover {
    box-shadow: ${shadow('md')};
    border-color: ${color('border.strong')};
  }
`;

const Grid = styled.div`
  display: grid;
  gap: ${space(4)};
  grid-template-columns: 1fr;
  
  ${responsive.sm('grid-template-columns: repeat(2, 1fr);')}
  ${responsive.lg('grid-template-columns: repeat(3, 1fr);')}
`;

const Button = styled.button<{ variant?: 'primary' | 'secondary' | 'ghost' }>`
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: ${space(2)} ${space(4)};
  border-radius: ${radius('md')};
  font-size: ${({ theme }) => theme.typography.fontSize.sm[0]};
  font-weight: ${({ theme }) => theme.typography.fontWeight.medium};
  cursor: pointer;
  transition: all ${({ theme }) => theme.animation.duration.fast} ${({ theme }) => theme.animation.easing['ease-out']};
  border: 1px solid transparent;
  
  ${({ variant = 'primary', theme }) => {
    const buttonColors = theme.colors.button[variant];
    return `
      background-color: ${buttonColors.background};
      color: ${buttonColors.text};
      border-color: ${buttonColors.border};
      
      &:hover {
        background-color: ${buttonColors.backgroundHover};
      }
      
      &:active {
        background-color: ${buttonColors.backgroundActive};
        transform: scale(0.98);
      }
    `;
  }}
  
  &:focus {
    outline: 2px solid ${color('border.focus')};
    outline-offset: 2px;
  }
  
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
`;

const ColorPalette = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(60px, 1fr));
  gap: ${space(2)};
  margin-top: ${space(4)};
`;

const ColorSwatch = styled.div<{ color: string }>`
  width: 60px;
  height: 60px;
  border-radius: ${radius('md')};
  background-color: ${({ color }) => color};
  border: 1px solid ${color('border.default')};
  box-shadow: ${shadow('xs')};
`;

const ResponsiveText = styled.p`
  font-size: ${({ theme }) => theme.typography.fontSize.sm[0]};
  color: ${color('text.secondary')};
  margin-top: ${space(2)};
  
  ${responsive.sm(`
    font-size: ${({ theme }: any) => theme.typography.fontSize.base[0]};
  `)}
  
  ${responsive.lg(`
    font-size: ${({ theme }: any) => theme.typography.fontSize.lg[0]};
  `)}
`;

const ExampleComponent: React.FC = () => {
  const theme = useTheme();
  const breakpoint = useBreakpoint();
  
  return (
    <Card>
      <h3>Theme Information</h3>
      <p><strong>Current theme:</strong> {theme.mode}</p>
      <p><strong>Current breakpoint:</strong> {breakpoint.current}</p>
      <p><strong>Is mobile:</strong> {breakpoint.isMobile ? 'Yes' : 'No'}</p>
      <p><strong>Is desktop:</strong> {breakpoint.isDesktop ? 'Yes' : 'No'}</p>
      
      <div style={{ marginTop: space(4)({ theme }) }}>
        <h4>Primary Colors</h4>
        <ColorPalette>
          {Object.entries(theme.raw.colors.primary).map(([shade, color]) => (
            <div key={shade}>
              <ColorSwatch color={color} />
              <small style={{ fontSize: '12px', color: theme.colors.text.muted }}>
                {shade}
              </small>
            </div>
          ))}
        </ColorPalette>
      </div>
      
      <div style={{ marginTop: space(4)({ theme }) }}>
        <h4>Button Variants</h4>
        <div style={{ display: 'flex', gap: space(2)({ theme }), marginTop: space(2)({ theme }) }}>
          <Button variant="primary">Primary</Button>
          <Button variant="secondary">Secondary</Button>
          <Button variant="ghost">Ghost</Button>
        </div>
      </div>
      
      <ResponsiveText>
        This text changes size based on screen size: {breakpoint.current}
      </ResponsiveText>
    </Card>
  );
};

const ExampleApp: React.FC = () => {
  return (
    <ThemeProvider defaultTheme="system">
      <AppContainer>
        <Header>
          <Title>OllamaMax Theme System</Title>
          <ThemeToggle variant="dropdown" size="md" />
        </Header>
        
        <Grid>
          <ExampleComponent />
          
          <Card>
            <h3>Typography Scale</h3>
            <h1 style={{ margin: '8px 0' }}>Heading 1</h1>
            <h2 style={{ margin: '8px 0' }}>Heading 2</h2>
            <h3 style={{ margin: '8px 0' }}>Heading 3</h3>
            <h4 style={{ margin: '8px 0' }}>Heading 4</h4>
            <h5 style={{ margin: '8px 0' }}>Heading 5</h5>
            <h6 style={{ margin: '8px 0' }}>Heading 6</h6>
            <p>Body text with proper line height and spacing.</p>
            <small style={{ color: 'var(--color-text-muted)' }}>Small text</small>
          </Card>
          
          <Card>
            <h3>Elevation Levels</h3>
            {[0, 1, 2, 3, 4, 5].map(level => (
              <div
                key={level}
                style={{
                  padding: '16px',
                  margin: '8px 0',
                  borderRadius: '8px',
                  backgroundColor: 'var(--color-surface-elevated)',
                  boxShadow: level === 0 ? 'none' : `var(--shadow-${level})`,
                  border: level === 0 ? '1px solid var(--color-border-default)' : 'none'
                }}
              >
                Elevation {level}
              </div>
            ))}
          </Card>
        </Grid>
      </AppContainer>
    </ThemeProvider>
  );
};

export default ExampleApp;