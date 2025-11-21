# 多阶段构建Dockerfile for UniSmith前端项目
# 阶段1: 构建环境
FROM node:20-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制package.json和package-lock.json
COPY frontend/package.json ./
COPY frontend/package-lock.json ./

# 设置网络配置（提高超时时间和重试次数）
RUN npm config set fetch-retry-mintimeout 20000 && \
    npm config set fetch-retry-maxtimeout 120000 && \
    npm config set fetch-retries 5 && \
    npm config set fetch-timeout 600000 && \
    npm cache clean --force

# 安装依赖 (使用 npm ci 可以保证每次安装的依赖版本都是固定的，提升构建的可复现性)
# 如果 npm ci 失败，会自动尝试 npm install
RUN npm ci || npm install --verbose

# 复制源代码和配置文件
COPY frontend/src ./src
COPY frontend/public ./public
COPY frontend/config ./config
COPY frontend/nginx.conf .

# 构建项目
RUN npx max build

# 阶段2: 生产环境
FROM nginx:1.25-alpine AS production

# 安装必要的工具
RUN apk add --no-cache tzdata curl

# 设置时区为中国时区
ENV TZ=Asia/Shanghai

# 创建nginx用户目录
RUN mkdir -p /var/cache/nginx/client_temp && \
    mkdir -p /var/cache/nginx/proxy_temp && \
    mkdir -p /var/cache/nginx/fastcgi_temp && \
    mkdir -p /var/cache/nginx/uwsgi_temp && \
    mkdir -p /var/cache/nginx/scgi_temp

# 复制构建产物到nginx静态文件目录
COPY --from=builder /app/dist /usr/share/nginx/html

# 复制nginx配置文件
COPY frontend/nginx.conf /etc/nginx/nginx.conf

# 暴露80端口
EXPOSE 80

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD curl -f http://localhost/ || exit 1

# 启动nginx
CMD ["nginx", "-g", "daemon off;"] 