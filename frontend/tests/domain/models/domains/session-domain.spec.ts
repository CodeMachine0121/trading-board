import { describe, expect, it } from 'vitest'
import { Session } from '~/domain/models/entities/session'

const NOW = new Date('2026-11-25T12:00:00Z')

describe('SessionDomain：手上的通行憑證該不該先換新', () => {
  it.each([
    ['還有 10 分鐘 → 直接用', 'token', '2026-11-25T12:10:00Z', false],
    ['剛好還有 1 分鐘 → 直接用', 'token', '2026-11-25T12:01:00Z', false],
    ['只剩 59.999 秒 → 先換新', 'token', '2026-11-25T12:00:59.999Z', true],
    ['已經過期 → 先換新', 'token', '2026-11-25T11:00:00Z', true],
    ['憑證是空的 → 先換新', '', '2026-11-26T12:00:00Z', true],
  ])('%s', (_, token, expiresAt, needsRenewal) => {
    expect(new Session(token, new Date(expiresAt)).toDomain().needsRenewalAt(NOW)).toBe(needsRenewal)
  })
})
