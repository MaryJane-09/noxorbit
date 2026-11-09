import js from '@eslint/js'
import globals from 'globals'
import reactHooks from 'eslint-plugin-react-hooks'
import reactRefresh from 'eslint-plugin-react-refresh'
import { defineConfig, globalIgnores } from 'eslint/config'

// ESLint 9 does not count JSX component tags as references to imports.
// Mark only names actually used as JSX elements; keep no-unused-vars active
// for genuinely unused imports (including unused component names).
const jsxUsesVars = {
  meta: { type: 'problem', schema: [] },
  create(context) {
    return {
      JSXOpeningElement(node) {
        let tag = node.name
        // For <Icons.Lock />, the referenced variable is "Icons".
        while (tag.type === 'JSXMemberExpression') tag = tag.object
        if (tag.type === 'JSXIdentifier' && /^[A-Z]/.test(tag.name)) {
          context.sourceCode.markVariableAsUsed(tag.name, tag)
        }
      },
    }
  },
}

export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.{js,jsx}'],
    extends: [
      js.configs.recommended,
      reactHooks.configs.flat.recommended,
      reactRefresh.configs.vite,
    ],
    languageOptions: {
      globals: globals.browser,
      parserOptions: { ecmaFeatures: { jsx: true } },
    },
    plugins: {
      local: { rules: { 'jsx-uses-vars': jsxUsesVars } },
    },
    rules: {
      'local/jsx-uses-vars': 'error',
    },
  },
])
