import {
  defineConfig,
  minimal2023Preset as preset,
} from '@vite-pwa/assets-generator/config'

export default defineConfig({
  headLinkOptions: {
    preset: 'all',
  },
  preset,
  images: ['public/tailnet.svg'],
})
