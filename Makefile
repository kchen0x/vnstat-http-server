.PHONY: build clean help

# 项目名称
BINARY_NAME=vnstat-http-server
BIN_DIR=bin

help:
	@echo "可用命令:"
	@echo "  make build    - 编译多平台二进制文件"
	@echo "  make clean    - 清理编译产物"
	@echo "  make help     - 显示此帮助信息"

build:
	@echo "开始编译多平台二进制文件..."
	@mkdir -p $(BIN_DIR)
	@echo "编译 Linux amd64..."
	@GOOS=linux GOARCH=amd64 go build -o $(BIN_DIR)/$(BINARY_NAME)-linux-amd64 .
	@echo "编译 Linux arm64..."
	@GOOS=linux GOARCH=arm64 go build -o $(BIN_DIR)/$(BINARY_NAME)-linux-arm64 .
	@echo "编译完成！"
	@echo "输出文件:"
	@ls -lh $(BIN_DIR)/$(BINARY_NAME)-*

clean:
	@echo "清理编译产物..."
	@rm -rf $(BIN_DIR)
	@echo "清理完成！"

