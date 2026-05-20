FROM golang:1.21-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY .env ./
COPY libs/go-common ./libs/go-common
COPY apps/api-gateway ./apps/api-gateway

WORKDIR /app/apps/api-gateway
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-gateway .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/apps/api-gateway/api-gateway .
COPY --from=builder /app/.env ./.env

EXPOSE 8080

CMD ["./api-gateway"]
