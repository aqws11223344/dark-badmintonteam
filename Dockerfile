FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app

# ---- 第 1 層：依賴（只在 go.mod 改變時失效）----
COPY go.mod ./
RUN go mod download

# ---- 第 2 層：原始碼 + build ----
COPY . .
RUN go mod tidy && CGO_ENABLED=0 go build -ldflags="-s -w" -o /server ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Taipei
WORKDIR /
COPY --from=builder /server /server
COPY --from=builder /app/web /web
EXPOSE 8080
CMD ["/server"]
