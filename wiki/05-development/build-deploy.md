# 构建部署

## 环境要求

- **Go**: 1.24.2+
- **Git**: 用于依赖管理

## 获取源码

```bash
git clone https://github.com/tlslink/sslcon.git
cd sslcon
```

## 依赖管理

```bash
# 下载依赖
go mod download

# 整理依赖
go mod tidy
```

---

## 构建

### 构建 CLI 工具

```bash
# 当前平台
go build -o sslcon ./sslcon.go

# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o sslcon-linux-amd64 ./sslcon.go

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o sslcon-linux-arm64 ./sslcon.go

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o sslcon-windows-amd64.exe ./sslcon.go

# macOS AMD64
GOOS=darwin GOARCH=amd64 go build -o sslcon-darwin-amd64 ./sslcon.go

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o sslcon-darwin-arm64 ./sslcon.go
```

### 构建 VPN Agent

```bash
# 当前平台
go build -o vpnagent ./vpnagent.go

# Linux AMD64
GOOS=linux GOARCH=amd64 go build -o vpnagent-linux-amd64 ./vpnagent.go

# Windows AMD64
GOOS=windows GOARCH=amd64 go build -o vpnagent-windows-amd64.exe ./vpnagent.go

# macOS ARM64
GOOS=darwin GOARCH=arm64 go build -o vpnagent-darwin-arm64 ./vpnagent.go
```

### 优化构建

```bash
# 减小二进制大小
go build -ldflags="-s -w" -o sslcon ./sslcon.go

# 静态链接 (Linux)
CGO_ENABLED=0 go build -o sslcon ./sslcon.go
```

---

## 部署

### Linux (systemd)

1. 复制可执行文件：
```bash
sudo cp vpnagent /usr/local/bin/
sudo cp sslcon /usr/local/bin/
sudo chmod +x /usr/local/bin/vpnagent
sudo chmod +x /usr/local/bin/sslcon
```

2. 安装服务：
```bash
sudo /usr/local/bin/vpnagent install
```

3. 验证服务：
```bash
sudo systemctl status sslcon.service
```

### Linux (OpenWrt)

1. 上传到路由器：
```bash
scp vpnagent root@192.168.1.1:/usr/bin/
```

2. 设置权限：
```bash
chmod +x /usr/bin/vpnagent
```

3. 安装服务：
```bash
/usr/bin/vpnagent install
```

### Windows

1. 下载 WinTun 驱动（如需要）

2. 以管理员身份运行：
```powershell
.\vpnagent.exe install
```

3. 验证服务：
```powershell
sc query SSLCon
```

### macOS

1. 复制可执行文件：
```bash
sudo cp vpnagent /usr/local/bin/
sudo cp sslcon /usr/local/bin/
```

2. 安装服务：
```bash
sudo /usr/local/bin/vpnagent install
```

---

## 卸载

### Linux (systemd)

```bash
sudo /usr/local/bin/vpnagent uninstall
sudo rm /usr/local/bin/vpnagent
sudo rm /usr/local/bin/sslcon
```

### Windows

```powershell
.\vpnagent.exe uninstall
```

### macOS

```bash
sudo /usr/local/bin/vpnagent uninstall
sudo rm /usr/local/bin/vpnagent
sudo rm /usr/local/bin/sslcon
```

---

## 开发构建

### 运行测试

```bash
go test ./...
```

### 运行特定测试

```bash
go test -v ./utils/...
```

### 代码检查

```bash
# 安装 golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# 运行检查
golangci-lint run
```

### 格式化代码

```bash
go fmt ./...
```

---

## 交叉编译注意事项

### CGO 依赖

某些平台可能需要 CGO：

```bash
# 需要 CGO 时
CGO_ENABLED=1 go build ...

# 禁用 CGO
CGO_ENABLED=0 go build ...
```

### Windows 构建

Windows 构建需要 WinTun 驱动支持，首次运行时会自动安装。

### macOS 签名

生产环境部署可能需要代码签名：

```bash
codesign -s "Developer ID" vpnagent
```

---

## 发布清单

1. 更新版本号
2. 运行测试
3. 多平台构建
4. 创建 Release
5. 上传构建产物
