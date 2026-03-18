FROM node:20-alpine AS web-builder
WORKDIR /web
COPY web/package*.json ./
RUN npm ci
COPY web/ ./
RUN npm run build

FROM golang:1.23-alpine AS go-builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web-builder /web/dist ./web/dist
ARG VERSION=dev
RUN CGO_ENABLED=1 go build \
    -ldflags="-s -w -X main.version=${VERSION}" \
    -o /openfiltr ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -S openfiltr \
    && adduser -S -G openfiltr openfiltr
WORKDIR /app
COPY --from=go-builder /openfiltr ./openfiltr
COPY --from=web-builder /web/dist ./web/dist
COPY openapi/ ./openapi/
RUN mkdir -p /data /etc/openfiltr \
    && chown -R openfiltr:openfiltr /data /etc/openfiltr /app
USER openfiltr
EXPOSE 3000 5353/udp
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:3000/api/v1/system/health || exit 1
ENTRYPOINT ["/app/openfiltr", "--config", "/etc/openfiltr/app.yaml"]
