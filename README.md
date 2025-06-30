# 测试服务器

这是一个用于发送AI图像生成请求并处理回调的测试服务器。

## 功能特性

1. 从CSV文件读取测试参数并发送请求
2. 接收API回调并下载生成的图片
3. 提供Web界面查看测试结果
4. 自动组织文件夹结构
5. 支持systemd服务管理
6. 提供完整的部署和运维脚本

## 快速开始

### 开发环境运行

```bash
# 编译并运行
make run

# 或直接使用go
go run main.go
```

### 生产环境部署

```bash
# 一键部署为系统服务
sudo make install

# 或使用部署脚本
sudo ./deploy.sh install
```

## 使用方法

### 1. 启动Web服务器

```bash
go run main.go
```

服务器将在 http://localhost:9983 启动

### 2. 运行测试用例

```bash
go run main.go test test_data.csv
```

这将：
- 创建一个名为 `test-YYYYMMDD-HHMMSS` 的测试文件夹
- 从CSV文件读取测试参数
- 向API发送请求
- 为每个任务创建对应的文件夹

### 3. CSV文件格式

CSV文件需要包含以下列：
- `app_id`: 应用ID
- `role_desc`: 角色描述
- `scenes`: 场景列表（用逗号分隔）
- `style`: 风格

示例：
```csv
app_id,role_desc,scenes,style
11,女 白色裙子 23岁,friends「天生一对」,real
11,男 黑色西装 30岁,office「商务精英」,real
```

## 文件夹结构

```
test-20240630-143022/
├── scene_text2img___098971d7-4703-11f0-8d7f-08bfb88182a2/
│   ├── desc.txt
│   ├── friends.jpg
│   └── 「天生一对」.jpg
└── scene_text2img___098971d7-4703-11f0-8d7f-08bfb88182a3/
    ├── desc.txt
    └── office.jpg
```

## API配置

默认配置：
- API地址: `http://192.168.11.4:7865/comfyui/scene_text2img`
- 回调地址: `http://localhost:9983/callback`

可以在代码中修改这些配置。

## Web界面

访问 http://localhost:9983 可以：
- 查看所有测试文件夹
- 浏览每个测试的任务列表
- 查看生成的图片和描述信息

## 服务管理

### 使用Makefile（推荐）

```bash
# 编译程序
make build

# 本地运行
make run

# 运行测试
make test

# 安装为系统服务
sudo make install

# 服务管理
sudo make start      # 启动服务
sudo make stop       # 停止服务
sudo make restart    # 重启服务
sudo make status     # 查看状态
sudo make logs       # 查看日志

# 卸载服务
sudo make uninstall
```

### 使用部署脚本

```bash
# 查看帮助
./deploy.sh help

# 完整安装
sudo ./deploy.sh install

# 服务管理
sudo ./deploy.sh start|stop|restart|status|logs

# 卸载服务
sudo ./deploy.sh uninstall
```

### 使用systemctl（服务安装后）

```bash
# 服务状态管理
sudo systemctl start util-server
sudo systemctl stop util-server
sudo systemctl restart util-server
sudo systemctl status util-server

# 开机自启
sudo systemctl enable util-server
sudo systemctl disable util-server

# 查看日志
sudo journalctl -u util-server -f
```

## 配置文件

- `config.env` - 基本配置参数
- `util-server.service` - systemd服务配置
- `test_data.csv` - 示例测试数据

## 部署结构

生产环境文件结构：
```
/opt/util-server/
├── server                    # 主程序
├── templates/               # Web模板
├── test_data.csv           # 示例数据
└── test-YYYYMMDD-HHMMSS/   # 测试结果
```

## 详细文档

- [DEPLOYMENT.md](DEPLOYMENT.md) - 详细部署指南
- [config.env](config.env) - 配置文件说明

## 注意事项

1. 生产环境建议使用systemd服务管理
2. 服务默认运行在`www-data`用户下，确保安全性
3. 定期清理测试结果文件夹，避免磁盘空间不足
4. 修改配置后需要重启服务生效
