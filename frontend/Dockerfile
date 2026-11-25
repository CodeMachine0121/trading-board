# 前端是一包靜態檔，所以最後跑的東西只是一台 nginx。沒有 Node、沒有 SSR——
# server/ 底下的模擬板子只在 `nuxt dev` 時存在，不會進到這個映像檔裡。

FROM oven/bun:1-alpine AS builder

WORKDIR /src

# 先只複製依賴清單，改前端程式碼不會讓安裝那一層失效。
COPY package.json bun.lock ./
# --frozen-lockfile：lockfile 與 package.json 對不起來就失敗。
# --ignore-scripts：postinstall 的 `nuxt prepare` 這時候還沒有原始碼可以看，下面複製完再補跑。
RUN bun install --frozen-lockfile --ignore-scripts

COPY . .
RUN bunx nuxt prepare

# 板子的 API 網址在這裡**烤進 bundle**，不是執行時讀的（Nuxt public runtime config，
# 靜態輸出時直接寫進 HTML）。所以這個映像檔綁定環境：換網址要重新 build，改 ConfigMap 沒有用。
ARG NUXT_PUBLIC_API_BASE_URL=https://trading-board-api.coding-afternoon.com/api/v1
ENV NUXT_PUBLIC_API_BASE_URL=${NUXT_PUBLIC_API_BASE_URL}

RUN bun run generate

# nginx-unprivileged：本來就是非 root，聽 8080，k8s 的非 root 設定跑得動。
FROM nginxinc/nginx-unprivileged:1.27-alpine

COPY nginx.conf /etc/nginx/nginx.conf
COPY --from=builder /src/.output/public /usr/share/nginx/html

EXPOSE 8080

# 給 `docker run` 用的。Kubernetes 用自己的探針。
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --spider http://127.0.0.1:8080/health || exit 1
