import { fileURLToPath, URL } from 'node:url'

import vue from '@vitejs/plugin-vue'
import tailwindcss from '@tailwindcss/vite'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { defineConfig, loadEnv } from 'vite'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const proxyTarget = env.VITE_DEV_PROXY_TARGET || 'http://127.0.0.1:8080'
  const wsProxyTarget = proxyTarget.replace(/^http/i, 'ws')

  return {
    plugins: [
      vue(),
      tailwindcss(),
      AutoImport({
        imports: ['vue', 'vue-router', 'pinia'],
        dts: 'src/auto-imports.d.ts',
      }),
      Components({
        dts: 'src/components.d.ts',
      }),
    ],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url)),
      },
    },
    server: {
      host: '0.0.0.0',
      port: 5173,
      proxy: {
        '/api': { target: proxyTarget, changeOrigin: true },
        '/ws': { target: wsProxyTarget, ws: true, changeOrigin: true },
      },
    },
    build: {
      rollupOptions: {
        output: {
          manualChunks(id) {
            if (id.includes('/vue-echarts/')) {
              return 'vue-echarts'
            }

            if (id.includes('/echarts/charts/')) {
              return 'echarts-charts'
            }

            if (id.includes('/echarts/components/') || id.includes('/echarts/features/')) {
              return 'echarts-components'
            }

            if (id.includes('/echarts/renderers/') || id.includes('/echarts/lib/renderer/')) {
              return 'echarts-renderers'
            }

            if (id.includes('/zrender/')) {
              return 'zrender'
            }

            if (
              id.includes('/echarts/core/') ||
              id.includes('/echarts/lib/core/') ||
              id.includes('/echarts/lib/chart/') ||
              id.includes('/echarts/lib/component/') ||
              id.includes('/echarts/lib/visual/') ||
              id.includes('/echarts/lib/layout/') ||
              id.includes('/echarts/lib/util/') ||
              id.includes('/echarts/lib/coord/') ||
              id.includes('/echarts/lib/label/') ||
              id.includes('/echarts/lib/animation/') ||
              id.includes('/echarts/lib/data/') ||
              id.includes('/echarts/lib/model/') ||
              id.includes('/echarts/lib/view/') ||
              id.includes('/echarts/lib/export/')
            ) {
              return 'echarts-runtime'
            }

            if (id.includes('/echarts/')) {
              return 'echarts-runtime'
            }

            if (id.includes('/vue-router/') || id.includes('/pinia/') || id.includes('/@vueuse/core/')) {
              return 'vue-ecosystem'
            }

            if (id.includes('/axios/')) {
              return 'network'
            }

            if (id.includes('/node_modules/vue/')) {
              return 'vue-core'
            }
          },
        },
      },
    },
  }
})
