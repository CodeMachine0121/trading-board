import type { IAuthProxy } from '~/domain/interface/i-auth-proxy'
import type { ISessionStorageProxy } from '~/domain/interface/i-session-storage-proxy'
import type { CredentialsDto } from '~/domain/models/dto/credentials-dto'

/** 主人的登入：換一段登入並記住、問它還能不能用、登出。 */
export class UserSessionService {
  constructor(
    private readonly authProxy: IAuthProxy,
    private readonly sessionStorageProxy: ISessionStorageProxy,
  ) {}

  async signIn(credentialsDto: CredentialsDto): Promise<void> {
    const session = await this.authProxy.signIn(credentialsDto.username, credentialsDto.password)
    this.sessionStorageProxy.writeSession(session)
  }

  isSignedIn(): boolean {
    const session = this.sessionStorageProxy.readSession()

    return session !== null && session.toDomain().isUsableAt(new Date())
  }

  signOut(): void {
    this.sessionStorageProxy.clearSession()
  }
}
