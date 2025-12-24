# RPC 服务模块 (rpc)

## 概述

`rpc/` 模块提供 WebSocket JSON-RPC 2.0 接口，允许外部应用（如 GUI 客户端）控制 VPN 连接。

## 文件结构

```
rpc/
├── connect.go   # 连接/断开逻辑
└── rpc.go       # RPC 服务实现
```

## 服务端点

- **地址**: `ws://127.0.0.1:6210/rpc`
- **协议**: WebSocket + JSON-RPC 2.0

## 方法定义

### 方法 ID 常量

```go
const (
    STATUS     = iota  // 0 - 获取状态
    CONFIG            // 1 - 配置
    CONNECT           // 2 - 连接
    DISCONNECT        // 3 - 断开
    RECONNECT         // 4 - 重连
    INTERFACE         // 5 - 设置接口
    ABORT             // 6 - 异常断开通知
    STAT              // 7 - 流量统计
)
```

## API 详解

### 1. STATUS (ID=0)

获取当前 VPN 连接状态。

**请求：**
```json
{
    "jsonrpc": "2.0",
    "method": "status",
    "id": 0
}
```

**响应（已连接）：**
```json
{
    "jsonrpc": "2.0",
    "id": 0,
    "result": {
        "ServerAddress": "192.168.1.1",
        "LocalAddress": "192.168.1.100",
        "Hostname": "vpn.example.com",
        "TunName": "SSLCon",
        "VPNAddress": "10.0.0.2",
        "VPNMask": "255.255.255.0",
        "DNS": ["8.8.8.8"],
        "MTU": 1399,
        "TLSCipherSuite": "TLS_AES_256_GCM_SHA384",
        "DTLSCipherSuite": "ECDHE-RSA-AES256-GCM-SHA384"
    }
}
```

**响应（未连接）：**
```json
{
    "jsonrpc": "2.0",
    "id": 0,
    "error": {"code": 1, "message": "disconnected from ..."}
}
```

---

### 2. CONFIG (ID=1)

配置客户端参数。

**请求：**
```json
{
    "jsonrpc": "2.0",
    "method": "config",
    "params": {
        "log_level": "Debug",
        "log_path": "/var/log/vpn",
        "skip_verify": true,
        "cisco_compat": true,
        "no_dtls": false,
        "agent_name": "AnyConnect",
        "agent_version": "4.10.07062"
    },
    "id": 1
}
```

**参数说明：**

| 参数 | 类型 | 说明 |
|------|------|------|
| `log_level` | string | 日志级别: Debug/Info/Warn/Error/Fatal |
| `log_path` | string | 日志目录，空则输出到 stdout |
| `skip_verify` | bool | 跳过 TLS 证书验证 |
| `cisco_compat` | bool | Cisco 兼容模式 |
| `no_dtls` | bool | 禁用 DTLS |
| `agent_name` | string | 客户端名称 |
| `agent_version` | string | 客户端版本 |

---

### 3. CONNECT (ID=2)

连接到 VPN 服务器。

**请求：**
```json
{
    "jsonrpc": "2.0",
    "method": "connect",
    "params": {
        "host": "vpn.example.com",
        "username": "user",
        "password": "pass",
        "group": "default",
        "secret": ""
    },
    "id": 2
}
```

**参数说明：**

| 参数 | 类型 | 说明 |
|------|------|------|
| `host` | string | 服务器地址（可带端口） |
| `username` | string | 用户名 |
| `password` | string | 密码 |
| `group` | string | 用户组 |
| `secret` | string | 密钥（可选） |

---

### 4. DISCONNECT (ID=3)

断开 VPN 连接。

**请求：**
```json
{
    "jsonrpc": "2.0",
    "method": "disconnect",
    "id": 3
}
```

---

### 5. RECONNECT (ID=4)

重新建立 VPN 隧道（保持认证状态）。

**请求：**
```json
{
    "jsonrpc": "2.0",
    "method": "reconnect",
    "id": 4
}
```

---

### 6. INTERFACE (ID=5)

设置本地网络接口信息。

**请求：**
```json
{
    "jsonrpc": "2.0",
    "method": "interface",
    "params": {
        "name": "eth0",
        "ip4": "192.168.1.100",
        "mac": "00:11:22:33:44:55",
        "gateway": "192.168.1.1"
    },
    "id": 5
}
```

---

### 7. STAT (ID=7)

获取流量统计。

**请求：**
```json
{
    "jsonrpc": "2.0",
    "method": "stat",
    "id": 7
}
```

**响应：**
```json
{
    "jsonrpc": "2.0",
    "id": 7,
    "result": {
        "bytesSent": 123456,
        "bytesReceived": 654321
    }
}
```

## 服务端推送

### ABORT (ID=6)

服务端异常断开时推送给所有客户端：

```json
{
    "jsonrpc": "2.0",
    "id": 6,
    "result": "disconnected from vpn.example.com"
}
```

### DISCONNECT (ID=3)

用户主动断开时推送：

```json
{
    "jsonrpc": "2.0",
    "id": 3,
    "result": "disconnected from vpn.example.com"
}
```

## 连接管理

### Connect()

```go
func Connect() error
```

完整连接流程：
1. 处理主机端口
2. 获取本地网络接口（如未设置）
3. 初始化认证
4. 密码认证
5. 建立隧道

### SetupTunnel()

```go
func SetupTunnel(reconnect bool) error
```

建立隧道，支持断线重连模式。

### DisConnect()

```go
func DisConnect()
```

主动断开连接，重置路由并关闭会话。

## 客户端管理

```go
var Clients []*jsonrpc2.Conn  // 已连接的客户端列表
```

服务端维护所有 WebSocket 客户端连接，断开时从列表移除。

## 监控协程

```go
func monitor()
```

监听 `session.Sess.CloseChan`，连接断开时通知所有客户端。
