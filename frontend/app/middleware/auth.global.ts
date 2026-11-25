import { LOGIN_PATH, SIGN_IN_CALLBACK_PATH } from '~/plugins/dependencies'

/** 不用登入就看得到的頁面：登入畫面，以及板子把登入結果轉回來的那一頁。 */
const PUBLIC_PATHS = [LOGIN_PATH, SIGN_IN_CALLBACK_PATH]

/**
 * 把關：沒登入就只看得到登入畫面；已登入的人不必再看到登入畫面。
 * 頁面剛打開時手上沒有通行憑證，先用瀏覽器帶著的續期 cookie 試著找回登入。
 * 全域中介層而不是每一頁自己判斷——一頁忘了寫就是一個洞。
 */
export default defineNuxtRouteMiddleware(async (to) => {
  if (to.path === SIGN_IN_CALLBACK_PATH) {
    return
  }

  const { $userSessionApplication } = useNuxtApp()
  const signedIn = await $userSessionApplication.restoreSession()
  const goingToPublicPage = PUBLIC_PATHS.includes(to.path)

  if (!signedIn && !goingToPublicPage) {
    return navigateTo(LOGIN_PATH)
  }
  if (signedIn && to.path === LOGIN_PATH) {
    return navigateTo('/')
  }
})
