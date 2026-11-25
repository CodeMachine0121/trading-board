import type { IAuthProxy } from '~/domain/interface/i-auth-proxy'
import { Session } from '~/domain/models/entities/session'
import { BoardApiProxy } from '~/infrastructure/proxy/board-api-proxy'

/** 後端回傳的原始 wire 形狀，只存在於本檔內。 */
type AccessTokenWire = {
  accessToken: string
  expiresAt: string
}

/** Proxy：`POST /auth/login`。帳密不對（401）由共同出口翻成 CredentialsRejectedError。 */
export class AuthProxy extends BoardApiProxy implements IAuthProxy {
  async signIn(username: string, password: string): Promise<Session> {
    const accessTokenWire = await this.requestBoard<AccessTokenWire>('/auth/login', {
      method: 'POST',
      body: { username, password },
      authenticated: false,
    })

    return new Session(accessTokenWire.accessToken, new Date(accessTokenWire.expiresAt))
  }
}
