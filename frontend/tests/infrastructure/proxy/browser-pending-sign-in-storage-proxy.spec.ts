import { afterEach, describe, expect, it, vi } from 'vitest'
import { PendingSignIn } from '~/domain/models/entities/pending-sign-in'
import { BrowserPendingSignInStorageProxy } from '~/infrastructure/proxy/browser-pending-sign-in-storage-proxy'

const KEY = 'trading-board:pending-sign-in'

afterEach(() => {
  sessionStorage.clear()
  vi.restoreAllMocks()
})

describe('BrowserPendingSignInStorageProxy：跨過轉址記住這一次登入，只用一次', () => {
  it('記下的密語與標記取得回來，取過一次就沒了', () => {
    const proxy = new BrowserPendingSignInStorageProxy()
    proxy.save(new PendingSignIn('verifier', 'state'))

    const taken = proxy.take()

    expect([taken?.codeVerifier, taken?.state]).toEqual(['verifier', 'state'])
    expect(proxy.take()).toBeNull()
  })

  it.each([
    ['什麼都沒記過', null],
    ['記的不是 JSON', 'oops'],
    ['形狀不對', JSON.stringify({ codeVerifier: 1, state: 'x' })],
  ])('%s → null', (_, stored) => {
    if (stored !== null) {
      sessionStorage.setItem(KEY, stored)
    }

    expect(new BrowserPendingSignInStorageProxy().take()).toBeNull()
  })

  it('瀏覽器不讓寫 → 不拋，回來時當作沒記過', () => {
    vi.spyOn(sessionStorage, 'setItem').mockImplementation(() => {
      throw new Error('denied')
    })
    vi.spyOn(console, 'warn').mockImplementation(() => {})
    const proxy = new BrowserPendingSignInStorageProxy()

    expect(() => proxy.save(new PendingSignIn('v', 's'))).not.toThrow()
    expect(proxy.take()).toBeNull()
  })
})
