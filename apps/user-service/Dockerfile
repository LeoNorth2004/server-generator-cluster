FROM golang:1.21-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY .env ./
COPY libs/go-common ./libs/go-common
COPY apps/user-service ./apps/user-service

WORKDIR /app/apps/user-service
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o user-service .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/apps/user-service/user-service .
COPY --from=builder /app/.env ./.env

EXPOSE 8081

CMD ["./user-service"]
