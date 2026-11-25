import type { IAuthProxy } from '~/domain/interface/i-auth-proxy'
import type { OAuthClientVo } from '~/domain/models/vo/oauth-client-vo'
import { Owner } from '~/domain/models/entities/owner'
import { Session } from '~/domain/models/entities/session'
import { BoardApiProxy } from '~/infrastructure/proxy/board-api-proxy'
import type { AccessTokenSource } from '~/infrastructure/proxy/board-api-proxy'

/** 後端回傳的原始 wire 形狀，只存在於本檔內。 */
type AuthorizationRedirectWire = {
  redirectTo: string
}

type TokenWire = {
  access_token: string
  token_type: string
  expires_in: number
}

type OwnerWire = {
  email: string
}

/**
 * Proxy：板子的 `oauth/authorize`、`oauth/token`、`oauth/revoke` 與 `me`。
 *
 * 換授權碼、續期、登出都讓瀏覽器帶上（並收下）板子的 HttpOnly 續期 cookie；
 * 續期憑證本身從不經過這裡的程式碼。
 */
export class AuthProxy extends BoardApiProxy implements IAuthProxy {
  constructor(
    baseUrl: string,
    accessTokenSource: AccessTokenSource,
    onSignedOut: () => void,
    private readonly clientSettings: OAuthClientVo,
  ) {
    super(baseUrl, accessTokenSource, onSignedOut)
  }

  async authorize(email: string, password: string, codeChallenge: string, state: string): Promise<string> {
    const redirectWire = await this.requestBoard<AuthorizationRedirectWire>('/oauth/authorize', {
      method: 'POST',
      body: {
        email,
        password,
        response_type: 'code',
        client_id: this.clientSettings.clientId,
        redirect_uri: this.clientSettings.redirectUri,
        code_challenge: codeChallenge,
        code_challenge_method: 'S256',
        state,
      },
      authenticated: false,
    })

    return redirectWire.redirectTo
  }

  async exchangeCode(code: string, codeVerifier: string): Promise<Session> {
    return this.requestToken({
      grant_type: 'authorization_code',
      code,
      code_verifier: codeVerifier,
      client_id: this.clientSettings.clientId,
      redirect_uri: this.clientSettings.redirectUri,
    })
  }

  async refresh(): Promise<Session> {
    return this.requestToken({ grant_type: 'refresh_token', client_id: this.clientSettings.clientId })
  }

  async revoke(): Promise<void> {
    await this.requestBoard<unknown>('/oauth/revoke', {
      method: 'POST',
      body: { client_id: this.clientSettings.clientId },
      form: true,
      withCredentials: true,
      authenticated: false,
    })
  }

  async fetchCurrentOwner(): Promise<Owner> {
    const ownerWire = await this.requestBoard<OwnerWire>('/me')

    return new Owner(ownerWire.email)
  }

  /** 兩種 grant 回的是同一種東西：一張短效通行憑證（續期 cookie 由瀏覽器自己收下）。 */
  private async requestToken(form: Record<string, string>): Promise<Session> {
    const requestedAt = new Date()
    const tokenWire = await this.requestBoard<TokenWire>('/oauth/token', {
      method: 'POST',
      body: form,
      form: true,
      withCredentials: true,
      authenticated: false,
    })

    return new Session(tokenWire.access_token, new Date(requestedAt.getTime() + tokenWire.expires_in * 1000))
  }
}
