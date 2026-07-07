import { createConfigForNuxt } from '@nuxt/eslint-config/flat'

export default createConfigForNuxt(
  {
    features: {
      typescript: true,
    },
  },
  {
    files: ['app/components/Breadcrumbs.vue'],
    rules: {
      'vue/multi-word-component-names': 'off',
    },
  },
  {
    files: ['app/pages/**/*.vue'],
    rules: {
      'vue/no-v-html': 'off',
    },
  },
  {
    rules: {
      'vue/html-self-closing': [
        'error',
        {
          html: { void: 'always', normal: 'always', component: 'always' },
          svg: 'always',
          math: 'always',
        },
      ],
    },
  },
)
