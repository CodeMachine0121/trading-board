import { fileURLToPath } from 'node:url'
import { defineVitestConfig } from '@nuxt/test-utils/config'

const appDirectory = fileURLToPath(new URL('./app', import.meta.url))
// 與 nuxt.config.ts 等價的 SCSS abstracts 注入。兩邊要一起改。
const scssAbstractsPath = fileURLToPath(new URL('./app/assets/styles/abstracts/_index.scss', import.meta.url))

export default defineVitestConfig({
  test: {
    // 所有測試都是純 TypeScript 或單一 SFC 的掛載，不需要 Nuxt runtime。
    environment: 'happy-dom',
    include: ['tests/**/*.spec.ts'],
    globals: true,
    coverage: {
      provider: 'v8',
      include: ['app/**/*.ts', 'app/**/*.vue'],
      // 組裝根、頁面、版面、中介層與應用程式的根只做接線。
      // server/ 的模擬板子只在 nuxt dev 時存在、不進正式映像檔，不列入覆蓋率門檻（另有冒煙測試）。
      exclude: ['app/plugins/**', 'app/pages/**', 'app/layouts/**', 'app/middleware/**', 'app/app.vue', 'app/**/*.d.ts'],
    },
  },
  css: {
    preprocessorOptions: {
      scss: {
        additionalData: `@use "${scssAbstractsPath}" as *;\n`,
      },
    },
  },
  resolve: {
    alias: {
      '~': appDirectory,
      '@': appDirectory,
    },
  },
})
