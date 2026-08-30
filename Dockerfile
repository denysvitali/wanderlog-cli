FROM golang:1.26.6-alpine AS builder

ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE=unknown

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags "-s -w -X github.com/denysvitali/wanderlog-cli/cmd.Version=${VERSION} -X github.com/denysvitali/wanderlog-cli/cmd.Commit=${COMMIT} -X github.com/denysvitali/wanderlog-cli/cmd.BuildDate=${BUILD_DATE}" \
    -o wanderlog .

FROM alpine:3.23
ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE=unknown

LABEL org.opencontainers.image.title="wanderlog-cli" \
      org.opencontainers.image.description="Command-line client for Wanderlog" \
      org.opencontainers.image.source="https://github.com/denysvitali/wanderlog-cli" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}" \
      org.opencontainers.image.created="${BUILD_DATE}"

RUN apk add --no-cache ca-certificates \
    && addgroup -S wanderlog \
    && adduser -S -G wanderlog -h /home/wanderlog wanderlog
COPY --from=builder /build/wanderlog /usr/local/bin/wanderlog
USER wanderlog:wanderlog
WORKDIR /home/wanderlog
ENTRYPOINT ["wanderlog"]
