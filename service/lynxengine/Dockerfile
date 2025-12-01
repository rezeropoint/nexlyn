# syntax=docker/dockerfile:1.7
FROM golang:1.25.1-alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED=0
ENV GOPROXY=https://goproxy.cn,direct
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

RUN apk update --no-cache && apk add --no-cache tzdata git openssh ca-certificates
RUN go env -w GOPRIVATE=github.com/rezeropoint/* GONOSUMDB=github.com/rezeropoint/*

WORKDIR /build

# 预热 SSH known_hosts，避免首次连接交互
RUN mkdir -p -m 0700 /root/.ssh && ssh-keyscan github.com >> /root/.ssh/known_hosts

# 使用构建参数传递 SSH 密钥
ARG SSH_PRIVATE_KEY
RUN if [ -n "$SSH_PRIVATE_KEY" ]; then \
        echo "$SSH_PRIVATE_KEY" > /root/.ssh/id_ed25519 && \
        chmod 600 /root/.ssh/id_ed25519; \
    fi

# 配置 Git 使用 SSH 而不是 HTTPS 访问私有仓库
RUN git config --global url."git@github.com:".insteadOf "https://github.com/"

# 先复制依赖文件，利用Docker缓存层
ADD go.mod .
ADD go.sum .
# 使用 BuildKit 缓存挂载加速依赖下载
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# 复制 etc 配置目录
COPY ./service/lynxengine/etc /app/etc


# 只复制后端需要的源代码，避免复制前端文件
COPY ./internal ./internal
COPY ./pkg ./pkg
COPY ./restful ./restful
COPY ./service ./service

# 构建可执行文件（使用缓存加速编译）
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -tags postgres -ldflags="-s -w" -o /app/lynxengine ./service/lynxengine

# 清理 SSH 密钥（安全考虑）
RUN rm -f /root/.ssh/id_ed25519

# 使用upx压缩主程序
#RUN chmod a+x ./upx
#RUN ./upx -9 /app/lynxengine

FROM alpine:latest

RUN apk --no-cache add ca-certificates wget

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /usr/share/zoneinfo/Asia/Shanghai /usr/share/zoneinfo/Asia/Shanghai
ENV TZ=Asia/Shanghai

WORKDIR /app
COPY --from=builder /app/lynxengine /app/lynxengine
COPY --from=builder /app/etc /app/etc

CMD ["./lynxengine", "-f", "etc/config.yaml"]
