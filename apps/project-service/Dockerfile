FROM golang:1.21-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY .env ./
COPY libs/go-common ./libs/go-common
COPY apps/project-service ./apps/project-service

WORKDIR /app/apps/project-service
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o project-service .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/apps/project-service/project-service .
COPY --from=builder /app/.env ./.env

EXPOSE 8083

CMD ["./project-service"]
