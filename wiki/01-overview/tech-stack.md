# 技术栈

## Go 版本

- **最低版本**: Go 1.24.2

## 核心依赖

### 网络协议

| 依赖 | 版本 | 用途 |
|------|------|------|
| `github.com/pion/dtls/v3` | v3.0.7 | DTLS 协议实现 |
| `golang.org/x/crypto` | v0.45.0 | TLS 加密支持 |
| `golang.org/x/net` | v0.47.0 | 网络扩展功能 |

### WebSocket & RPC

| 依赖 | 版本 | 用途 |
|------|------|------|
| `github.com/gorilla/websocket` | v1.5.3 | WebSocket 服务 |
| `github.com/sourcegraph/jsonrpc2` | v0.2.1 | JSON-RPC 2.0 实现 |

### 命令行

| 依赖 | 版本 | 用途 |
|------|------|------|
| `github.com/spf13/cobra` | v1.10.1 | CLI 命令框架 |

### 系统服务

| 依赖 | 版本 | 用途 |
|------|------|------|
| `github.com/kardianos/service` | v1.2.4 | 跨平台系统服务 |

### 网络接口

| 依赖 | 版本 | 用途 |
|------|------|------|
| `github.com/vishvananda/netlink` | v1.3.1 | Linux 网络配置 |
| `github.com/jackpal/gateway` | v1.1.1 | 网关发现 |
| `github.com/lysShub/wintun-go` | - | Windows TUN 驱动 |
| `golang.zx2c4.com/wireguard/windows` | v0.5.3 | Windows 网络配置 |

### 数据包处理

| 依赖 | 版本 | 用途 |
|------|------|------|
| `github.com/gopacket/gopacket` | v1.5.0 | 网络数据包解析 |

### 系统信息

| 依赖 | 版本 | 用途 |
|------|------|------|
| `github.com/elastic/go-sysinfo` | v1.15.4 | 系统信息获取 |

### JSON 处理

| 依赖 | 版本 | 用途 |
|------|------|------|
| `github.com/apieasy/gson` | v0.2.3 | JSON 操作 |

### 并发控制

| 依赖 | 版本 | 用途 |
|------|------|------|
| `go.uber.org/atomic` | v1.11.0 | 原子操作封装 |

## 平台特定依赖

### Windows
- `golang.org/x/sys/windows` - Windows 系统调用
- `golang.zx2c4.com/wireguard/windows/tunnel/winipcfg` - Windows IP 配置

### Linux
- `github.com/vishvananda/netlink` - Netlink 套接字操作
- `github.com/vishvananda/netns` - 网络命名空间

### macOS
- 使用 `exec.Command` 调用系统命令配置网络

## 构建标签

```go
//go:build linux || darwin || windows
```

项目使用构建标签实现平台特定代码的条件编译。

## 协议规范参考

| 规范 | 链接 |
|------|------|
| OpenConnect Protocol v04 | https://datatracker.ietf.org/doc/html/draft-mavrogiannopoulos-openconnect-04 |
| OpenConnect Protocol v03 | https://datatracker.ietf.org/doc/html/draft-mavrogiannopoulos-openconnect-03 |
| OpenConnect Protocol v02 | https://datatracker.ietf.org/doc/html/draft-mavrogiannopoulos-openconnect-02 |
| RFC 3706 (DPD) | https://datatracker.ietf.org/doc/html/rfc3706 |
| RFC 8446 (TLS 1.3) | https://datatracker.ietf.org/doc/html/rfc8446 |

## 模块依赖图

```
go.mod
├── github.com/apieasy/gson v0.2.3
├── github.com/elastic/go-sysinfo v1.15.4
├── github.com/gopacket/gopacket v1.5.0
├── github.com/gorilla/websocket v1.5.3
├── github.com/jackpal/gateway v1.1.1
├── github.com/kardianos/service v1.2.4
├── github.com/lysShub/wintun-go
├── github.com/pion/dtls/v3 v3.0.7
├── github.com/sourcegraph/jsonrpc2 v0.2.1
├── github.com/spf13/cobra v1.10.1
├── github.com/vishvananda/netlink v1.3.1
├── go.uber.org/atomic v1.11.0
├── golang.org/x/crypto v0.45.0
├── golang.org/x/net v0.47.0
├── golang.org/x/sys v0.38.0
└── golang.zx2c4.com/wireguard/windows v0.5.3
```
