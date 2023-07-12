# 定义变量
APP_NAME=pon-stick-exporter
PORT=9001
IMAGE_NAME=hellowoodes/pon-stick-exporter

# 构建应用程序
build:
	go build -o $(APP_NAME)

# 构建 Docker 镜像
docker-build:
	docker build -t $(IMAGE_NAME) .

# 推送 Docker 镜像
docker-push:
	docker push $(IMAGE_NAME)

# 运行 Docker 容器
docker-run:
	docker run -p $(PORT):$(PORT) --name $(APP_NAME) $(IMAGE_NAME)

# 停止并删除 Docker 容器
docker-stop:
	docker stop $(APP_NAME) && docker rm $(APP_NAME)

# 清除二进制文件和 Docker 镜像
clean:
	rm -f $(APP_NAME)
	docker rmi $(IMAGE_NAME)