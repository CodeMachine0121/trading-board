import { describe, expect, it } from 'vitest'
import { Session } from '~/domain/models/entities/session'

const NOW = new Date('2026-11-25T12:00:00Z')

describe('SessionDomain：一段登入還能不能用', () => {
  it.each([
    ['還沒到期 → 能用', 'token', '2026-11-25T12:00:01Z', true],
    ['到期的那一刻 → 不能用', 'token', '2026-11-25T12:00:00Z', false],
    ['已經過期 → 不能用', 'token', '2026-11-25T11:00:00Z', false],
    ['憑證是空的 → 不能用', '', '2026-11-26T12:00:00Z', false],
  ])('%s', (_, token, expiresAt, usable) => {
    expect(new Session(token, new Date(expiresAt)).toDomain().isUsableAt(NOW)).toBe(usable)
  })
})
