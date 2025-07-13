import js from '@eslint/js';
import svelte from 'eslint-plugin-svelte';
import globals from 'globals';
import ts from 'typescript-eslint';
import svelteConfig from './svelte.config.js';

export default ts.config(
  js.configs.recommended,
  ...ts.configs.recommended,
  ...svelte.configs.recommended,
  {
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.node
      }
    }
  },
  {
    files: ['**/*.svelte', '**/*.svelte.ts', '**/*.svelte.js'],
    languageOptions: {
      parserOptions: {
        projectService: true,
        extraFileExtensions: ['.svelte'],
        parser: ts.parser,
        svelteConfig
      }
    }
  },
  {
    rules: {
      // Override or add rule settings here
      '@typescript-eslint/no-explicit-any': 'warn', // Make any warnings instead of errors
      '@typescript-eslint/no-unused-vars': 'warn', // Make unused vars warnings
      'svelte/require-each-key': 'warn' // Make missing each keys warnings
    }
  },
  {
    ignores: ['build/', 'dist/', '.svelte-kit/', 'node_modules/', 'bindings/']
  }
);