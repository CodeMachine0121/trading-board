import type { PendingSignIn } from '~/domain/models/entities/pending-sign-in'

/**
 * 跨過轉址還要記得的那一次登入（只在這個分頁）。兩個方法都不拋：記不住的話，
 * 回來時就當作來回失效，請主人重新登入。
 */
export interface IPendingSignInStorageProxy {
  save(pendingSignIn: PendingSignIn): void
  /** 取出並忘掉——一次登入只能用一次。 */
  take(): PendingSignIn | null
}
