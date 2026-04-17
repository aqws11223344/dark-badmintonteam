FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
COPY . .
RUN go mod tidy && CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Taipei
WORKDIR /
COPY --from=builder /server /server
COPY --from=builder /app/web /web
EXPOSE 8080
CMD ["/server"]
