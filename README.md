# vnstat-http-server

A lightweight single-file tool written in Go that exposes vnstat statistics from Linux servers via HTTP API, enabling remote monitoring by mobile apps, web frontends, or scripts.

## Features

- 🚀 **Zero Dependencies**: No need to install Python, PHP, Node, or Docker (except vnstat itself)
- 📦 **Single Binary**: Just one file after compilation, ready to run
- 🔒 **Secure**: Simple token-based authentication support
- 🌐 **CORS Support**: All endpoints support cross-origin requests
- 📊 **Multiple Formats**: Supports both JSON and plain text output

## Requirements

- Linux system (amd64 / arm64)
- `vnstat` tool installed
- Go 1.21+ (only needed for compilation)

## Quick Start

### 1. Build

```bash
make build
```

After compilation, binary files are located in the `bin/` directory:
- `bin/vnstat-http-server-linux-amd64`
- `bin/vnstat-http-server-linux-arm64`

### 2. Run

```bash
# Basic run (no authentication)
./bin/vnstat-http-server-linux-amd64 -port 8080

# Enable token authentication
./bin/vnstat-http-server-linux-amd64 -port 8080 -token your-secret-token

# Specify network interface
./bin/vnstat-http-server-linux-amd64 -port 8080 -token your-secret-token -interface eth0
```

### 3. Command Line Arguments

- `-port`: Listening port, default `8080`
- `-token`: Authentication token, default empty (no authentication)
- `-interface`: (Optional) Specify network interface name, default empty (query all)

## API Endpoints

All endpoints support CORS cross-origin requests and can be authenticated via query parameter `?token=YOUR_TOKEN` (if token is enabled).

### 1. Get JSON Data

**Endpoint**: `GET /json`

**Description**: Returns complete vnstat JSON data with all statistics

**Parameters**:
- `token` (optional): Required if authentication is enabled

**Response**: `Content-Type: application/json`

**Example**:
```bash
curl http://localhost:8080/json?token=your-secret-token
```

### 2. Text View Endpoints

The following endpoints return `Content-Type: text/plain; charset=utf-8` formatted text data.

#### 2.1 Default Summary View

**Endpoint**: `GET /summary`

**Description**: Returns vnstat's default summary view showing overall statistics

**Example**:
```bash
curl http://localhost:8080/summary?token=your-secret-token
```

#### 2.2 Monthly View

**Endpoint**: `GET /` or `GET /monthly`

**Description**: Returns monthly traffic statistics view (default endpoint)

**Example**:
```bash
curl http://localhost:8080/?token=your-secret-token
```

#### 2.3 Daily View

**Endpoint**: `GET /daily`

**Description**: Returns daily traffic statistics view

**Example**:
```bash
curl http://localhost:8080/daily?token=your-secret-token
```

#### 2.4 Hourly View

**Endpoint**: `GET /hourly`

**Description**: Returns hourly traffic statistics view

**Example**:
```bash
curl http://localhost:8080/hourly?token=your-secret-token
```

#### 2.5 Weekly View

**Endpoint**: `GET /weekly`

**Description**: Returns weekly traffic statistics view

**Example**:
```bash
curl http://localhost:8080/weekly?token=your-secret-token
```

#### 2.6 Yearly View

**Endpoint**: `GET /yearly`

**Description**: Returns yearly traffic statistics view

**Example**:
```bash
curl http://localhost:8080/yearly?token=your-secret-token
```

#### 2.7 Top Traffic Interfaces

**Endpoint**: `GET /top`

**Description**: Returns list of top traffic network interfaces

**Example**:
```bash
curl http://localhost:8080/top?token=your-secret-token
```

#### 2.8 One-line Output

**Endpoint**: `GET /oneline`

**Description**: Returns concise one-line format output, suitable for script parsing

**Example**:
```bash
curl http://localhost:8080/oneline?token=your-secret-token
```

### 3. Health Check

**Endpoint**: `GET /health`

**Description**: Health check endpoint, no authentication required

**Response**: `Content-Type: application/json`

**Example**:
```bash
curl http://localhost:8080/health
```

**Response Example**:
```json
{
  "status": "ok"
}
```

## Endpoint Summary

| Endpoint | Function | Output Format | Use Case |
|----------|----------|---------------|----------|
| `/json` | Complete JSON data | JSON | API integration, data analysis |
| `/summary` | Default summary | Text | Quick overview |
| `/daily` | Daily statistics | Text | Daily traffic trends |
| `/hourly` | Hourly statistics | Text | Hourly traffic changes |
| `/weekly` | Weekly statistics | Text | Weekly traffic trends |
| `/` or `/monthly` | Monthly statistics | Text | Monthly traffic statistics |
| `/yearly` | Yearly statistics | Text | Annual traffic summary |
| `/top` | Top interfaces | Text | Highest traffic interfaces |
| `/oneline` | One-line output | Text | Script parsing, monitoring alerts |

## iOS Scriptable Widget

The project includes a Widget script designed for iOS Scriptable, which can display server traffic statistics on iPhone home screen in 4x4 small size.

### Quick Start

1. Install [Scriptable](https://apps.apple.com/app/scriptable/id1405459188) from App Store
2. Create a new script in Scriptable and copy the content of `vnstat-widget.js`
3. Modify `SERVER_URL` and `TOKEN` configuration in the script
4. Add Scriptable Widget to iPhone home screen (select Small size)
5. Select your created script

### Detailed Documentation

For complete usage instructions, configuration options, and troubleshooting, please refer to: [SCRIPTABLE_WIDGET.md](./SCRIPTABLE_WIDGET.md)

### Widget Features

- 📱 Perfect fit for 4x4 Widget size
- 🎨 Auto-adapts to dark/light mode
- 📊 Displays today, monthly traffic and monthly usage progress
- 📈 Visual progress bar with half-fill support
- 🔄 Configurable refresh interval (default 5 minutes)
- ⚡ Fast response, 10 second timeout

## Systemd Service Configuration

1. Copy the compiled binary to system directory:
```bash
sudo cp bin/vnstat-http-server-linux-amd64 /usr/local/bin/vnstat-http-server
sudo chmod +x /usr/local/bin/vnstat-http-server
```

2. Copy service configuration file:
```bash
sudo cp vnstat-server.service /etc/systemd/system/
```

3. Edit service configuration file, modify `ExecStart` path and parameters:
```bash
sudo nano /etc/systemd/system/vnstat-server.service
```

4. Start the service:
```bash
sudo systemctl daemon-reload
sudo systemctl enable vnstat-server
sudo systemctl start vnstat-server
```

5. Check service status:
```bash
sudo systemctl status vnstat-server
```

## Project Structure

```
vnstat-http-server/
├── main.go           # Main program logic
├── handler.go        # HTTP handler functions
├── service.go        # vnstat command execution wrapper
├── go.mod            # Go Module file
├── Makefile          # Build commands
├── README.md         # Project documentation (English)
├── README.zh-CN.md   # Project documentation (Chinese)
└── vnstat-server.service # Systemd service configuration template
```

## Development

### Local Development

```bash
# Run the program
go run . -port 8080 -token test123

# Build for current platform
go build -o vnstat-http-server .
```

### Testing

```bash
# Test health check
curl http://localhost:8080/health

# Test JSON endpoint
curl http://localhost:8080/json?token=test123

# Test text endpoint
curl http://localhost:8080/?token=test123
```

## Security Recommendations

1. **Enable token authentication in production** to prevent unauthorized access
2. Use firewall to restrict access sources
3. Regularly rotate tokens
4. Consider using HTTPS (can be implemented via reverse proxy like Nginx)

## License

MIT License

## Contributing

Issues and Pull Requests are welcome!
