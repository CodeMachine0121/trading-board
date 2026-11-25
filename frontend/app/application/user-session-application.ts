import type { UserSessionService } from '~/domain/service/user-session-service'
import type { CredentialsDto } from '~/domain/models/dto/credentials-dto'

/** Application：登入用例。純 TypeScript，全程只碰 DTO。 */
export class UserSessionApplication {
  constructor(private readonly userSessionService: UserSessionService) {}

  async signIn(credentialsDto: CredentialsDto): Promise<void> {
    return this.userSessionService.signIn(credentialsDto)
  }

  isSignedIn(): boolean {
    return this.userSessionService.isSignedIn()
  }

  signOut(): void {
    this.userSessionService.signOut()
  }
}
