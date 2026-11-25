import { describe, expect, it } from 'vitest'
import { Session } from '~/domain/models/entities/session'
import { MemorySessionStorageProxy } from '~/infrastructure/proxy/memory-session-storage-proxy'

describe('MemorySessionStorageProxy：通行憑證只在這個頁面的記憶體裡', () => {
  it('記住、讀回、忘掉，且從不寫進瀏覽器儲存', () => {
    const proxy = new MemorySessionStorageProxy()
    const session = new Session('at', new Date('2026-11-25T12:15:00Z'))

    expect(proxy.readSession()).toBeNull()
    proxy.writeSession(session)
    expect(proxy.readSession()).toBe(session)
    expect(localStorage.length + sessionStorage.length).toBe(0)
    proxy.clearSession()
    expect(proxy.readSession()).toBeNull()
  })
})
