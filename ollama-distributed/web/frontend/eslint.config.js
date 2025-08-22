// For more info, see https://github.com/storybookjs/eslint-plugin-storybook#configuration-flat-config-format
import storybook from "eslint-plugin-storybook";

import globals from 'globals'
import pluginReact from 'eslint-plugin-react'
import pluginReactHooks from 'eslint-plugin-react-hooks'
import tseslint from '@typescript-eslint/eslint-plugin'
import tsParser from '@typescript-eslint/parser'
import jsxA11y from 'eslint-plugin-jsx-a11y'
import importPlugin from 'eslint-plugin-import'

export default [{
  files: ['**/*.{ts,tsx}'],
  ignores: ['dist/**', 'node_modules/**'],
  languageOptions: {
    parser: tsParser,
    globals: globals.browser,
    ecmaVersion: 'latest',
    sourceType: 'module',
  },
  plugins: {
    '@typescript-eslint': tseslint,
    react: pluginReact,
    'react-hooks': pluginReactHooks,
    'jsx-a11y': jsxA11y,
    import: importPlugin,
  },
  rules: {
    'react/react-in-jsx-scope': 'off',
    'react/prop-types': 'off',
    '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_', varsIgnorePattern: '^_' }],
    'import/order': ['error', { 'newlines-between': 'always', groups: [['builtin', 'external', 'internal']] }],
  },
  settings: {
    react: { version: 'detect' },
  },
}, ...storybook.configs["flat/recommended"]];

