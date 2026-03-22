# 构建阶段
FROM golang:1.26-alpine AS builder

# 设置环境变量，设置go国内代理
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=https://goproxy.cn

WORKDIR /app

COPY . .

RUN go mod download

RUN go build -o webapi main.go

# 运行阶段
FROM alpine:latest

# ca-certificates: 用于 HTTPS 请求
# tzdata: 用于设置时区
RUN apk --no-cache add ca-certificates tzdata

ENV TZ=Asia/Shanghai

WORKDIR /app

# 从构建阶段复制编译好的二进制文件
COPY --from=builder /app/webapi .

# 复制配置文件和资源文件
COPY conf ./conf
RUN mkdir -p core/p5cc/assets
COPY core/p5cc/assets ./core/p5cc/assets

EXPOSE 8080

# 启动命令
CMD ["./webapi"]
