# 构建阶段
FROM golang:1.25-alpine AS builder

# 设置环境变量，设置go国内代理
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.cn

# 设置工作目录
WORKDIR /app

# 复制 go.mod 和 go.sum (如果存在) 并下载依赖
# 使用通配符 go.sum* 避免文件不存在时报错
COPY go.mod go.sum* ./
RUN go mod download

# 复制项目源码
COPY . .

# 编译应用，输出名为 webapi 的二进制文件
RUN go build -o webapi main.go

# 运行阶段
FROM alpine:latest

# 安装必要的系统运行时依赖
# ca-certificates: 用于 HTTPS 请求
# tzdata: 用于设置时区
RUN apk --no-cache add ca-certificates tzdata

# 设置时区为上海 (可选)
ENV TZ=Asia/Shanghai

# 设置工作目录
WORKDIR /app

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/webapi .

# 复制配置文件和资源文件
# 根据 main.go 的逻辑，程序会读取 conf/config.yaml
COPY conf ./conf

# 根据 config.yaml 的配置，需要 core/p5cc/assets 目录下的资源
# 并且要保持相对路径结构 core/p5cc/assets/...
# 我们先创建目录结构，然后复制
RUN mkdir -p core/p5cc/assets
COPY core/p5cc/assets ./core/p5cc/assets

EXPOSE 8080

# 启动命令
CMD ["./webapi"]
