import 'styled-components';
import { Theme } from './theme';

// Extend the styled-components DefaultTheme interface
declare module 'styled-components' {
  export interface DefaultTheme extends Theme {}
}