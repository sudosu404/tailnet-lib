import { defineConfig, normalizePath } from `vite`;
import { resolve } from 'path';
import tailwindcss from '@tailwindcss/vite';
import { compression } from 'vite-plugin-compression2';
import { viteStaticCopy } from 'vite-plugin-static-copy';

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
