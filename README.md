# 测试服务器

这是一个用于发送AI图像生成请求并处理回调的测试服务器。

## 功能特性

1. 从CSV文件读取测试参数并发送请求
2. 接收API回调并下载生成的图片
3. 提供Web界面查看测试结果
4. 自动组织文件夹结构

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
