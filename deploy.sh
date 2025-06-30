#!/bin/bash

# Util Server 部署脚本
# 用途：自动化部署和管理 util-server 服务

set -e

# 配置变量
SERVICE_NAME="util-server"
SERVICE_USER="www-data"
SERVICE_GROUP="www-data"
INSTALL_DIR="/home/tiger/util-server"
SYSTEMD_DIR="/etc/systemd/system"
BINARY_NAME="server"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# 检查是否为root用户
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "此脚本需要root权限运行"
        exit 1
    fi
}

# 检查系统是否支持systemd
check_systemd() {
    if ! command -v systemctl &> /dev/null; then
        log_error "系统不支持systemd"
        exit 1
    fi
}

# 创建服务用户
create_service_user() {
    if ! id "$SERVICE_USER" &>/dev/null; then
        log_step "创建服务用户: $SERVICE_USER"
        useradd --system --no-create-home --shell /bin/false $SERVICE_USER
    else
        log_info "服务用户 $SERVICE_USER 已存在"
    fi
}

# 编译程序
build_binary() {
    log_step "编译程序..."
    go build -o $BINARY_NAME main.go
    if [[ ! -f $BINARY_NAME ]]; then
        log_error "编译失败"
        exit 1
    fi
    log_info "编译完成"
}

# 安装程序
install_binary() {
    log_step "安装程序到 $INSTALL_DIR"
    
    # 创建安装目录
    mkdir -p $INSTALL_DIR
    
    # 复制二进制文件
    cp $BINARY_NAME $INSTALL_DIR/
    chmod +x $INSTALL_DIR/$BINARY_NAME
    
    # 复制模板文件
    cp -r templates $INSTALL_DIR/
    
    # 复制示例CSV文件
    cp test_data.csv $INSTALL_DIR/
    
    # 创建日志目录
    mkdir -p $INSTALL_DIR/logs
    
    # 设置文件权限
    chown -R $SERVICE_USER:$SERVICE_GROUP $INSTALL_DIR
    chmod -R 755 $INSTALL_DIR
    
    log_info "程序安装完成"
}

# 安装systemd服务
install_service() {
    log_step "安装systemd服务"
    
    # 复制服务文件
    cp $SERVICE_NAME.service $SYSTEMD_DIR/
    
    # 重新加载systemd配置
    systemctl daemon-reload
    
    # 启用服务
    systemctl enable $SERVICE_NAME
    
    log_info "systemd服务安装完成"
}

# 启动服务
start_service() {
    log_step "启动服务"
    systemctl start $SERVICE_NAME
    
    # 检查服务状态
    if systemctl is-active --quiet $SERVICE_NAME; then
        log_info "服务启动成功"
        systemctl status $SERVICE_NAME --no-pager
    else
        log_error "服务启动失败"
        systemctl status $SERVICE_NAME --no-pager
        exit 1
    fi
}

# 停止服务
stop_service() {
    log_step "停止服务"
    if systemctl is-active --quiet $SERVICE_NAME; then
        systemctl stop $SERVICE_NAME
        log_info "服务已停止"
    else
        log_warn "服务未运行"
    fi
}

# 重启服务
restart_service() {
    log_step "重启服务"
    systemctl restart $SERVICE_NAME
    
    if systemctl is-active --quiet $SERVICE_NAME; then
        log_info "服务重启成功"
    else
        log_error "服务重启失败"
        exit 1
    fi
}

# 卸载服务
uninstall_service() {
    log_step "卸载服务"
    
    # 停止服务
    if systemctl is-active --quiet $SERVICE_NAME; then
        systemctl stop $SERVICE_NAME
    fi
    
    # 禁用服务
    if systemctl is-enabled --quiet $SERVICE_NAME; then
        systemctl disable $SERVICE_NAME
    fi
    
    # 删除服务文件
    rm -f $SYSTEMD_DIR/$SERVICE_NAME.service
    
    # 重新加载systemd配置
    systemctl daemon-reload
    
    # 删除安装目录（可选）
    read -p "是否删除安装目录 $INSTALL_DIR? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf $INSTALL_DIR
        log_info "安装目录已删除"
    fi
    
    log_info "服务卸载完成"
}

# 查看服务状态
show_status() {
    echo "=== 服务状态 ==="
    systemctl status $SERVICE_NAME --no-pager
    echo
    echo "=== 服务日志 (最近20行) ==="
    journalctl -u $SERVICE_NAME -n 20 --no-pager
}

# 查看日志
show_logs() {
    echo "=== 实时日志 ==="
    journalctl -u $SERVICE_NAME -f
}

# 显示帮助信息
show_help() {
    echo "用法: $0 {install|start|stop|restart|status|logs|uninstall|help}"
    echo
    echo "命令说明:"
    echo "  install   - 编译、安装并启动服务"
    echo "  start     - 启动服务"
    echo "  stop      - 停止服务"
    echo "  restart   - 重启服务"
    echo "  status    - 查看服务状态和日志"
    echo "  logs      - 实时查看日志"
    echo "  uninstall - 卸载服务"
    echo "  help      - 显示此帮助信息"
    echo
    echo "示例:"
    echo "  sudo $0 install    # 完整安装服务"
    echo "  sudo $0 restart    # 重启服务"
    echo "  sudo $0 status     # 查看状态"
    echo "  sudo $0 logs       # 查看实时日志"
}

# 完整安装
full_install() {
    check_root
    check_systemd
    create_service_user
    build_binary
    install_binary
    install_service
    start_service
    
    echo
    log_info "=== 安装完成 ==="
    log_info "服务名称: $SERVICE_NAME"
    log_info "安装目录: $INSTALL_DIR"
    log_info "Web界面: http://localhost:9983"
    log_info "运行测试: cd $INSTALL_DIR && ./server test test_data.csv"
    echo
    log_info "常用命令:"
    log_info "  sudo systemctl status $SERVICE_NAME    # 查看状态"
    log_info "  sudo systemctl restart $SERVICE_NAME   # 重启服务"
    log_info "  sudo journalctl -u $SERVICE_NAME -f    # 查看日志"
}

# 主函数
main() {
    case "${1:-}" in
        install)
            full_install
            ;;
        start)
            check_root
            start_service
            ;;
        stop)
            check_root
            stop_service
            ;;
        restart)
            check_root
            restart_service
            ;;
        status)
            show_status
            ;;
        logs)
            show_logs
            ;;
        uninstall)
            check_root
            uninstall_service
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

# 执行主函数
main "$@"
