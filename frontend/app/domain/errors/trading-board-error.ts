/**
 * 所有哨兵錯誤的共同祖先。它的 `message` 一律是可以直接給使用者看的繁體中文，
 * 元件據此判斷「這是我們認得的錯誤，可以照實顯示」。
 */
export abstract class TradingBoardError extends Error {}
