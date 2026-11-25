import type { UserSessionService } from '~/domain/service/user-session-service'
import type { CredentialsDto } from '~/domain/models/dto/credentials-dto'
import type { OwnerDto } from '~/domain/models/dto/owner-dto'
import type { SignInReturnDto } from '~/domain/models/dto/sign-in-return-dto'

/** Application：登入、轉址回來、續期、登出、我是誰。純 TypeScript，全程只碰 DTO。 */
export class UserSessionApplication {
  constructor(private readonly userSessionService: UserSessionService) {}

  /** 回板子要畫面轉去的網址。 */
  async beginSignIn(credentialsDto: CredentialsDto): Promise<string> {
    return this.userSessionService.beginSignIn(credentialsDto)
  }

  async completeSignIn(signInReturnDto: SignInReturnDto): Promise<void> {
    return this.userSessionService.completeSignIn(signInReturnDto)
  }

  async restoreSession(): Promise<boolean> {
    return this.userSessionService.restoreSession()
  }

  async accessToken(): Promise<string | null> {
    return this.userSessionService.accessToken()
  }

  async signOut(): Promise<void> {
    return this.userSessionService.signOut()
  }

  forgetSession(): void {
    this.userSessionService.forgetSession()
  }

  async fetchCurrentOwner(): Promise<OwnerDto> {
    return this.userSessionService.fetchCurrentOwner()
  }
}
