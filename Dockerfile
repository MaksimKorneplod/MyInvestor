FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /myinvestor ./cmd/api

FROM alpine:3.22
RUN apk --no-cache add ca-certificates tzdata
RUN addgroup -S -g 10001 app && adduser -S -D -H -u 10001 -G app app
WORKDIR /app
COPY --from=builder --chown=app:app /myinvestor ./myinvestor
USER app
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q -O /dev/null http://127.0.0.1:8080/api/v1/health || exit 1
CMD ["/app/myinvestor"]