# HTML2Image-Go 部署要求

## 系统依赖

### 1. Chrome/Chromium 浏览器（必需）
这是最关键的依赖，因为 chromedp 需要控制 Chrome 浏览器来渲染 HTML。

**Linux 安装方式：**
```bash
# Ubuntu/Debian
apt-get update && apt-get install -y \
    chromium-browser \
    chromium-chromedriver

# 或者安装 Google Chrome
wget -q -O - https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add -
echo "deb [arch=amd64] http://dl.google.com/linux/chrome/deb/ stable main" >> /etc/apt/sources.list.d/google.list
apt-get update && apt-get install -y google-chrome-stable

# CentOS/RHEL/Amazon Linux
yum install -y chromium chromium-headless

# Alpine Linux
apk add --no-cache chromium chromium-chromedriver
```

### 2. 系统字体（可选但推荐）
为了正确显示中文和其他语言：
```bash
# Ubuntu/Debian
apt-get install -y \
    fonts-liberation \
    fonts-noto-cjk \
    fonts-wqy-zenhei \
    fonts-wqy-microhei

# Alpine Linux
apk add --no-cache \
    font-noto \
    font-noto-cjk \
    wqy-zenhei
```

### 3. 其他系统依赖
```bash
# Ubuntu/Debian
apt-get install -y \
    ca-certificates \
    libnss3 \
    libatk1.0-0 \
    libatk-bridge2.0-0 \
    libcups2 \
    libdrm2 \
    libxkbcommon0 \
    libxcomposite1 \
    libxdamage1 \
    libxrandr2 \
    libgbm1 \
    libgtk-3-0 \
    libasound2
```

## Go 依赖

项目使用的 Go 模块（已在 go.mod 中定义）：
- `github.com/chromedp/chromedp v0.13.7` - Chrome DevTools Protocol 客户端
- 及其相关依赖

## Docker 部署（推荐）

### Dockerfile 示例
```dockerfile
# 构建阶段
FROM golang:1.23-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache git

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o html2image main.go

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

# 设置 Chrome 运行所需的环境变量
ENV CHROME_BIN=/usr/bin/chromium-browser \
    CHROME_PATH=/usr/lib/chromium/ \
    CHROMIUM_FLAGS="--disable-software-rasterizer --disable-dev-shm-usage"

# 切换到非 root 用户
USER html2image

EXPOSE 8080

CMD ["./html2image"]
```

### Docker Compose 示例
```yaml
version: '3.8'

services:
  html2image:
    build: .
    ports:
      - "8080:8080"
    environment:
      - CHROME_BIN=/usr/bin/chromium-browser
      - CHROME_PATH=/usr/lib/chromium/
    volumes:
      - ./output:/app/output
    security_opt:
      - no-new-privileges:true
    cap_drop:
      - ALL
    cap_add:
      - SYS_ADMIN  # Chrome 需要这个权限
```

## 环境变量配置

建议设置的环境变量：
```bash
# Chrome 路径
export CHROME_BIN=/usr/bin/chromium-browser

# 禁用 GPU（在服务器环境推荐）
export CHROMIUM_FLAGS="--disable-gpu --no-sandbox --disable-dev-shm-usage"

# 设置临时目录（如果默认 /tmp 空间不足）
export TMPDIR=/var/tmp
```

## 资源要求

### 最小配置
- CPU: 1 核心
- 内存: 512MB
- 磁盘: 1GB（包括 Chrome 和依赖）

### 推荐配置
- CPU: 2 核心
- 内存: 2GB
- 磁盘: 2GB

### 性能优化建议
1. 使用 SSD 存储提高 I/O 性能
2. 增加 `/dev/shm` 大小（Chrome 使用共享内存）：
   ```bash
   mount -o remount,size=2G /dev/shm
   ```
3. 考虑使用 Chrome 实例池来处理并发请求

## 安全建议

1. **不要以 root 用户运行**
2. **使用沙箱模式**（除非有特殊需求）
3. **限制网络访问**（如果只处理本地 HTML）
4. **设置资源限制**：
   ```bash
   # systemd 服务文件示例
   [Service]
   MemoryLimit=2G
   CPUQuota=200%
   PrivateTmp=true
   NoNewPrivileges=true
   ```

## 监控和日志

建议监控的指标：
- Chrome 进程数量和内存使用
- 转换请求的响应时间
- 失败率和错误类型
- 磁盘空间（临时文件）

## 故障排查

常见问题：
1. **Chrome 启动失败**：检查是否安装了所有系统依赖
2. **字体显示问题**：安装相应的字体包
3. **内存不足**：增加系统内存或调整 Chrome 参数
4. **权限问题**：确保用户有访问 Chrome 和临时目录的权限

## 云平台特定注意事项

### AWS Lambda
- 使用 Lambda Layer 部署 Chrome
- 注意 Lambda 的临时存储限制（512MB）

### Google Cloud Run
- 使用预构建的 Chrome 镜像
- 设置足够的内存限制（至少 1GB）

### Kubernetes
- 考虑使用 InitContainer 安装 Chrome
- 设置适当的资源请求和限制 