/** 板子記錄的委託狀態。「已過期」不在這裡——它只是畫面上的說法，見 OrderProgressDomain。 */
export type OrderStatus = 'pending' | 'partiallyFilled' | 'filled' | 'skipped'
