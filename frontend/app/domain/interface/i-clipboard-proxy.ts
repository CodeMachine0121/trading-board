/** 把一段文字放進剪貼簿。放不進去時拋 ClipboardUnavailableError。 */
export interface IClipboardProxy {
  copy(text: string): Promise<void>
}
