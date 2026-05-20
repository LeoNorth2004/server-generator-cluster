FROM golang:1.21-alpine AS builder

ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /app

COPY .env ./
COPY libs/go-common ./libs/go-common
COPY apps/cluster-service ./apps/cluster-service

WORKDIR /app/apps/cluster-service
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cluster-service .

FROM alpine:latest

RUN apk --no-cache add ca-certificates curl sed

RUN curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | TAG=v5.8.3 sh

WORKDIR /root/

COPY --from=builder /app/apps/cluster-service/cluster-service /usr/local/bin/cluster-service
COPY --from=builder /app/.env ./.env
COPY apps/cluster-service/docker-entrypoint.sh ./docker-entrypoint.sh
RUN chmod +x ./docker-entrypoint.sh

EXPOSE 8086

CMD ["./docker-entrypoint.sh"]
