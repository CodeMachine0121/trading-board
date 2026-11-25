import type { Position } from '~/domain/models/entities/position'

/** 板子算出來的持股：還持有的，以及某個月出清的。 */
export interface IPositionProxy {
  listOpenPositions(): Promise<Position[]>
  listClosedPositions(month: string): Promise<Position[]>
}
