import { useSignInFlow } from '~/composables/use-sign-in-flow'

/**
 * 把關：每次換頁前問 useSignInFlow 該不該改去別頁（規則與測試都在那裡）。
 * 全域中介層而不是每一頁自己判斷——一頁忘了寫就是一個洞。
 */
export default defineNuxtRouteMiddleware(async (to) => {
  const redirectTo = await useSignInFlow(useNuxtApp().$userSessionApplication, navigateTo).guard(to.path)
  if (redirectTo !== undefined) {
    return navigateTo(redirectTo)
  }
})
