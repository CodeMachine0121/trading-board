import { BoardRejectedError } from '~/domain/errors/board-rejected-error'
import { BoardUnreachableError } from '~/domain/errors/board-unreachable-error'
import { CredentialsRejectedError } from '~/domain/errors/credentials-rejected-error'
import { SignInLockedError } from '~/domain/errors/sign-in-locked-error'
import { SignedOutError } from '~/domain/errors/signed-out-error'

/**
 * 板子回應失敗時的原始形狀。一般端點照契約 §4 回 `{ "error": { "code", "message" } }`；
 * 登入相關（OAuth）端點照 OAuth 回 `{ "error": "<code>", "error_description": "..." }`。
 * 與 wire 型別同理：描述的是外部契約，只存在於本檔內、不外流進 domain。
 */
type BoardFailure = {
  response?: { status?: number }
  data?: {
    error?: { code?: string, message?: string } | string
    error_description?: string
  }
}

/** 送出去的那一包。契約的請求內容都是一層的欄位。 */
export type BoardRequestBody = { readonly [key: string]: string | number }

export type BoardRequestOptions = {
  method?: 'GET' | 'POST' | 'PUT'
  query?: Record<string, string>
  body?: BoardRequestBody
  /** 以表單編碼送出（OAuth 的 token / revoke 端點只收這種）。 */
  form?: boolean
  /** 讓瀏覽器帶上（並收下）板子的 HttpOnly 續期 cookie。只有登入相關端點需要。 */
  withCredentials?: boolean
  /**
   * 這一發要不要帶著這個頁面的通行憑證。預設要。
   *
   * 登入相關端點不帶：它們被回 401 代表帳密不對，不是一段登入失效了——
   * 當成失效的話，會在登入畫面上把「密碼打錯」演成一次被登出。
   */
  authenticated?: boolean
}

/** 這一發該帶哪張通行憑證；沒有就回 `null`。由組裝根接到會續期的那一層。 */
export type AccessTokenSource = () => Promise<string | null>

const UNAUTHORIZED_STATUS = 401
const NOT_FOUND_STATUS = 404
const TOO_MANY_REQUESTS_STATUS = 429
/** 從這個狀態碼開始，問題出在板子自己身上，不是這次請求的內容。 */
const SERVER_ERROR_STATUS_FLOOR = 500
const FALLBACK_REJECTION_MESSAGE = '板子拒絕了這次操作'

/**
 * 所有打板子的 proxy 共用的請求執行與錯誤翻譯。
 *
 * 「板子以業務規則拒絕」「登入失效」「被鎖住」「連不上」對主人的意義完全不同（改輸入／重新登入／
 * 等一下／稍後再試），這條翻譯規則必須所有 proxy 一致，所以只寫在這裡一次。
 */
export abstract class BoardApiProxy {
  constructor(
    private readonly baseUrl: string,
    private readonly accessTokenSource: AccessTokenSource,
    /** 板子說這一段登入不算數時要做的事（忘掉登入、回登入畫面）。由組裝根決定。 */
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
    const authenticated = options.authenticated ?? true
    const accessToken = authenticated ? await this.accessTokenSource() : null

    return $fetch<TWire>(`${this.baseUrl}${path}`, {
      method: options.method ?? 'GET',
      query: options.query,
      body: options.form === true && options.body !== undefined
        ? new URLSearchParams(Object.entries(options.body).map(([name, value]) => [name, String(value)]))
        : options.body,
      headers: accessToken === null ? {} : { Authorization: `Bearer ${accessToken}` },
      ...(options.withCredentials === true ? { credentials: 'include' as const } : {}),
    })
  }

  private failureOf(error: unknown, options: BoardRequestOptions): Error {
    const failure = this.failureShapeOf(error)
    const status = failure.response?.status

    if (status === undefined || status >= SERVER_ERROR_STATUS_FLOOR) {
      return new BoardUnreachableError()
    }
    if (status === TOO_MANY_REQUESTS_STATUS) {
      return new SignInLockedError()
    }
    if (status === UNAUTHORIZED_STATUS) {
      if (options.authenticated === false) {
        return new CredentialsRejectedError()
      }

      this.onSignedOut()

      return new SignedOutError()
    }

    const errorField = failure.data?.error
    const message = typeof errorField === 'string' ? failure.data?.error_description : errorField?.message

    return new BoardRejectedError(message ?? FALLBACK_REJECTION_MESSAGE)
  }

  private failureShapeOf(error: unknown): BoardFailure {
    return typeof error === 'object' && error !== null ? error : {}
  }
}
