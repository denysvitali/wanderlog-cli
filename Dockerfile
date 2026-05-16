FROM golang:1.26-alpine AS builder

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o wanderlog .

FROM alpine:3.23
RUN apk add --no-cache ca-certificates
COPY --from=builder /build/wanderlog /usr/local/bin/wanderlog
ENTRYPOINT ["wanderlog"]
