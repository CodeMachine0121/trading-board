import { fileURLToPath } from 'node:url'

// SCSS 的 abstracts（token map、存取函式、mixin）自動注入到每一個 .scss 與
// <style lang="scss"> 區塊，元件因此不必逐檔 @use 就能用 color() / spacing() / respond-to()。
// 注意：vitest.config.ts 有一份等價設定，兩邊要一起改（見 .claude/rules/component-design.md）。
const scssAbstractsPath = fileURLToPath(new URL('./app/assets/styles/abstracts/_index.scss', import.meta.url))

export default defineNuxtConfig({
  modules: ['@nuxt/eslint'],
  $development: {
    // dev 冷啟動時 #app-manifest 的 race（見 go-trading-frontend README）。
    experimental: {
      appManifest: false,
    },
  },
  // 只在瀏覽器裡畫：部署是 `nuxt generate` 出來的一包靜態檔交給 nginx，
  // 伺服器端不知道你是誰，先算一份只會先畫出一片空白再跳去登入。
  ssr: false,
  // 原子化設計的資料夾只表示層級，不進元件名字。
  components: [{ path: '~/components', pathPrefix: false }],
  devtools: { enabled: false },
  app: {
    head: {
      htmlAttrs: { lang: 'zh-Hant-TW' },
      title: 'Trading Board',
      meta: [
        { name: 'viewport', content: 'width=device-width, initial-scale=1, viewport-fit=cover' },
      ],
    },
  },
  css: ['~/assets/styles/main.scss'],
  runtimeConfig: {
    public: {
      // 板子 REST API 的 base URL（含 /api/v1）；以 NUXT_PUBLIC_API_BASE_URL 覆寫。
      //
      // 預設指向 server/routes/mock/ 的開發用模擬後端——它擋在網路層，
      // 程式碼裡沒有任何假實作，所以換成真的後端只是改這一個設定。
      // 靜態輸出時這個值會被寫進 bundle：正式環境由 Dockerfile 的 build 參數決定。
      apiBaseUrl: '/mock/api/v1',
    },
  },
  srcDir: 'app/',
  // 模擬後端住在專案根目錄的 server/（不在 app/ 底下，它不屬於前端的分層）。
  serverDir: 'server/',
  devServer: { port: 3000 },
  compatibilityDate: '2026-08-30',
  vite: {
    css: {
      preprocessorOptions: {
        scss: {
          additionalData: `@use "${scssAbstractsPath}" as *;\n`,
        },
      },
    },
  },
  typescript: {
    strict: true,
    typeCheck: false,
  },
  eslint: {
    config: {
      stylistic: true,
    },
  },
})
