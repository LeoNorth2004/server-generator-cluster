
FROM node:20-alpine AS builder

WORKDIR /app

# 只复制web-admin目录的文件
COPY apps/web-admin/package*.json ./
RUN npm install

# 复制 .env 文件（确保构建时使用正确的 API URL）
COPY apps/web-admin/.env ./
COPY apps/web-admin/ ./
RUN npm run build

FROM nginx:alpine

ARG DEPLOY_MODE=k8s

# 根据 DEPLOY_MODE 选择 Nginx 配置文件
COPY --from=builder /app/dist /usr/share/nginx/html

# 默认使用 K8s 配置，如果 DEPLOY_MODE=docker 则使用 Docker Compose 配置
COPY apps/web-admin/nginx.k8s.conf /etc/nginx/conf.d/default.conf.k8s
COPY apps/web-admin/nginx.docker.conf /etc/nginx/conf.d/default.conf.docker

# 使用脚本在运行时选择正确的配置
COPY apps/web-admin/docker-entrypoint.sh /docker-entrypoint.sh
RUN chmod +x /docker-entrypoint.sh

EXPOSE 3000

CMD ["/bin/sh", "/docker-entrypoint.sh"]
