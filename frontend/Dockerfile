FROM node:26.5.0-alpine3.23@sha256:0473b6671ff22c8eeb570c0e1e51408595d3171e73f8002c269b763f0a943149 AS build

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:1.29.8-alpine3.23@sha256:5616878291a2eed594aee8db4dade5878cf7edcb475e59193904b198d9b830de

RUN apk upgrade --no-cache && \
    chown -R nginx:nginx /var/cache/nginx /run /etc/nginx/conf.d /usr/share/nginx/html

COPY --chown=nginx:nginx nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build --chown=nginx:nginx /app/dist /usr/share/nginx/html

USER nginx
EXPOSE 8080
HEALTHCHECK --interval=5s --timeout=3s --retries=10 \
    CMD wget -q --spider http://127.0.0.1:8080/frontend-healthz || exit 1
