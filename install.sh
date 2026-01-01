#!/bin/bash

# vnstat-http-server 一键安装脚本
# 支持安装、升级、卸载、配置

# 注意：不使用 set -e，因为我们需要在菜单中处理错误并继续
# set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置变量
REPO="kchen0x/vnstat-http-server"
BINARY_NAME="vnstat-http-server"
INSTALL_DIR="/usr/local/bin"
SERVICE_NAME="vnstat-server"
SERVICE_FILE="/etc/systemd/system/${SERVICE_NAME}.service"
CONFIG_FILE="/etc/vnstat-http-server.conf"

# 检测是否为 root 用户，如果是 root 就不使用 sudo
if [ "$(id -u)" -eq 0 ]; then
    SUDO=""
else
    SUDO="sudo"
fi

# 检测系统架构
detect_arch() {
    local arch=$(uname -m)
    case $arch in
        x86_64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        *)
            echo "unsupported"
            ;;
    esac
}

# 获取最新版本号
get_latest_version() {
    curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
}

# 获取下载 URL
get_download_url() {
    local arch=$1
    local version=$2
    echo "https://github.com/${REPO}/releases/download/${version}/vnstat-http-server-linux-${arch}"
}

# 下载二进制文件
download_binary() {
    local arch=$1
    local version=$2
    local url=$(get_download_url $arch $version)
    local temp_file=$(mktemp)
    
    # 将提示信息输出到 stderr，避免被命令替换捕获
    echo -e "${BLUE}正在下载 vnstat-http-server ${version}...${NC}" >&2
    if curl -L -f -s -o "$temp_file" "$url" 2>/dev/null; then
        echo "$temp_file"
    else
        echo -e "${RED}下载失败: $url${NC}" >&2
        rm -f "$temp_file"
        return 1
    fi
}

# 检查 vnstat 是否安装
check_vnstat() {
    if ! command -v vnstat &> /dev/null; then
        echo -e "${YELLOW}警告: 未检测到 vnstat，请先安装 vnstat${NC}"
        echo -e "安装方法: ${BLUE}https://humdi.net/vnstat/${NC}"
        
        # 如果是交互式终端，询问用户
        if is_interactive_terminal; then
            read -p "是否继续安装? (y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                return 1
            fi
        else
            # 非交互式环境，默认继续
            echo -e "${YELLOW}非交互式环境，继续安装...${NC}"
        fi
    fi
    return 0
}

# 安装
install() {
    echo -e "${GREEN}=== 安装 vnstat-http-server ===${NC}"
    
    # 检查是否已安装
    if [ -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        echo -e "${YELLOW}检测到已安装的版本，请使用 'upgrade' 命令升级${NC}"
        return 1
    fi
    
    # 检查 vnstat
    if ! check_vnstat; then
        return 1
    fi
    
    # 检测架构
    local arch=$(detect_arch)
    if [ "$arch" = "unsupported" ]; then
        echo -e "${RED}不支持的系统架构: $(uname -m)${NC}"
        return 1
    fi
    
    echo -e "${BLUE}检测到系统架构: ${arch}${NC}"
    
    # 获取最新版本
    echo -e "${BLUE}正在获取最新版本...${NC}"
    local version=$(get_latest_version)
    if [ -z "$version" ]; then
        echo -e "${RED}无法获取最新版本${NC}"
        return 1
    fi
    
    echo -e "${GREEN}最新版本: ${version}${NC}"
    
    # 下载二进制文件
    local temp_file=$(download_binary $arch $version)
    
    # 安装二进制文件
    echo -e "${BLUE}正在安装到 ${INSTALL_DIR}...${NC}"
    ${SUDO} mv "$temp_file" "${INSTALL_DIR}/${BINARY_NAME}"
    ${SUDO} chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    
    echo -e "${GREEN}二进制文件安装完成${NC}"
    
    # 配置
    configure
    
    # 创建 systemd 服务
    create_service
    
    # 启动服务
    echo -e "${BLUE}正在启动服务...${NC}"
    ${SUDO} systemctl daemon-reload
    ${SUDO} systemctl enable ${SERVICE_NAME}
    ${SUDO} systemctl start ${SERVICE_NAME}
    
    echo -e "${GREEN}安装完成！${NC}"
    echo -e "${BLUE}服务状态:${NC}"
    ${SUDO} systemctl status ${SERVICE_NAME} --no-pager -l || true
}

# 升级
upgrade() {
    echo -e "${GREEN}=== 升级 vnstat-http-server ===${NC}"
    
    # 检查是否已安装
    if [ ! -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        echo -e "${YELLOW}未检测到已安装的版本，请使用 'install' 命令安装${NC}"
        return 1
    fi
    
    # 检测架构
    local arch=$(detect_arch)
    if [ "$arch" = "unsupported" ]; then
        echo -e "${RED}不支持的系统架构: $(uname -m)${NC}"
        exit 1
    fi
    
    # 获取最新版本
    echo -e "${BLUE}正在获取最新版本...${NC}"
    local version=$(get_latest_version)
    if [ -z "$version" ]; then
        echo -e "${RED}无法获取最新版本${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}最新版本: ${version}${NC}"
    
    # 下载二进制文件
    local temp_file=$(download_binary $arch $version)
    
    # 停止服务
    if systemctl is-active --quiet ${SERVICE_NAME}; then
        echo -e "${BLUE}正在停止服务...${NC}"
        ${SUDO} systemctl stop ${SERVICE_NAME}
    fi
    
    # 备份旧版本
    if [ -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        echo -e "${BLUE}备份旧版本...${NC}"
        ${SUDO} cp "${INSTALL_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/${BINARY_NAME}.bak"
    fi
    
    # 安装新版本
    echo -e "${BLUE}正在安装新版本...${NC}"
    ${SUDO} mv "$temp_file" "${INSTALL_DIR}/${BINARY_NAME}"
    ${SUDO} chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
    
    # 启动服务
    echo -e "${BLUE}正在启动服务...${NC}"
    ${SUDO} systemctl daemon-reload
    ${SUDO} systemctl start ${SERVICE_NAME}
    
    # 删除备份
    if [ -f "${INSTALL_DIR}/${BINARY_NAME}.bak" ]; then
        echo -e "${BLUE}删除备份文件...${NC}"
        ${SUDO} rm -f "${INSTALL_DIR}/${BINARY_NAME}.bak"
    fi
    
    echo -e "${GREEN}升级完成！${NC}"
    echo -e "${BLUE}服务状态:${NC}"
    ${SUDO} systemctl status ${SERVICE_NAME} --no-pager -l || true
}

# 卸载
uninstall() {
    echo -e "${YELLOW}=== 卸载 vnstat-http-server ===${NC}"
    
    # 检查是否已安装
    if [ ! -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        echo -e "${YELLOW}未检测到已安装的版本${NC}"
        exit 0
    fi
    
    # 如果是交互式终端，询问用户
    if is_interactive_terminal; then
        read -p "确定要卸载 vnstat-http-server? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            echo -e "${BLUE}已取消${NC}"
            return 0
        fi
    else
        # 非交互式环境，直接卸载
        echo -e "${YELLOW}非交互式环境，确认卸载...${NC}"
    fi
    
    # 停止并禁用服务
    if systemctl is-active --quiet ${SERVICE_NAME}; then
        echo -e "${BLUE}正在停止服务...${NC}"
        ${SUDO} systemctl stop ${SERVICE_NAME}
    fi
    
    if systemctl is-enabled --quiet ${SERVICE_NAME} 2>/dev/null; then
        echo -e "${BLUE}正在禁用服务...${NC}"
        ${SUDO} systemctl disable ${SERVICE_NAME}
    fi
    
    # 删除服务文件
    if [ -f "${SERVICE_FILE}" ]; then
        echo -e "${BLUE}正在删除服务文件...${NC}"
        ${SUDO} rm -f "${SERVICE_FILE}"
        ${SUDO} systemctl daemon-reload
    fi
    
    # 删除二进制文件
    echo -e "${BLUE}正在删除二进制文件...${NC}"
    ${SUDO} rm -f "${INSTALL_DIR}/${BINARY_NAME}"
    
    # 删除配置文件（可选）
    if [ -f "${CONFIG_FILE}" ]; then
        if is_interactive_terminal; then
            read -p "是否删除配置文件 ${CONFIG_FILE}? (y/N): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                ${SUDO} rm -f "${CONFIG_FILE}"
            fi
        else
            # 非交互式环境，保留配置文件
            echo -e "${YELLOW}非交互式环境，保留配置文件${NC}"
        fi
    fi
    
    echo -e "${GREEN}卸载完成！${NC}"
}

# 检查是否为交互式终端
is_interactive_terminal() {
    [ -t 0 ] && [ -t 1 ]
}

# 交互式配置
configure() {
    echo -e "${GREEN}=== 配置 vnstat-http-server ===${NC}"
    
    # 读取现有配置（如果存在）
    PORT="8080"
    TOKEN=""
    INTERFACE=""
    GRAFANA_URL=""
    GRAFANA_USER=""
    GRAFANA_TOKEN=""
    GRAFANA_INTERVAL="30s"
    
    if [ -f "${CONFIG_FILE}" ]; then
        source "${CONFIG_FILE}"
        # 确保变量有值（兼容旧配置文件）
        PORT="${PORT:-8080}"
        TOKEN="${TOKEN:-}"
        INTERFACE="${INTERFACE:-}"
        GRAFANA_URL="${GRAFANA_URL:-}"
        GRAFANA_USER="${GRAFANA_USER:-}"
        GRAFANA_TOKEN="${GRAFANA_TOKEN:-}"
        GRAFANA_INTERVAL="${GRAFANA_INTERVAL:-30s}"
    fi
    
    # 检查是否为交互式终端
    if ! is_interactive_terminal; then
        echo -e "${YELLOW}非交互式环境，使用默认配置${NC}"
    else
        # 配置端口
        echo -e "${BLUE}配置 HTTP 端口 (默认: ${PORT}):${NC}"
        read -p "端口: " input_port
        if [ -n "$input_port" ]; then
            PORT="$input_port"
        fi
        
        # 配置 Token
        echo -e "${BLUE}配置认证 Token (留空禁用认证):${NC}"
        read -p "Token: " input_token
        if [ -n "$input_token" ]; then
            TOKEN="$input_token"
        fi
        
        # 配置网络接口
        echo -e "${BLUE}配置网络接口 (留空监控所有接口):${NC}"
        read -p "接口名称 (如 eth0): " input_interface
        if [ -n "$input_interface" ]; then
            INTERFACE="$input_interface"
        fi
        
        # 配置 Grafana Cloud
        echo -e "${BLUE}是否启用 Grafana Cloud 推送? (y/N):${NC}"
        read -p "启用: " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            echo -e "${BLUE}Grafana Cloud Remote Write URL:${NC}"
            echo -e "${YELLOW}示例: https://YOUR_PROMETHEUS_INSTANCE.grafana.net/api/prom/push${NC}"
            read -p "URL: " input_grafana_url
            if [ -n "$input_grafana_url" ]; then
                GRAFANA_URL="$input_grafana_url"
            fi
            
            echo -e "${BLUE}Grafana Cloud Instance ID:${NC}"
            read -p "Instance ID: " input_grafana_user
            if [ -n "$input_grafana_user" ]; then
                GRAFANA_USER="$input_grafana_user"
            fi
            
            echo -e "${BLUE}Grafana Cloud API Token:${NC}"
            read -p "API Token: " input_grafana_token
            if [ -n "$input_grafana_token" ]; then
                GRAFANA_TOKEN="$input_grafana_token"
            fi
            
            echo -e "${BLUE}推送间隔 (默认: ${GRAFANA_INTERVAL}):${NC}"
            read -p "间隔: " input_grafana_interval
            if [ -n "$input_grafana_interval" ]; then
                GRAFANA_INTERVAL="$input_grafana_interval"
            fi
        fi
    fi
    
    # 保存配置
    echo -e "${BLUE}正在保存配置...${NC}"
    ${SUDO} tee "${CONFIG_FILE}" > /dev/null <<EOF
# vnstat-http-server 配置文件
# 生成时间: $(date)

PORT="${PORT}"
TOKEN="${TOKEN}"
INTERFACE="${INTERFACE}"
GRAFANA_URL="${GRAFANA_URL}"
GRAFANA_USER="${GRAFANA_USER}"
GRAFANA_TOKEN="${GRAFANA_TOKEN}"
GRAFANA_INTERVAL="${GRAFANA_INTERVAL}"
EOF
    
    ${SUDO} chmod 600 "${CONFIG_FILE}"
    echo -e "${GREEN}配置已保存到 ${CONFIG_FILE}${NC}"
}

# 创建 systemd 服务
create_service() {
    echo -e "${BLUE}正在创建 systemd 服务...${NC}"
    
    # 读取配置
    if [ ! -f "${CONFIG_FILE}" ]; then
        echo -e "${RED}配置文件不存在，请先运行配置${NC}"
        exit 1
    fi
    
    source "${CONFIG_FILE}"
    
    # 设置默认值（如果配置文件中没有设置或为空）
    PORT="${PORT:-8080}"
    TOKEN="${TOKEN:-}"
    INTERFACE="${INTERFACE:-}"
    GRAFANA_URL="${GRAFANA_URL:-}"
    GRAFANA_USER="${GRAFANA_USER:-}"
    GRAFANA_TOKEN="${GRAFANA_TOKEN:-}"
    GRAFANA_INTERVAL="${GRAFANA_INTERVAL:-30s}"
    
    # 构建 ExecStart 命令
    local exec_start="${INSTALL_DIR}/${BINARY_NAME} -port ${PORT}"
    
    if [ -n "$TOKEN" ]; then
        exec_start="${exec_start} -token ${TOKEN}"
    fi
    
    if [ -n "$INTERFACE" ]; then
        exec_start="${exec_start} -interface ${INTERFACE}"
    fi
    
    if [ -n "$GRAFANA_URL" ] && [ -n "$GRAFANA_USER" ] && [ -n "$GRAFANA_TOKEN" ]; then
        exec_start="${exec_start} -grafana-url \"${GRAFANA_URL}\" -grafana-user \"${GRAFANA_USER}\" -grafana-token \"${GRAFANA_TOKEN}\" -grafana-interval ${GRAFANA_INTERVAL}"
    fi
    
    # 创建服务文件
    ${SUDO} tee "${SERVICE_FILE}" > /dev/null <<EOF
[Unit]
Description=vnstat HTTP Server
After=network.target

[Service]
Type=simple
User=root
ExecStart=${exec_start}
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF
    
    echo -e "${GREEN}服务文件已创建: ${SERVICE_FILE}${NC}"
}

# 显示交互式菜单
show_menu() {
    # 清屏（如果支持）
    if command -v clear &> /dev/null; then
        clear 2>/dev/null || true
    fi
    
    echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║  vnstat-http-server 管理脚本            ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
    echo ""
    
    # 检查安装状态
    local is_installed=false
    if [ -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        is_installed=true
        echo -e "${GREEN}✓ 已安装${NC}"
    else
        echo -e "${YELLOW}○ 未安装${NC}"
    fi
    
    echo ""
    echo -e "${BLUE}请选择操作:${NC}"
    echo ""
    
    if [ "$is_installed" = false ]; then
        echo -e "  ${GREEN}1)${NC} 安装 vnstat-http-server"
        echo -e "  ${YELLOW}2)${NC} 查看帮助"
        echo -e "  ${YELLOW}0)${NC} 退出"
    else
        echo -e "  ${GREEN}1)${NC} 查看服务状态"
        echo -e "  ${GREEN}2)${NC} 升级到最新版本"
        echo -e "  ${GREEN}3)${NC} 配置服务参数"
        echo -e "  ${YELLOW}4)${NC} 卸载 vnstat-http-server"
        echo -e "  ${YELLOW}5)${NC} 查看帮助"
        echo -e "  ${YELLOW}0)${NC} 退出"
    fi
    
    echo ""
    # 直接读取输入，不使用循环
    # 如果 stdin 不是终端，read 会失败或返回空，但不会无限循环
    read -p "请输入选项 [0-5]: " choice || choice=""
    
    # 处理空输入或无效输入
    if [ -z "$choice" ]; then
        echo -e "${RED}无效选项，请重新选择${NC}"
        sleep 1
        show_menu
        return
    fi
    
    case $choice in
        1)
            if [ "$is_installed" = false ]; then
                if install; then
                    echo ""
                    read -p "按 Enter 键返回菜单..." dummy
                else
                    echo ""
                    echo -e "${RED}安装失败，按 Enter 键返回菜单...${NC}"
                    read dummy
                fi
                show_menu
            else
                show_status
                echo ""
                read -p "按 Enter 键返回菜单..." dummy
                show_menu
            fi
            ;;
        2)
            if [ "$is_installed" = false ]; then
                usage
                echo ""
                read -p "按 Enter 键返回菜单..." dummy
                show_menu
            else
                if upgrade; then
                    echo ""
                    read -p "按 Enter 键返回菜单..." dummy
                else
                    echo ""
                    echo -e "${RED}升级失败，按 Enter 键返回菜单...${NC}"
                    read dummy
                fi
                show_menu
            fi
            ;;
        3)
            if [ "$is_installed" = true ]; then
                configure
                if [ -f "${SERVICE_FILE}" ]; then
                    echo -e "${BLUE}正在重新加载服务配置...${NC}"
                    create_service
                    ${SUDO} systemctl daemon-reload
                    if systemctl is-active --quiet ${SERVICE_NAME}; then
                        ${SUDO} systemctl restart ${SERVICE_NAME}
                        echo -e "${GREEN}服务已重启${NC}"
                    fi
                fi
                echo ""
                read -p "按 Enter 键返回菜单..." dummy
                show_menu
            else
                echo -e "${RED}无效选项${NC}"
                sleep 1
                show_menu
            fi
            ;;
        4)
            if [ "$is_installed" = true ]; then
                uninstall
                echo ""
                read -p "按 Enter 键返回菜单..." dummy
                show_menu
            else
                echo -e "${RED}无效选项${NC}"
                sleep 1
                show_menu
            fi
            ;;
        5)
            if [ "$is_installed" = true ]; then
                usage
                echo ""
                read -p "按 Enter 键返回菜单..." dummy
                show_menu
            else
                echo -e "${RED}无效选项${NC}"
                sleep 1
                show_menu
            fi
            ;;
        0)
            echo -e "${BLUE}再见！${NC}"
            exit 0
            ;;
        *)
            echo -e "${RED}无效选项，请重新选择${NC}"
            sleep 1
            show_menu
            return
            ;;
    esac
}

# 显示使用帮助
usage() {
    cat <<EOF
${GREEN}vnstat-http-server 一键安装脚本${NC}

用法: $0 [命令]

命令:
  install     安装 vnstat-http-server
  upgrade     升级到最新版本
  uninstall   卸载 vnstat-http-server
  configure   配置服务参数
  status      查看服务状态
  help        显示此帮助信息

示例:
  $0 install      # 安装并配置
  $0 upgrade      # 升级到最新版本
  $0 configure    # 重新配置
  $0 status       # 查看服务状态
  $0 uninstall    # 卸载

快速安装:
  curl -fsSL https://raw.githubusercontent.com/${REPO}/main/install.sh | bash

或者直接运行脚本进入交互式菜单:
  curl -fsSL https://raw.githubusercontent.com/${REPO}/main/install.sh | bash

EOF
}

# 显示服务状态
show_status() {
    if [ ! -f "${INSTALL_DIR}/${BINARY_NAME}" ]; then
        echo -e "${YELLOW}vnstat-http-server 未安装${NC}"
        return
    fi
    
    echo -e "${GREEN}=== vnstat-http-server 状态 ===${NC}"
    echo -e "${BLUE}二进制文件:${NC} ${INSTALL_DIR}/${BINARY_NAME}"
    
    if [ -f "${CONFIG_FILE}" ]; then
        echo -e "${BLUE}配置文件:${NC} ${CONFIG_FILE}"
    fi
    
    if systemctl list-unit-files | grep -q "${SERVICE_NAME}.service"; then
        echo -e "${BLUE}服务状态:${NC}"
        ${SUDO} systemctl status ${SERVICE_NAME} --no-pager -l
    else
        echo -e "${YELLOW}服务未配置${NC}"
    fi
}

# 主函数
main() {
    # 如果没有参数，显示交互式菜单
    if [ -z "${1:-}" ]; then
        show_menu
        return
    fi
    
    # 如果有参数，执行对应命令
    case "${1}" in
        install)
            install
            ;;
        upgrade)
            upgrade
            ;;
        uninstall)
            uninstall
            ;;
        configure)
            configure
            if [ -f "${SERVICE_FILE}" ]; then
                echo -e "${BLUE}正在重新加载服务配置...${NC}"
                create_service
                ${SUDO} systemctl daemon-reload
                if systemctl is-active --quiet ${SERVICE_NAME}; then
                    ${SUDO} systemctl restart ${SERVICE_NAME}
                    echo -e "${GREEN}服务已重启${NC}"
                fi
            fi
            ;;
        status)
            show_status
            ;;
        menu)
            show_menu
            ;;
        help|--help|-h)
            usage
            ;;
        *)
            echo -e "${RED}未知命令: $1${NC}"
            echo ""
            usage
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"

