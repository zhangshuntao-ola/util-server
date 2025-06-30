# Makefile for util-server

.PHONY: build run test clean install uninstall start stop restart status logs help

# 默认目标
all: build

# 编译程序
build:
	@echo "编译程序..."
	go build -o server main.go
	@echo "编译完成"

# 本地运行
run: build
	@echo "启动服务..."
	./server

# 运行测试
test: build
	@echo "运行测试用例..."
	./server test test_data.csv

# 清理构建文件
clean:
	@echo "清理文件..."
	rm -f server
	rm -rf test-*
	@echo "清理完成"

# 系统服务管理（需要sudo权限）
install:
	@echo "安装系统服务..."
	sudo ./deploy.sh install

uninstall:
	@echo "卸载系统服务..."
	sudo ./deploy.sh uninstall

start:
	@echo "启动服务..."
	sudo ./deploy.sh start

stop:
	@echo "停止服务..."
	sudo ./deploy.sh stop

restart:
	@echo "重启服务..."
	sudo ./deploy.sh restart

status:
	@echo "查看服务状态..."
	sudo ./deploy.sh status

logs:
	@echo "查看服务日志..."
	sudo ./deploy.sh logs

# 开发相关
dev: build
	@echo "开发模式运行..."
	GIN_MODE=debug ./server

# 格式化代码
fmt:
	@echo "格式化代码..."
	go fmt ./...

# 代码检查
vet:
	@echo "代码检查..."
	go vet ./...

# 下载依赖
deps:
	@echo "下载依赖..."
	go mod download
	go mod tidy

# 显示帮助
help:
	@echo "可用命令:"
	@echo "  build     - 编译程序"
	@echo "  run       - 本地运行服务"
	@echo "  test      - 运行测试用例"
	@echo "  dev       - 开发模式运行"
	@echo "  clean     - 清理构建文件"
	@echo ""
	@echo "系统服务管理:"
	@echo "  install   - 安装为系统服务（需要sudo）"
	@echo "  uninstall - 卸载系统服务（需要sudo）"
	@echo "  start     - 启动服务（需要sudo）"
	@echo "  stop      - 停止服务（需要sudo）"
	@echo "  restart   - 重启服务（需要sudo）"
	@echo "  status    - 查看服务状态"
	@echo "  logs      - 查看服务日志"
	@echo ""
	@echo "开发工具:"
	@echo "  fmt       - 格式化代码"
	@echo "  vet       - 代码检查"
	@echo "  deps      - 下载依赖"
	@echo "  help      - 显示此帮助信息"
