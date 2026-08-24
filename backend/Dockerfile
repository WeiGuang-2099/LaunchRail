FROM golang:1.27.0-alpine3.23@sha256:3747dcba41c8b0db3211fda4db61638b980e17ac5bb3c94460a975a9cfe19395 AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .

ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api && \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM alpine:3.23@sha256:fd791d74b68913cbb027c6546007b3f0d3bc45125f797758156952bc2d6daf40

RUN apk add --no-cache ca-certificates && \
    addgroup -S launchrail && \
    adduser -S -G launchrail -h /app launchrail

WORKDIR /app
COPY --from=build --chown=launchrail:launchrail /out/api /out/migrate ./

USER launchrail
EXPOSE 8080
HEALTHCHECK --interval=5s --timeout=3s --retries=10 \
    CMD wget -q --spider http://127.0.0.1:8080/readyz || exit 1
CMD ["/app/api"]
