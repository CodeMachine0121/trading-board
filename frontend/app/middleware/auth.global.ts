import { LOGIN_PATH } from '~/plugins/dependencies'

/**
 * 把關：沒登入（或登入已過期）就只看得到登入畫面；已登入的人不必再看到登入畫面。
 * 全域中介層而不是每一頁自己判斷——一頁忘了寫就是一個洞。
 */
export default defineNuxtRouteMiddleware((to) => {
  const { $userSessionApplication } = useNuxtApp()
  const signedIn = $userSessionApplication.isSignedIn()
  const goingToLogin = to.path === LOGIN_PATH

  if (!signedIn && !goingToLogin) {
    return navigateTo(LOGIN_PATH)
  }
  if (signedIn && goingToLogin) {
    return navigateTo('/')
  }
})
