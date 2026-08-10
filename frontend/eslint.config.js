import js from '@eslint/js';
import globals from 'globals';
import reactHooks from 'eslint-plugin-react-hooks';
import reactRefresh from 'eslint-plugin-react-refresh';
import reactX from 'eslint-plugin-react-x';
import reactDom from 'eslint-plugin-react-dom';
import simpleImportSort from 'eslint-plugin-simple-import-sort';
import tseslint from 'typescript-eslint';
import prettierConfig from 'eslint-config-prettier';

import { defineConfig, globalIgnores } from 'eslint/config';

export default defineConfig([
  globalIgnores(['dist', 'coverage']),

  {
    files: ['**/*.{ts,tsx}'],

    plugins: {
      'simple-import-sort': simpleImportSort,
    },

    extends: [
      js.configs.recommended,

      tseslint.configs.recommended,

      reactX.configs['recommended-typescript'],

      reactDom.configs.recommended,

      reactHooks.configs.flat.recommended,

      reactRefresh.configs.vite,

      prettierConfig,
    ],

    languageOptions: {
      globals: globals.browser,
    },

    rules: {
      'react-x/no-missing-key': 'error',

      'simple-import-sort/imports': [
        'error',
        {
          groups: [
            ['^\\u0000'],

            ['^node:'],

            ['^react$', '^react/'],

            ['^@?\\w'],

            ['^@/'],

            ['^\\.\\.'],

            ['^\\.'],

            ['^.+\\.(css|scss|sass|less)$'],
          ],
        },
      ],

      'simple-import-sort/exports': 'error',
    },
  },
]);
