# syntax=docker/dockerfile:1.6
FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app

# ---- 第 1 層快取：只依賴 go.mod / go.sum，程式碼改動時這層不會失效 ----
COPY go.mod ./
COPY go.sum* ./
RUN go mod download 2>/dev/null || go mod tidy

# ---- 第 2 層：複製原始碼 + build ----
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o /server ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Taipei
WORKDIR /
COPY --from=builder /server /server
COPY --from=builder /app/web /web
EXPOSE 8080
CMD ["/server"]
