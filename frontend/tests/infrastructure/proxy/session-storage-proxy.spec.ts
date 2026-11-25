import { afterEach, describe, expect, it, vi } from 'vitest'
import { Session } from '~/domain/models/entities/session'
import { SessionStorageProxy } from '~/infrastructure/proxy/session-storage-proxy'

const KEY = 'trading-board:session'

afterEach(() => {
  vi.unstubAllGlobals()
  localStorage.clear()
})

describe('SessionStorageProxy', () => {
  it('寫進去的登入讀得回來', () => {
    const proxy = new SessionStorageProxy()

    proxy.writeSession(new Session('abc', new Date('2026-11-26T00:00:00Z')))
    const session = proxy.readSession()

    expect(session?.accessToken).toBe('abc')
    expect(session?.expiresAt.toISOString()).toBe('2026-11-26T00:00:00.000Z')
  })

  it('登出之後讀不到', () => {
    const proxy = new SessionStorageProxy()
    proxy.writeSession(new Session('abc', new Date('2026-11-26T00:00:00Z')))

    proxy.clearSession()

    expect(proxy.readSession()).toBeNull()
  })

  it.each([
    ['什麼都沒記', null],
    ['不是 JSON', 'not json'],
    ['少了憑證', JSON.stringify({ expiresAt: '2026-11-26T00:00:00Z' })],
    ['到期時刻不是字串', JSON.stringify({ accessToken: 'abc', expiresAt: 1 })],
    ['到期時刻讀不出來', JSON.stringify({ accessToken: 'abc', expiresAt: 'someday' })],
  ])('%s → 當作沒登入過', (_, stored) => {
    if (stored !== null) {
      localStorage.setItem(KEY, stored)
    }

    expect(new SessionStorageProxy().readSession()).toBeNull()
  })

  it('瀏覽器把儲存關掉時，三個方法都不拋', () => {
    const blocked = () => {
      throw new Error('SecurityError')
    }
    vi.stubGlobal('localStorage', { getItem: blocked, setItem: blocked, removeItem: blocked })
    const proxy = new SessionStorageProxy()

    expect(() => proxy.writeSession(new Session('abc', new Date()))).not.toThrow()
    expect(proxy.readSession()).toBeNull()
    expect(() => proxy.clearSession()).not.toThrow()
  })
})
