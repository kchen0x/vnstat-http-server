# vnstat-http-server

一个基于 Go 语言的轻量级单文件工具，用于将 Linux 服务器上的 vnstat 统计数据通过 HTTP 接口暴露出来，方便手机 App、前端网页或脚本进行远程监控。

## 特性

- 🚀 **零依赖**：除 vnstat 本身外，无需安装 Python、PHP、Node 或 Docker
- 📦 **单二进制文件**：编译后只有一个文件，直接运行
- 🔒 **安全**：支持简单的 Token 鉴权
- 🌐 **CORS 支持**：所有接口支持跨域请求
- 📊 **多格式输出**：支持 JSON 和文本两种格式

## 系统要求

- Linux 系统（amd64 / arm64）
- 已安装 `vnstat` 工具
- Go 1.21+ （仅编译时需要）

## 快速开始

### 1. 编译

```bash
make build
```

编译完成后，二进制文件位于 `bin/` 目录：
- `bin/vnstat-http-server-linux-amd64`
- `bin/vnstat-http-server-linux-arm64`

### 2. 运行

```bash
# 基本运行（无鉴权）
./bin/vnstat-http-server-linux-amd64 -port 8080

# 启用 Token 鉴权
./bin/vnstat-http-server-linux-amd64 -port 8080 -token your-secret-token

# 指定网卡接口
./bin/vnstat-http-server-linux-amd64 -port 8080 -token your-secret-token -interface eth0
```

### 3. 命令行参数

- `-port`: 监听端口，默认 `8080`
- `-token`: 访问鉴权 Token，默认为空（即不开启鉴权）
- `-interface`: （可选）指定强制查询的网卡接口，默认为空（查询所有）

## API 接口

所有接口都支持 CORS 跨域请求，并且可以通过查询参数 `?token=YOUR_TOKEN` 进行鉴权（如果启用了 Token）。

### 1. 获取 JSON 数据

**接口**: `GET /json`

**描述**: 返回 vnstat 的完整 JSON 数据，包含所有统计信息

**参数**:
- `token` (可选): 如果启用了鉴权，需要传递此参数

**响应**: `Content-Type: application/json`

**示例**:
```bash
curl http://localhost:8080/json?token=your-secret-token
```

### 2. 文本视图接口

以下接口返回 `Content-Type: text/plain; charset=utf-8` 格式的文本数据。

#### 2.1 默认总览视图

**接口**: `GET /summary`

**描述**: 返回 vnstat 的默认总览视图，显示总体统计信息

**示例**:
```bash
curl http://localhost:8080/summary?token=your-secret-token
```

#### 2.2 月度视图

**接口**: `GET /` 或 `GET /monthly`

**描述**: 返回月度流量统计视图（默认接口）

**示例**:
```bash
curl http://localhost:8080/?token=your-secret-token
```

#### 2.3 日视图

**接口**: `GET /daily`

**描述**: 返回每日流量统计视图

**示例**:
```bash
curl http://localhost:8080/daily?token=your-secret-token
```

#### 2.4 小时视图

**接口**: `GET /hourly`

**描述**: 返回每小时流量统计视图

**示例**:
```bash
curl http://localhost:8080/hourly?token=your-secret-token
```

#### 2.5 周视图

**接口**: `GET /weekly`

**描述**: 返回每周流量统计视图

**示例**:
```bash
curl http://localhost:8080/weekly?token=your-secret-token
```

#### 2.6 年视图

**接口**: `GET /yearly`

**描述**: 返回年度流量统计视图

**示例**:
```bash
curl http://localhost:8080/yearly?token=your-secret-token
```

#### 2.7 顶部流量接口

**接口**: `GET /top`

**描述**: 返回流量最高的网络接口列表

**示例**:
```bash
curl http://localhost:8080/top?token=your-secret-token
```

#### 2.8 单行输出

**接口**: `GET /oneline`

**描述**: 返回单行格式的简洁输出，适合脚本解析

**示例**:
```bash
curl http://localhost:8080/oneline?token=your-secret-token
```

### 3. 健康检查

**接口**: `GET /health`

**描述**: 健康检查接口，无需鉴权

**响应**: `Content-Type: application/json`

**示例**:
```bash
curl http://localhost:8080/health
```

**响应示例**:
```json
{
  "status": "ok"
}
```

## 接口功能说明

| 接口 | 功能 | 输出格式 | 用途 |
|------|------|----------|------|
| `/json` | 完整 JSON 数据 | JSON | API 集成、数据分析 |
| `/summary` | 默认总览 | 文本 | 快速查看总体情况 |
| `/daily` | 日统计 | 文本 | 查看每日流量趋势 |
| `/hourly` | 小时统计 | 文本 | 查看每小时流量变化 |
| `/weekly` | 周统计 | 文本 | 查看每周流量趋势 |
| `/` 或 `/monthly` | 月统计 | 文本 | 查看每月流量统计 |
| `/yearly` | 年统计 | 文本 | 查看年度流量汇总 |
| `/top` | 顶部接口 | 文本 | 查看流量最高的接口 |
| `/oneline` | 单行输出 | 文本 | 脚本解析、监控告警 |

## iOS Scriptable Widget

项目包含一个专为 iOS Scriptable 设计的 Widget 脚本，可以在 iPhone 主屏幕上以 4x4 小尺寸显示服务器流量统计。

### 快速开始

1. 从 App Store 安装 [Scriptable](https://apps.apple.com/app/scriptable/id1405459188)
2. 在 Scriptable 中创建新脚本，复制 `vnstat-widget.js` 的内容
3. 修改脚本中的 `SERVER_URL` 和 `TOKEN` 配置
4. 在 iPhone 主屏幕添加 Scriptable Widget（选择 Small 尺寸）
5. 选择你创建的脚本

### 详细文档

完整的使用说明、配置选项和故障排除请参考：[SCRIPTABLE_WIDGET.md](./SCRIPTABLE_WIDGET.md)

### Widget 特性

- 📱 完美适配 4x4 Widget 尺寸
- 🎨 自动适配深色/浅色模式
- 📊 显示今日、本月流量和月度使用进度
- 📈 可视化进度条，支持半格填充
- 🔄 可配置刷新间隔（默认 5 分钟）
- ⚡ 快速响应，10 秒超时

## Systemd 服务配置

1. 将编译好的二进制文件复制到系统目录：
```bash
sudo cp bin/vnstat-http-server-linux-amd64 /usr/local/bin/vnstat-http-server
sudo chmod +x /usr/local/bin/vnstat-http-server
```

2. 复制服务配置文件：
```bash
sudo cp vnstat-server.service /etc/systemd/system/
```

3. 编辑服务配置文件，修改 `ExecStart` 路径和参数：
```bash
sudo nano /etc/systemd/system/vnstat-server.service
```

4. 启动服务：
```bash
sudo systemctl daemon-reload
sudo systemctl enable vnstat-server
sudo systemctl start vnstat-server
```

5. 查看服务状态：
```bash
sudo systemctl status vnstat-server
```

## 项目结构

```
vnstat-http-server/
├── main.go           # 主程序逻辑
├── handler.go        # HTTP 处理函数
├── service.go           # 执行 vnstat 命令的封装
├── go.mod            # Go Module 文件
├── Makefile          # 包含 build 命令
├── README.md         # 项目说明文档（英文）
├── README.zh-CN.md   # 项目说明文档（中文）
└── vnstat-server.service # Systemd 服务配置文件模板
```

## 开发

### 本地开发

```bash
# 运行程序
go run . -port 8080 -token test123

# 编译当前平台
go build -o vnstat-http-server .
```

### 测试

```bash
# 测试健康检查
curl http://localhost:8080/health

# 测试 JSON 接口
curl http://localhost:8080/json?token=test123

# 测试文本接口
curl http://localhost:8080/?token=test123
```

## 安全建议

1. **生产环境必须启用 Token 鉴权**，避免数据被未授权访问
2. 使用防火墙限制访问来源
3. 定期更换 Token
4. 考虑使用 HTTPS（可通过反向代理实现，如 Nginx）

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

