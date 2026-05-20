FROM golang:1.21-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY .env ./
COPY libs/go-common ./libs/go-common
COPY apps/authentication-service ./apps/authentication-service

WORKDIR /app/apps/authentication-service
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o auth-service .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/apps/authentication-service/auth-service .
COPY --from=builder /app/.env ./.env

EXPOSE 8082

CMD ["./auth-service"]
