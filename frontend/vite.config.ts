import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import tailwindcss from '@tailwindcss/vite'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [svelte(), tailwindcss()],
  resolve: {
    alias: {
      // Handle .js imports pointing to .ts files in bindings
      '@/bindings': new URL('./bindings', import.meta.url).pathname,
    },
    extensions: ['.js', '.ts', '.svelte']
  }
})
