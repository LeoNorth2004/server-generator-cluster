FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY .env ./
COPY libs/go-common ./libs/go-common
COPY apps/generator-service ./apps/generator-service

WORKDIR /app/apps/generator-service
RUN ls -la generator/
RUN cp generator/*.go .
RUN ls -la *.go | head -20
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o generator-service .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/apps/generator-service/generator-service .
COPY --from=builder /app/.env ./.env

EXPOSE 8084

CMD ["./generator-service"]
