# JSON-RPC API

## 概述

vpnagent 通过 WebSocket 提供 JSON-RPC 2.0 API，允许外部应用控制 VPN 连接。

## 连接信息

- **端点**: `ws://127.0.0.1:6210/rpc`
- **协议**: WebSocket + JSON-RPC 2.0

## 请求格式

```json
{
    "jsonrpc": "2.0",
    "method": "<method_name>",
    "params": { ... },
    "id": <method_id>
}
```

## 响应格式

### 成功响应

```json
{
    "jsonrpc": "2.0",
    "id": <method_id>,
    "result": <result_data>
}
```

### 错误响应

```json
{
    "jsonrpc": "2.0",
    "id": <method_id>,
    "error": {
        "code": 1,
        "message": "<error_message>"
    }
}
```

---

## API 方法

### status (ID=0)

获取当前 VPN 连接状态。

**请求：**
```json
{
    "jsonrpc": "2.0",
    "method": "status",
    "id": 0
}
```

**成功响应：**
```json
{
    "jsonrpc": "2.0",
    "id": 0,
    "result": {
        "ServerAddress": "203.0.113.1",
        "LocalAddress": "192.168.1.100",
        "Hostname": "vpn.example.com",
        "TunName": "SSLCon",
        "VPNAddress": "10.0.0.2",
        "VPNMask": "255.255.255.0",
        "DNS": ["8.8.8.8", "8.8.4.4"],
        "MTU": 1399,
        "SplitInclude": ["10.0.0.0/255.255.0.0"],
        "SplitExclude": [],
        "DynamicSplitTunneling": false,
        "TLSCipherSuite": "TLS_AES_256_GCM_SHA384",
        "TLSDpdTime": 30,
        "TLSKeepaliveTime": 20,
        "DTLSPort": "443",
        "DTLSDpdTime": 30,
        "DTLSKeepaliveTime": 20,
        "DTLSCipherSuite": "ECDHE-RSA-AES256-GCM-SHA384"
    }
}
```

**响应字段说明：**

| 字段 | 类型 | 说明 |
|------|------|------|
| `ServerAddress` | string | VPN 服务器 IP |
| `LocalAddress` | string | 本地 IP |
| `Hostname` | string | 服务器主机名 |
| `TunName` | string | TUN 设备名 |
| `VPNAddress` | string | VPN 分配 IP |
| `VPNMask` | string | VPN 子网掩码 |
| `DNS` | []string | DNS 服务器列表 |
| `MTU` | int | MTU 值 |
| `SplitInclude` | []string | 包含路由 |
| `SplitExclude` | []string | 排除路由 |
| `TLSCipherSuite` | string | TLS 加密套件 |
| `DTLSCipherSuite` | string | DTLS 加密套件 |

---

### config (ID=1)

配置客户端参数。应在 connect 之前调用。

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
        "agent_name": "",
        "agent_version": "4.10.07062"
    },
    "id": 1
}
```

**参数说明：**

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `log_level` | string | `Debug` | Debug/Info/Warn/Error/Fatal |
| `log_path` | string | `""` | 日志目录，空则输出到控制台 |
| `skip_verify` | bool | `true` | 跳过 TLS 证书验证 |
| `cisco_compat` | bool | `true` | Cisco 兼容模式 |
| `no_dtls` | bool | `false` | 禁用 DTLS |
| `agent_name` | string | `""` | 客户端名称 |
| `agent_version` | string | `4.10.07062` | 客户端版本 |

---

### connect (ID=2)

连接到 VPN 服务器。

**请求：**
```json
{
    "jsonrpc": "2.0",
    "method": "connect",
    "params": {
        "host": "vpn.example.com",
        "username": "user",
        "password": "password123",
        "group": "default",
        "secret": ""
    },
    "id": 2
}
```

**参数说明：**

| 参数 | 类型 | 必需 | 说明 |
|------|------|------|------|
| `host` | string | ✓ | 服务器地址，可带端口 |
| `username` | string | ✓ | 用户名 |
| `password` | string | ✓ | 密码 |
| `group` | string | | 用户组 |
| `secret` | string | | 密钥 |

**成功响应：**
```json
{
    "jsonrpc": "2.0",
    "id": 2,
    "result": "connected to vpn.example.com"
}
```

---

### disconnect (ID=3)

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

### reconnect (ID=4)

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

### interface (ID=5)

设置本地网络接口信息。可选，通常由 vpnagent 自动检测。

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

### stat (ID=7)

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
        "bytesSent": 1234567,
        "bytesReceived": 7654321
    }
}
```

---

## 服务端推送

### 主动断开通知 (ID=3)

用户调用 disconnect 后推送：

```json
{
    "jsonrpc": "2.0",
    "id": 3,
    "result": "disconnected from vpn.example.com"
}
```

### 异常断开通知 (ID=6)

网络异常或服务器断开时推送：

```json
{
    "jsonrpc": "2.0",
    "id": 6,
    "result": "disconnected from vpn.example.com"
}
```

---

## 使用示例

### JavaScript (浏览器)

```javascript
const ws = new WebSocket('ws://127.0.0.1:6210/rpc');

ws.onopen = () => {
    // 配置
    ws.send(JSON.stringify({
        jsonrpc: '2.0',
        method: 'config',
        params: { log_level: 'Info' },
        id: 1
    }));
};

ws.onmessage = (event) => {
    const response = JSON.parse(event.data);
    console.log('Response:', response);
    
    if (response.id === 1) {
        // 配置成功，开始连接
        ws.send(JSON.stringify({
            jsonrpc: '2.0',
            method: 'connect',
            params: {
                host: 'vpn.example.com',
                username: 'user',
                password: 'pass'
            },
            id: 2
        }));
    }
};
```

### Python

```python
import asyncio
import websockets
import json

async def vpn_client():
    async with websockets.connect('ws://127.0.0.1:6210/rpc') as ws:
        # 获取状态
        await ws.send(json.dumps({
            'jsonrpc': '2.0',
            'method': 'status',
            'id': 0
        }))
        response = await ws.recv()
        print(json.loads(response))

asyncio.run(vpn_client())
```

---

## 错误码

| 码 | 说明 |
|----|------|
| 1 | 通用错误 |

错误消息在 `error.message` 字段中返回具体原因。
