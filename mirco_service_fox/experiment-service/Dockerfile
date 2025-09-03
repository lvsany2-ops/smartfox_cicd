FROM golang:1.25-alpine AS builder

# 安装依赖
RUN apk add --no-cache git

# 设置工作目录
WORKDIR /app

# 设置镜像源
ENV GOPROXY=https://mirrors.aliyun.com/goproxy,direct

# 复制go.mod和go.sum文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o experiment-service .

# 创建最终镜像
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/experiment-service .

# 暴露端口
EXPOSE 8082

# 运行应用
CMD ["./experiment-service"]