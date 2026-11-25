import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import FillReportForm from '~/components/molecules/FillReportForm.vue'
import type { FillReportDto } from '~/domain/models/dto/fill-report-dto'
import { anOrderDto } from '../../fixtures/dtos'

function mountForm(submitting = false, errorMessage: string | null = null) {
  return mount(FillReportForm, { props: { order: anOrderDto('buy-1', true, 30), submitting, errorMessage } })
}

describe('FillReportForm', () => {
  it('初始值：限價、剩餘股數、手續費空白、現在時間，並提示剩餘股數', () => {
    const wrapper = mountForm()

    expect((wrapper.find('[data-testid="fill-price"]').element as HTMLInputElement).value).toBe('52.5')
    expect((wrapper.find('[data-testid="fill-shares"]').element as HTMLInputElement).value).toBe('30')
    expect((wrapper.find('[data-testid="fill-fee"]').element as HTMLInputElement).value).toBe('')
    expect((wrapper.find('[data-testid="fill-filled-at"]').element as HTMLInputElement).value).toBe('2026-11-25T20:30')
    expect(wrapper.text()).toContain('剩餘 30 股')
    expect(wrapper.text()).toContain('留空由系統估算')
  })

  it('送出 → 交出改過的每一格', async () => {
    const wrapper = mountForm()
    await wrapper.find('[data-testid="fill-price"]').setValue('52.4')
    await wrapper.find('[data-testid="fill-fee"]').setValue('3')
    await wrapper.find('[data-testid="fill-filled-at"]').setValue('2026-11-26T09:10')

    await wrapper.find('form').trigger('submit')

    const [fillReport] = wrapper.emitted('submit')?.[0] as [FillReportDto]
    expect([fillReport.priceText, fillReport.sharesText, fillReport.feeText, fillReport.filledAtText])
      .toEqual(['52.4', '30', '3', '2026-11-26T09:10'])
  })

  it('有錯誤 → 顯示理由', () => {
    expect(mountForm(false, '股數不可超過剩餘 30 股').find('[data-testid="fill-error"]').text()).toBe('股數不可超過剩餘 30 股')
  })

  it('送出中 → 送出鍵不能再按（避免重複回報）', () => {
    expect(mountForm(true).find('[data-testid="fill-submit"]').attributes('disabled')).toBeDefined()
  })

  it('取消 → 通知外面', async () => {
    const wrapper = mountForm()

    await wrapper.findAll('button').find(button => button.text() === '取消')?.trigger('click')

    expect(wrapper.emitted('cancel')).toHaveLength(1)
  })
})
