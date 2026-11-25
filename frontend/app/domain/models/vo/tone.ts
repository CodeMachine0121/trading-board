/**
 * 一個狀態或數字在畫面上的語氣。由 Domain Model 決定，元件只把它接到 variant。
 *
 * `profit` / `loss` 依台股慣例分別是紅色與綠色——元件不必知道這件事。
 */
export type Tone = 'primary' | 'info' | 'warning' | 'neutral' | 'profit' | 'loss'
