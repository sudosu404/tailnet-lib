import { defineConfig, normalizePath } from `vite`;
import { resolve } from 'path';
import tailwindcss from '@tailwindcss/vite';
import { compression } from 'vite-plugin-compression2';
import { viteStaticCopy } from 'vite-plugin-static-copy';
import { VitePWA } from 'vite-plugin-pwa'

export default defineConfig({

  server: {
    proxy: {
      '/list': 'http://localhost:8080',
      '/stream': 'http://localhost:8080'
    }
  },
  plugins: [
    tailwindcss(),

    compression(),
    compression({ algorithm: 'brotliCompress' }),

    VitePWA({
      registerType: 'autoUpdate',
      injectRegister: false,

      pwaAssets: {
        disabled: false,
        config: true,
      },

      manifest: {
        name: 'TSDProxy',
        short_name: 'TSDProxy',
        description: 'TSDProxy',
        theme_color: '#ffffff',
      },

      workbox: {
        globPatterns: ['**/*.{js,css,html,ico}'],
        cleanupOutdatedCaches: true,
        clientsClaim: true,
      },

      devOptions: {
        enabled: false,
        navigateFallback: 'index.html',
        suppressWarnings: true,
        type: 'module',
      },
    }),


    viteStaticCopy({
      targets: [
        {
          src: normalizePath(resolve(__dirname, 'node_modules/simple-icons/icons/*')),
          dest: 'icons/si'
        },
        {
          src: normalizePath(resolve(__dirname, 'node_modules/@mdi/svg/svg/*')),
          dest: 'icons/mdi'
        },
        {
          src: normalizePath(resolve(__dirname, 'public/icons/sh/*')),
          dest: 'icons/sh'
        },
        {
          src: normalizePath(resolve(__dirname, 'public/icons/*')),
          dest: 'icons'
        }

      ]
    }),

  ],
  build: {
    rollupOptions: {
      output: {
        entryFileNames: `[name]-[hash].js`,
        chunkFileNames: `[name]-[hash].js`,
        assetFileNames: `[name]-[hash].[ext]`
      }
    }
  }
});
