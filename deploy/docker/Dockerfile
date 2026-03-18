FROM golang:1.24-alpine AS go-builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown
RUN CGO_ENABLED=1 go build \
    -ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildDate=${BUILD_DATE}" \
    -o /openfiltr ./cmd/server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata wget \
    && addgroup -S openfiltr \
    && adduser -S -G openfiltr openfiltr
WORKDIR /app
COPY --from=go-builder /openfiltr ./openfiltr
COPY openapi/ ./openapi/
RUN mkdir -p /data /etc/openfiltr \
    && chown -R openfiltr:openfiltr /data /etc/openfiltr /app
USER openfiltr
EXPOSE 3000 5353/udp
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:3000/api/v1/system/health || exit 1
ENTRYPOINT ["/app/openfiltr", "--config", "/etc/openfiltr/app.yaml"]
