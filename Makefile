.PHONY: build run test clean deps

# 构建目标
BINARY_NAME=transcribeserver
BUILD_DIR=build

# 默认目标
all: deps build

# 安装依赖
deps:
	go mod tidy
	go mod download

# 构建
build:
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) main.go

# 运行
run:
	go run main.go

# 运行（指定配置文件）
run-config:
	go run main.go -config=config.yaml

# 测试
test:
	go test ./...

# 清理
clean:
	rm -rf $(BUILD_DIR)
	go clean

# 安装
install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

# 开发模式运行
dev:
	go run main.go -config=config.yaml

# 构建 Docker 镜像
docker-build:
	docker build -t transcribeserver .

# 运行 Docker 容器
docker-run:
	docker run -p 8080:8080 -v $(PWD)/models:/app/models transcribeserver

# 帮助
help:
	@echo "可用的目标:"
	@echo "  build      - 构建可执行文件"
	@echo "  run        - 运行服务器"
	@echo "  test       - 运行测试"
	@echo "  clean      - 清理构建文件"
	@echo "  install    - 安装到系统"
	@echo "  deps       - 安装依赖"
	@echo "  dev        - 开发模式运行"
	@echo "  docker-build - 构建 Docker 镜像"
	@echo "  docker-run   - 运行 Docker 容器" 