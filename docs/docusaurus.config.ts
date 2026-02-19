import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'Drift',
  tagline: 'Fast, open-source database migration tool',
  favicon: 'img/favicon.svg',

  future: {
    v4: true,
  },

  url: 'https://dunkinfrunkin.github.io',
  baseUrl: '/drift/',

  organizationName: 'dunkinfrunkin',
  projectName: 'drift',
  trailingSlash: false,

  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',

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
          editUrl: 'https://github.com/dunkinfrunkin/drift/edit/main/docs/',
          showLastUpdateTime: true,
          breadcrumbs: true,
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    colorMode: {
      defaultMode: 'dark',
      disableSwitch: false,
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'Drift',
      logo: {
        alt: 'Drift Logo',
        src: 'img/logo.svg',
      },
      hideOnScroll: false,
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docsSidebar',
          position: 'left',
          label: 'Docs',
        },
        {
          href: 'https://github.com/dunkinfrunkin/drift',
          position: 'right',
          className: 'header-github-link',
          'aria-label': 'GitHub repository',
        },
      ],
    },
    footer: {
      style: 'dark',
      links: [
        {
          title: 'Documentation',
          items: [
            {label: 'Getting Started', to: '/docs/getting-started'},
            {label: 'CLI Reference', to: '/docs/cli-reference'},
            {label: 'Configuration', to: '/docs/configuration'},
          ],
        },
        {
          title: 'Resources',
          items: [
            {label: 'Migration Files', to: '/docs/migration-files'},
            {label: 'Database Support', to: '/docs/databases'},
            {label: 'Linting', to: '/docs/linting'},
          ],
        },
        {
          title: 'Links',
          items: [
            {
              label: 'GitHub',
              href: 'https://github.com/dunkinfrunkin/drift',
            },
            {
              label: 'Releases',
              href: 'https://github.com/dunkinfrunkin/drift/releases',
            },
          ],
        },
      ],
      copyright: `Copyright \u00a9 ${new Date().getFullYear()} Drift Contributors. Built with Docusaurus.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['bash', 'yaml', 'json', 'go', 'sql', 'toml'],
    },
    tableOfContents: {
      minHeadingLevel: 2,
      maxHeadingLevel: 3,
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
