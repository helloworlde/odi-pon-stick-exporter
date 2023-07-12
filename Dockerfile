# 第一阶段：构建二进制文件
FROM golang:1.20-alpine AS build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pon-stick-exporter .

# 第二阶段：构建最终镜像
FROM alpine:3.14
# 暴露端口
EXPOSE 9001
WORKDIR /root/
COPY --from=build /app/pon-stick-exporter .
CMD ["./pon-stick-exporter"]