import type { Tone } from '~/domain/models/vo/tone'

/** 「API 金鑰」清單上的一列，文字都已經排好。 */
export class ApiKeyDto {
  constructor(
    readonly id: string,
    readonly memberName: string,
    readonly note: string,
    readonly displayPrefixText: string,
    readonly statusLabel: string,
    readonly statusTone: Tone,
    readonly createdAtText: string,
    readonly lastUsedAtText: string,
    readonly canRevoke: boolean,
  ) {}
}
