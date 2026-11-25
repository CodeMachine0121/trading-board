import { createHash } from 'node:crypto'
import { describe, expect, it } from 'vitest'
import { WebCryptoSignInChallengeProxy } from '~/infrastructure/proxy/web-crypto-sign-in-challenge-proxy'

describe('WebCryptoSignInChallengeProxy：PKCE 的密語、證明與標記', () => {
  it('證明是密語的 SHA-256（base64url、無補齊）；密語與標記是 43 字的 base64url、每次都不同', async () => {
    const proxy = new WebCryptoSignInChallengeProxy()

    const first = await proxy.createChallenge()
    const second = await proxy.createChallenge()

    expect(first.codeChallenge).toBe(createHash('sha256').update(first.codeVerifier).digest('base64url'))
    expect(first.codeVerifier).toMatch(/^[\w-]{43}$/)
    expect(first.state).toMatch(/^[\w-]{43}$/)
    expect(second.codeVerifier).not.toBe(first.codeVerifier)
    expect(second.state).not.toBe(first.state)
  })
})
