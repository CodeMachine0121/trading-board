import { ref } from 'vue'
import type { UserSessionApplication } from '~/application/user-session-application'
import type { CredentialsDto } from '~/domain/models/dto/credentials-dto'
import type { SignInReturnDto } from '~/domain/models/dto/sign-in-return-dto'
import { failureMessageOf } from '~/composables/failure-message'

/** 換頁的方式由呼叫端給（Nuxt 的 navigateTo）；external 表示整頁跳去另一個網址。 */
export type Navigate = (to: string, options?: { external?: boolean, replace?: boolean }) => unknown

/** 登入畫面、登入後的首頁、與每一頁的「沒登入就回登入畫面」都看這幾個路徑。 */
export const HOME_PATH = '/'
export const LOGIN_PATH = '/login'
export const SIGN_IN_CALLBACK_PATH = '/auth/callback'

/**
 * 登入的每一段「接下來看到哪一頁」：送出帳密、轉址回來、每次換頁的把關、登出。
 * 規則只寫在這裡；頁面、版面與中介層只把 Nuxt 的 navigateTo 交進來。
 */
export function useSignInFlow(userSessionApplication: UserSessionApplication, navigate: Navigate) {
  const submitting = ref(false)
  const errorMessage = ref<string | null>(null)

  /** 帳密對了就整頁跳去板子回的網址（帶一次性授權碼）；錯了留在登入畫面說明原因。 */
  async function signIn(credentialsDto: CredentialsDto): Promise<void> {
    submitting.value = true
    errorMessage.value = null

    try {
      await navigate(await userSessionApplication.beginSignIn(credentialsDto), { external: true })
    }
    catch (error: unknown) {
      errorMessage.value = failureMessageOf(error)
      submitting.value = false
    }
  }

  /** 轉址回來：成功進首頁；失敗留在這頁說明，並給一條回登入畫面的路。 */
  async function completeSignIn(signInReturnDto: SignInReturnDto): Promise<void> {
    try {
      await userSessionApplication.completeSignIn(signInReturnDto)
      await navigate(HOME_PATH, { replace: true })
    }
    catch (error: unknown) {
      errorMessage.value = failureMessageOf(error)
    }
  }

  /** 每次換頁前：沒登入（也續不到）只能看登入畫面；已登入不必再看登入畫面。回要改去的路徑，不必改就是 undefined。 */
  async function guard(path: string): Promise<string | undefined> {
    if (path === SIGN_IN_CALLBACK_PATH) {
      return undefined
    }

    const signedIn = await userSessionApplication.restoreSession()
    if (!signedIn && path !== LOGIN_PATH) {
      return LOGIN_PATH
    }
    if (signedIn && path === LOGIN_PATH) {
      return HOME_PATH
    }

    return undefined
  }

  /** 登出在畫面上一定成功：結束之後回登入畫面。 */
  async function signOut(): Promise<void> {
    await userSessionApplication.signOut()
    await navigate(LOGIN_PATH)
  }

  return { submitting, errorMessage, signIn, completeSignIn, guard, signOut }
}
