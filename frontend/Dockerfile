FROM node:20.20.2-alpine3.23@sha256:fb4cd12c85ee03686f6af5362a0b0d56d50c58a04632e6c0fb8363f609372293 AS build

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:1.31.0-alpine3.23@sha256:2f07d83bf561b506400dc183b1b2003803e39efbd22451f848adaba14d28c7c7

RUN apk upgrade --no-cache && \
    chown -R nginx:nginx /var/cache/nginx /run /etc/nginx/conf.d /usr/share/nginx/html

COPY --chown=nginx:nginx nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build --chown=nginx:nginx /app/dist /usr/share/nginx/html

USER nginx
EXPOSE 8080
HEALTHCHECK --interval=5s --timeout=3s --retries=10 \
    CMD wget -q --spider http://127.0.0.1:8080/frontend-healthz || exit 1
