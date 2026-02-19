import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docsSidebar: [
    'getting-started',
    'installation',
    'configuration',
    'migration-files',
    {
      type: 'category',
      label: 'CLI Reference',
      collapsed: false,
      items: [
        'cli-reference',
        'cli/migrate',
        'cli/undo',
        'cli/info',
        'cli/validate',
        'cli/diff',
        'cli/snapshot',
        'cli/lint',
        'cli/baseline',
        'cli/repair',
        'cli/clean',
        'cli/squash',
        'cli/report',
        'cli/ui',
        'cli/init',
      ],
    },
    'databases',
    'linting',
    'web-ui',
    'building-from-source',
  ],
};

export default sidebars;
