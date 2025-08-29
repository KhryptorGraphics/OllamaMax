import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

// This runs in Node.js - Don't use client-side code here (browser APIs, JSX...)

const config: Config = {
  title: 'Ollama Distributed',
  tagline: 'Enterprise-grade distributed AI model inference platform',
  favicon: 'img/favicon.ico',

  // Future flags, see https://docusaurus.io/docs/api/docusaurus-config#future
  future: {
    v4: true, // Improve compatibility with the upcoming Docusaurus v4
  },

  // Set the production url of your site here
  url: 'https://docs.ollama-distributed.example.com',
  // Set the /<baseUrl>/ pathname under which your site is served
  // For GitHub pages deployment, it is often '/<projectName>/'
  baseUrl: '/',

  // GitHub pages deployment config.
  // If you aren't using GitHub pages, you don't need these.
  organizationName: 'ollama', // Usually your GitHub org/user name.
  projectName: 'ollama-distributed', // Usually your repo name.

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',

  // Even if you don't use internationalization, you can use this field to set
  // useful metadata like html lang. For example, if your site is Chinese, you
  // may want to replace "en" with "zh-Hans".
  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl:
            'https://github.com/facebook/docusaurus/tree/main/packages/create-docusaurus/templates/shared/',
        },
        blog: {
          showReadingTime: true,
          feedOptions: {
            type: ['rss', 'atom'],
            xslt: true,
          },
          // Please change this to your repo.
          // Remove this to remove the "edit this page" links.
          editUrl:
            'https://github.com/facebook/docusaurus/tree/main/packages/create-docusaurus/templates/shared/',
          // Useful options to enforce blogging best practices
          onInlineTags: 'warn',
          onInlineAuthors: 'warn',
          onUntruncatedBlogPosts: 'warn',
        },
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    // Replace with your project's social card
    image: 'img/ollama-distributed-social-card.jpg',
    navbar: {
      title: 'Ollama Distributed',
      logo: {
        alt: 'Ollama Distributed Logo',
        src: 'img/logo.svg',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'userGuideSidebar',
          position: 'left',
          label: 'User Guide',
        },
        {
          type: 'docSidebar',
          sidebarId: 'developerGuideSidebar',
          position: 'left',
          label: 'Developer',
        },
        {
          type: 'docSidebar',
          sidebarId: 'operationsGuideSidebar',
          position: 'left',
          label: 'Operations',
        },
        {
          type: 'docSidebar',
          sidebarId: 'apiSidebar',
          position: 'left',
          label: 'API',
        },
        {to: '/api-playground', label: 'API Playground', position: 'left'},
        {to: '/training', label: 'Training', position: 'left'},
        {
          href: 'https://github.com/ollama/ollama-distributed',
          label: 'GitHub',
          position: 'right',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Documentation',
          items: [
            {
              label: 'User Guide',
              to: '/docs/user-guide/getting-started',
            },
            {
              label: 'Developer Guide',
              to: '/docs/developer-guide/architecture',
            },
            {
              label: 'API Reference',
              to: '/docs/api/reference',
            },
            {
              label: 'Operations Guide',
              to: '/docs/operations-guide/deployment',
            },
          ],
        },
        {
          title: 'Tools',
          items: [
            {
              label: 'API Playground',
              to: '/api-playground',
            },
            {
              label: 'Training Materials',
              to: '/training',
            },
            {
              label: 'Troubleshooting',
              to: '/docs/troubleshooting',
            },
          ],
        },
        {
          title: 'Community',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/ollama/ollama-distributed',
            },
            {
              label: 'Discord',
              href: 'https://discord.gg/ollama',
            },
            {
              label: 'Discussions',
              href: 'https://github.com/ollama/ollama-distributed/discussions',
            },
          ],
        },
        {
          title: 'Resources',
          items: [
            {
              label: 'Examples',
              href: 'https://github.com/ollama/ollama-distributed/tree/main/examples',
            },
            {
              label: 'Performance Benchmarks',
              to: '/docs/performance',
            },
            {
              label: 'Security Guide',
              to: '/docs/security',
            },
          ],
        },
      ],
      copyright: `Copyright © ${new Date().getFullYear()} Ollama Distributed Team. Built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
