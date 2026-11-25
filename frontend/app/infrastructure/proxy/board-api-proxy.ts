import type { ISessionStorageProxy } from '~/domain/interface/i-session-storage-proxy'
import { BoardRejectedError } from '~/domain/errors/board-rejected-error'
import { BoardUnreachableError } from '~/domain/errors/board-unreachable-error'
import { CredentialsRejectedError } from '~/domain/errors/credentials-rejected-error'
import { SignedOutError } from '~/domain/errors/signed-out-error'

/**
 * 板子回應失敗時的原始形狀（契約 §4：`{ "error": { "code", "message" } }`）。
 * 與 wire 型別同理：描述的是外部契約，只存在於本檔內、不外流進 domain。
 */
type BoardFailure = {
  response?: { status?: number }
  data?: { error?: { code?: string, message?: string } }
}

/** 送出去的那一包。契約的請求內容都是一層的欄位。 */
export type BoardRequestBody = { readonly [key: string]: string | number }

export type BoardRequestOptions = {
  method?: 'GET' | 'POST' | 'PUT'
  query?: Record<string, string>
  body?: BoardRequestBody
  /**
   * 這一發要不要帶著這台瀏覽器的登入。預設要。
   *
   * 唯一的例外是登入本身：那一發被回 401 代表帳密不對，不是一段登入失效了——
   * 當成失效的話，會在登入畫面上把「密碼打錯」演成一次被登出。
   */
  authenticated?: boolean
}

const UNAUTHORIZED_STATUS = 401
const NOT_FOUND_STATUS = 404
/** 從這個狀態碼開始，問題出在板子自己身上，不是這次請求的內容。 */
const SERVER_ERROR_STATUS_FLOOR = 500
const FALLBACK_REJECTION_MESSAGE = '板子拒絕了這次操作'

/**
 * 所有打板子的 proxy 共用的請求執行與錯誤翻譯。
 *
 * 「板子以業務規則拒絕」「登入失效」「連不上」對主人的意義完全不同（改輸入／重新登入／
 * 稍後再試），這條翻譯規則必須所有 proxy 一致，所以只寫在這裡一次。
 */
export abstract class BoardApiProxy {
  constructor(
    private readonly baseUrl: string,
    private readonly sessionStorageProxy: ISessionStorageProxy,
    /** 板子說這一段登入不算數時要做的事（清掉登入、回登入畫面）。由組裝根決定。 */
    private readonly onSignedOut: () => void = () => {},
  ) {}

  protected async requestBoard<TWire>(path: string, options: BoardRequestOptions = {}): Promise<TWire> {
    try {
      return await this.fetchBoard<TWire>(path, options)
    }
    catch (error: unknown) {
      throw this.failureOf(error, options)
    }
  }

  /** 同上，但「找不到」是一個正常的答案（例如還沒有任何交易卡）。 */
  protected async requestBoardOrNull<TWire>(path: string, options: BoardRequestOptions = {}): Promise<TWire | null> {
    try {
      return await this.fetchBoard<TWire>(path, options)
    }
    catch (error: unknown) {
      if (this.failureShapeOf(error).response?.status === NOT_FOUND_STATUS) {
        return null
      }

      throw this.failureOf(error, options)
    }
  }

  private async fetchBoard<TWire>(path: string, options: BoardRequestOptions): Promise<TWire> {
    const accessToken = this.sessionStorageProxy.readSession()?.accessToken
    const authenticated = options.authenticated ?? true

    return $fetch<TWire>(`${this.baseUrl}${path}`, {
      method: options.method ?? 'GET',
      query: options.query,
      body: options.body,
      headers: authenticated && accessToken !== undefined ? { Authorization: `Bearer ${accessToken}` } : {},
    })
  }

  private failureOf(error: unknown, options: BoardRequestOptions): Error {
    const failure = this.failureShapeOf(error)
    const status = failure.response?.status

    if (status === undefined || status >= SERVER_ERROR_STATUS_FLOOR) {
      return new BoardUnreachableError()
    }
    if (status === UNAUTHORIZED_STATUS) {
      if (options.authenticated === false) {
        return new CredentialsRejectedError()
      }

      this.onSignedOut()

      return new SignedOutError()
    }

    return new BoardRejectedError(failure.data?.error?.message ?? FALLBACK_REJECTION_MESSAGE)
  }

  private failureShapeOf(error: unknown): BoardFailure {
    return typeof error === 'object' && error !== null ? error : {}
  }
}
