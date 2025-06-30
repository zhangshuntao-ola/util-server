# 服务部署指南

本文档详细说明了如何在Linux系统上部署和管理util-server服务。

## 系统要求

- Linux系统（支持systemd）
- Go 1.21 或更高版本
- sudo权限

## 快速部署

### 1. 完整安装服务

```bash
# 克隆或下载项目代码
# cd到项目目录

# 一键安装（包含编译、安装、启动）
sudo ./deploy.sh install
```

### 2. 验证安装

```bash
# 查看服务状态
sudo systemctl status util-server

# 访问Web界面
curl http://localhost:9983

# 或在浏览器中打开
# http://localhost:9983
```

## 详细操作

### 服务管理

```bash
# 启动服务
sudo ./deploy.sh start
sudo systemctl start util-server

# 停止服务
sudo ./deploy.sh stop
sudo systemctl stop util-server

# 重启服务
sudo ./deploy.sh restart
sudo systemctl restart util-server

# 查看状态
sudo ./deploy.sh status
sudo systemctl status util-server

# 查看日志
sudo ./deploy.sh logs
sudo journalctl -u util-server -f

# 卸载服务
sudo ./deploy.sh uninstall
```

### 手动运行测试

```bash
# 进入安装目录
cd /opt/util-server

# 运行测试用例
sudo -u www-data ./server test test_data.csv
```

## 配置说明

### 服务配置文件位置

- **服务文件**: `/etc/systemd/system/util-server.service`
- **程序目录**: `/opt/util-server/`
- **配置文件**: `/opt/util-server/config.env`

### 修改配置

1. 编辑配置文件：
```bash
sudo nano /opt/util-server/config.env
```

2. 修改main.go中的配置变量（如需要）

3. 重新编译和部署：
```bash
# 停止服务
sudo systemctl stop util-server

# 重新编译
go build -o server main.go

# 复制新的二进制文件
sudo cp server /opt/util-server/

# 启动服务
sudo systemctl start util-server
```

### 端口配置

默认Web端口：9983

如需修改端口，编辑`main.go`文件中的`r.Run(":9983")`行。

### API配置

默认API地址：`http://192.168.11.4:7865/comfyui/scene_text2img`

如需修改，编辑`main.go`文件中的`apiURL`变量。

## 文件结构

```
/opt/util-server/
├── server              # 主程序
├── templates/          # Web模板
│   ├── index.html
│   ├── test_detail.html
│   └── task_detail.html
├── test_data.csv       # 示例测试数据
├── test-YYYYMMDD-HHMMSS/  # 测试结果目录
│   └── task_id/
│       ├── desc.txt
│       └── *.jpg
└── logs/              # 日志目录（如果需要）
```

## 安全注意事项

1. **用户权限**: 服务运行在`www-data`用户下，避免使用root权限
2. **文件权限**: 确保程序文件有适当的权限设置
3. **网络安全**: 如果在生产环境，考虑使用防火墙限制端口访问
4. **日志管理**: 定期清理日志文件，避免磁盘空间不足

## 故障排除

### 服务无法启动

```bash
# 查看详细错误信息
sudo journalctl -u util-server -n 50

# 检查配置文件语法
sudo systemctl daemon-reload
```

### 权限问题

```bash
# 重新设置文件权限
sudo chown -R www-data:www-data /opt/util-server
sudo chmod -R 755 /opt/util-server
sudo chmod +x /opt/util-server/server
```

### 端口被占用

```bash
# 检查端口使用情况
sudo netstat -tlnp | grep 9983
sudo lsof -i :9983

# 修改配置使用不同端口
```

### 编译错误

```bash
# 确保Go环境正确
go version

# 检查依赖
go mod tidy

# 重新编译
go build -o server main.go
```

## 监控和维护

### 日志轮转

创建logrotate配置：

```bash
sudo nano /etc/logrotate.d/util-server
```

内容：
```
/opt/util-server/logs/*.log {
    daily
    missingok
    rotate 7
    compress
    delaycompress
    notifempty
    create 644 www-data www-data
    postrotate
        systemctl reload util-server > /dev/null 2>&1 || true
    endscript
}
```

### 自动备份

创建备份脚本：

```bash
#!/bin/bash
# 备份测试结果
BACKUP_DIR="/backup/util-server/$(date +%Y%m%d)"
mkdir -p $BACKUP_DIR
cp -r /opt/util-server/test-* $BACKUP_DIR/
```

### 监控脚本

```bash
#!/bin/bash
# 检查服务状态
if ! systemctl is-active --quiet util-server; then
    echo "util-server服务异常，尝试重启"
    systemctl restart util-server
fi
```

## 更新部署

```bash
# 1. 停止服务
sudo systemctl stop util-server

# 2. 备份当前版本
sudo cp /opt/util-server/server /opt/util-server/server.backup

# 3. 编译新版本
go build -o server main.go

# 4. 部署新版本
sudo cp server /opt/util-server/

# 5. 启动服务
sudo systemctl start util-server

# 6. 验证更新
sudo systemctl status util-server
```
