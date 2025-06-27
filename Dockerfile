# 构建阶段
FROM golang:1.23-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o html2image .

# 运行阶段
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache \
    chromium \
    chromium-chromedriver \
    harfbuzz \
    freetype \
    ttf-freefont \
    font-noto \
    font-noto-cjk \
    wqy-zenhei \
    ca-certificates \
    && rm -rf /var/cache/apk/*

# 创建非 root 用户
RUN addgroup -g 1001 -S html2image && \
    adduser -S -u 1001 -G html2image html2image

WORKDIR /app

# 复制编译好的二进制文件
COPY --from=builder /app/html2image .
COPY test.html .

# 设置 Chrome 运行所需的环境变量
ENV CHROME_BIN=/usr/bin/chromium-browser \
    CHROME_PATH=/usr/lib/chromium/ \
    CHROMIUM_FLAGS="--disable-software-rasterizer --disable-dev-shm-usage"

# 创建输出目录
RUN mkdir -p /app/output && chown -R html2image:html2image /app

# 切换到非 root 用户
USER html2image

# 暴露 API 端口
EXPOSE 8080

# 运行 HTTP 服务器
CMD ["./html2image"]