import { MonthSummaryApplication } from '~/application/month-summary-application'
import { PositionApplication } from '~/application/position-application'
import { TradeCardApplication } from '~/application/trade-card-application'
import { ApiKeyApplication } from '~/application/api-key-application'
import { UserSessionApplication } from '~/application/user-session-application'
import { OAuthClientVo } from '~/domain/models/vo/oauth-client-vo'
import { ApiKeyService } from '~/domain/service/api-key-service'
import { MonthSummaryService } from '~/domain/service/month-summary-service'
import { PositionService } from '~/domain/service/position-service'
import { TradeCardService } from '~/domain/service/trade-card-service'
import { UserSessionService } from '~/domain/service/user-session-service'
import { ApiKeyProxy } from '~/infrastructure/proxy/api-key-proxy'
import { AuthProxy } from '~/infrastructure/proxy/auth-proxy'
import { BrowserClipboardProxy } from '~/infrastructure/proxy/browser-clipboard-proxy'
import { BrowserPendingSignInStorageProxy } from '~/infrastructure/proxy/browser-pending-sign-in-storage-proxy'
import { MemorySessionStorageProxy } from '~/infrastructure/proxy/memory-session-storage-proxy'
import { MonthSummaryProxy } from '~/infrastructure/proxy/month-summary-proxy'
import { OrderProxy } from '~/infrastructure/proxy/order-proxy'
import { PositionProxy } from '~/infrastructure/proxy/position-proxy'
import { TradeCardProxy } from '~/infrastructure/proxy/trade-card-proxy'
import { LOGIN_PATH, SIGN_IN_CALLBACK_PATH } from '~/composables/use-sign-in-flow'
import { WebCryptoSignInChallengeProxy } from '~/infrastructure/proxy/web-crypto-sign-in-challenge-proxy'

/**
 * 組裝根：唯一知道所有具體型別的地方。
 * 由外而內組裝 proxy → domain service → application，並把 application provide 給元件層：
 * `const { $tradeCardApplication } = useNuxtApp()`。
 */
export default defineNuxtPlugin(() => {
  // 打真的板子還是打開發用的模擬後端，只看這一個設定（NUXT_PUBLIC_API_BASE_URL）。
  const runtimeConfig = useRuntimeConfig().public
  const apiBaseUrl = runtimeConfig.apiBaseUrl
  const oauthClient = new OAuthClientVo(runtimeConfig.oauthClientId, `${window.location.origin}${SIGN_IN_CALLBACK_PATH}`)

  // 短效通行憑證只活在這個頁面的記憶體裡；每一發請求都經由 userSessionApplication 拿它（快過期會先續期）。
  const sessionStorageProxy = new MemorySessionStorageProxy()
  const accessTokenSource = (): Promise<string | null> => userSessionApplication.accessToken()

  // 板子說這一段登入不算數了：忘掉它、回登入畫面。發請求的那一層不懂導頁，所以在這裡接線。
  const onSignedOut = (): void => {
    userSessionApplication.forgetSession()
    void navigateTo(LOGIN_PATH)
  }

  const userSessionApplication: UserSessionApplication = new UserSessionApplication(
    new UserSessionService(
      new AuthProxy(apiBaseUrl, accessTokenSource, onSignedOut, oauthClient),
      sessionStorageProxy,
      new BrowserPendingSignInStorageProxy(),
      new WebCryptoSignInChallengeProxy(),
    ),
  )

  const tradeCardApplication = new TradeCardApplication(
    new TradeCardService(
      new TradeCardProxy(apiBaseUrl, accessTokenSource, onSignedOut),
      new OrderProxy(apiBaseUrl, accessTokenSource, onSignedOut),
    ),
  )

  const positionApplication = new PositionApplication(
    new PositionService(new PositionProxy(apiBaseUrl, accessTokenSource, onSignedOut)),
  )

  const monthSummaryApplication = new MonthSummaryApplication(
    new MonthSummaryService(new MonthSummaryProxy(apiBaseUrl, accessTokenSource, onSignedOut)),
  )

  const apiKeyApplication = new ApiKeyApplication(
    new ApiKeyService(new ApiKeyProxy(apiBaseUrl, accessTokenSource, onSignedOut)),
    new BrowserClipboardProxy(),
  )

  return {
    provide: {
      userSessionApplication,
      tradeCardApplication,
      positionApplication,
      monthSummaryApplication,
      apiKeyApplication,
    },
  }
})
