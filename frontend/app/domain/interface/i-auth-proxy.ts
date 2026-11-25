import type { Session } from '~/domain/models/entities/session'

/** 向板子換一段登入。帳密不對時拋 CredentialsRejectedError。 */
export interface IAuthProxy {
  signIn(username: string, password: string): Promise<Session>
}
