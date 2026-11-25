import { MonthSummaryApplication } from '~/application/month-summary-application'
import { PositionApplication } from '~/application/position-application'
import { TradeCardApplication } from '~/application/trade-card-application'
import { UserSessionApplication } from '~/application/user-session-application'
import { MonthSummaryService } from '~/domain/service/month-summary-service'
import { PositionService } from '~/domain/service/position-service'
import { TradeCardService } from '~/domain/service/trade-card-service'
import { UserSessionService } from '~/domain/service/user-session-service'
import { AuthProxy } from '~/infrastructure/proxy/auth-proxy'
import { MonthSummaryProxy } from '~/infrastructure/proxy/month-summary-proxy'
import { OrderProxy } from '~/infrastructure/proxy/order-proxy'
import { PositionProxy } from '~/infrastructure/proxy/position-proxy'
import { SessionStorageProxy } from '~/infrastructure/proxy/session-storage-proxy'
import { TradeCardProxy } from '~/infrastructure/proxy/trade-card-proxy'

/** 登入畫面的路徑。中介層與被登出時的導頁都看這一個。 */
export const LOGIN_PATH = '/login'

/**
 * 組裝根：唯一知道所有具體型別的地方。
 * 由外而內組裝 proxy → domain service → application，並把 application provide 給元件層：
 * `const { $tradeCardApplication } = useNuxtApp()`。
 */
export default defineNuxtPlugin(() => {
  // 打真的板子還是打開發用的模擬後端，只看這一個設定（NUXT_PUBLIC_API_BASE_URL）。
  const apiBaseUrl = useRuntimeConfig().public.apiBaseUrl

  // 記著這台瀏覽器那一段登入的只有一份：每一發請求都從它取憑證，被登出時也是它被清掉。
  const sessionStorageProxy = new SessionStorageProxy()

  const userSessionApplication = new UserSessionApplication(
    new UserSessionService(new AuthProxy(apiBaseUrl, sessionStorageProxy), sessionStorageProxy),
  )

  // 板子說這一段登入不算數了：清掉它、回登入畫面。發請求的那一層不懂導頁，所以在這裡接線。
  const onSignedOut = (): void => {
    userSessionApplication.signOut()
    void navigateTo(LOGIN_PATH)
  }

  const tradeCardApplication = new TradeCardApplication(
    new TradeCardService(
      new TradeCardProxy(apiBaseUrl, sessionStorageProxy, onSignedOut),
      new OrderProxy(apiBaseUrl, sessionStorageProxy, onSignedOut),
    ),
  )

  const positionApplication = new PositionApplication(
    new PositionService(new PositionProxy(apiBaseUrl, sessionStorageProxy, onSignedOut)),
  )

  const monthSummaryApplication = new MonthSummaryApplication(
    new MonthSummaryService(new MonthSummaryProxy(apiBaseUrl, sessionStorageProxy, onSignedOut)),
  )

  return {
    provide: {
      userSessionApplication,
      tradeCardApplication,
      positionApplication,
      monthSummaryApplication,
    },
  }
})
