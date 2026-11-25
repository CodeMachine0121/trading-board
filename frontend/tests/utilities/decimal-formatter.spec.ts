import Decimal from 'decimal.js'
import { describe, expect, it } from 'vitest'
import { formatDecimal, formatPercent, formatSignedTwd, formatTwd } from '~/utilities/decimal-formatter'

describe('decimal-formatter', () => {
  it.each([
    ['4725', '4,725'],
    ['52.5', '52.5'],
    ['1234567.125', '1,234,567.125'],
    ['-800', '-800'],
    ['0', '0'],
  ])('formatDecimal(%s) → %s', (value, text) => {
    expect(formatDecimal(new Decimal(value))).toBe(text)
  })

  it.each([
    ['10000', 'NT$10,000'],
    ['-800', '-NT$800'],
    ['0', 'NT$0'],
  ])('formatTwd(%s) → %s', (value, text) => {
    expect(formatTwd(new Decimal(value))).toBe(text)
  })

  it.each([
    ['1200', '+NT$1,200'],
    ['-800', '-NT$800'],
    ['0', 'NT$0'],
  ])('formatSignedTwd(%s) → %s', (value, text) => {
    expect(formatSignedTwd(new Decimal(value))).toBe(text)
  })

  it.each([
    ['66.6666', '66.67%'],
    ['8.5', '8.50%'],
    ['-3.125', '-3.13%'],
  ])('formatPercent(%s) → %s', (value, text) => {
    expect(formatPercent(new Decimal(value))).toBe(text)
  })
})
