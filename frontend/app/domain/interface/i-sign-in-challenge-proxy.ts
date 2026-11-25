import type { SignInChallenge } from '~/domain/models/entities/sign-in-challenge'

/** 為一次登入抽出密語與標記，並算出密語的 S256 證明（需要亂數與雜湊，所以在邊界外）。 */
export interface ISignInChallengeProxy {
  createChallenge(): Promise<SignInChallenge>
}
