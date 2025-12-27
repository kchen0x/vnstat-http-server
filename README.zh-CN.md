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

### 1. 下载预编译二进制文件

预编译的二进制文件可在 [Releases](https://github.com/kchen0x/vnstat-http-server/releases) 页面下载。

**最新版本**: [v0.1](https://github.com/kchen0x/vnstat-http-server/releases/tag/v0.1)

根据你的系统下载对应的二进制文件：
- `vnstat-http-server-linux-amd64` - 适用于 Linux x86_64 系统
- `vnstat-http-server-linux-arm64` - 适用于 Linux ARM64 系统

下载后，赋予执行权限：
```bash
chmod +x vnstat-http-server-linux-amd64
```

### 2. 从源码编译

如果你希望从源码编译：

```bash
make build
```

编译完成后，二进制文件位于 `bin/` 目录：
- `bin/vnstat-http-server-linux-amd64`
- `bin/vnstat-http-server-linux-arm64`

### 3. 运行

```bash
# 基本运行（无鉴权）
./bin/vnstat-http-server-linux-amd64 -port 8080

# 启用 Token 鉴权
./bin/vnstat-http-server-linux-amd64 -port 8080 -token your-secret-token

# 指定网卡接口
./bin/vnstat-http-server-linux-amd64 -port 8080 -token your-secret-token -interface eth0
```

### 4. 命令行参数

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

### 3. Prometheus Metrics

**接口**: `GET /metrics`

**描述**: 返回 Prometheus 格式的指标数据，用于与 Grafana Cloud、Prometheus 等监控系统集成

**响应**: `Content-Type: text/plain; version=0.0.4; charset=utf-8`

**提供的指标**:
- `vnstat_traffic_total_bytes{interface="<name>",direction="rx|tx"}` - 总流量（字节）
- `vnstat_traffic_month_bytes{interface="<name>",direction="rx|tx"}` - 月度流量（字节）
- `vnstat_traffic_today_bytes{interface="<name>",direction="rx|tx"}` - 今日流量（字节）

**示例**:
```bash
# 无鉴权（如果未设置 token）
curl http://localhost:8080/metrics

# 有鉴权（如果设置了 token）
curl http://localhost:8080/metrics?token=your-secret-token
```

**输出示例**:
```
# HELP vnstat_traffic_total_bytes Total traffic in bytes
# TYPE vnstat_traffic_total_bytes counter
vnstat_traffic_total_bytes{interface="eth0",direction="rx"} 1234567890
vnstat_traffic_total_bytes{interface="eth0",direction="tx"} 987654321
vnstat_traffic_month_bytes{interface="eth0",direction="rx"} 123456789
vnstat_traffic_month_bytes{interface="eth0",direction="tx"} 98765432
vnstat_traffic_today_bytes{interface="eth0",direction="rx"} 1234567
vnstat_traffic_today_bytes{interface="eth0",direction="tx"} 987654
```

### 4. 健康检查

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
| `/metrics` | Prometheus 指标 | Prometheus | Grafana Cloud、Prometheus 集成 |
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

## Grafana Cloud 集成

`/metrics` 接口提供 Prometheus 格式的指标数据，可以轻松与 Grafana Cloud 集成。

### 方案 1：使用 Grafana Agent（推荐）

1. **在服务器上安装 Grafana Agent**：
   ```bash
   # Linux 系统
   curl -O -L "https://github.com/grafana/agent/releases/latest/download/grafana-agent-linux-amd64.zip"
   unzip grafana-agent-linux-amd64.zip
   sudo mv grafana-agent-linux-amd64 /usr/local/bin/grafana-agent
   sudo chmod +x /usr/local/bin/grafana-agent
   ```

2. **创建 Grafana Agent 配置文件** (`/etc/grafana-agent/config.yaml`)：
   ```yaml
   metrics:
     configs:
       - name: vnstat
         remote_write:
           - url: https://prometheus-prod-01-eu-west-0.grafana.net/api/prom/push
             basic_auth:
               username: YOUR_INSTANCE_ID
               password: YOUR_API_TOKEN
         scrape_configs:
           - job_name: 'vnstat'
             static_configs:
               - targets: ['localhost:8080']
             metrics_path: '/metrics'
             scrape_interval: 30s
             params:
               token: ['your-vnstat-token']  # 如果启用了 token
   ```

3. **启动 Grafana Agent**：
   ```bash
   sudo grafana-agent --config.file=/etc/grafana-agent/config.yaml
   ```

### 方案 2：使用 Prometheus Remote Write

如果你正在运行 Prometheus，可以配置它抓取 `/metrics` 接口并远程写入到 Grafana Cloud：

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'vnstat'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    params:
      token: ['your-vnstat-token']  # 如果启用了 token

remote_write:
  - url: https://prometheus-prod-01-eu-west-0.grafana.net/api/prom/push
    basic_auth:
      username: YOUR_INSTANCE_ID
      password: YOUR_API_TOKEN
```

### 方案 3：直接 HTTP 推送（高级）

你也可以创建一个脚本，定期将指标推送到 Grafana Cloud 的 Prometheus remote write API。

### 在 Grafana 中创建仪表盘

一旦指标数据流入 Grafana Cloud，你可以使用以下查询创建仪表盘：

- **总流量**: `sum(vnstat_traffic_total_bytes)`
- **月度流量**: `sum(vnstat_traffic_month_bytes)`
- **今日流量**: `sum(vnstat_traffic_today_bytes)`
- **按接口**: `vnstat_traffic_total_bytes{interface="eth0"}`
- **上传 vs 下载**: 
  - 上传: `sum(vnstat_traffic_total_bytes{direction="tx"})`
  - 下载: `sum(vnstat_traffic_total_bytes{direction="rx"})`

### 获取 Grafana Cloud 凭证

1. 登录 [Grafana Cloud](https://grafana.com/auth/sign-up/create-user)
2. 进入 **My Account** → **Prometheus** → **Details**
3. 复制你的 **Instance ID**（用户名）和 **API Token**（密码）
4. 在 Grafana Agent 或 Prometheus 配置中使用这些凭证

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

