# VPN 隧道模块 (vpn)

## 概述

`vpn/` 模块实现 VPN 数据隧道，包括 TLS (CSTP) 和 DTLS 两种传输通道。

## 文件结构

```
vpn/
├── buffer.go   # 数据包缓冲池
├── dtls.go     # DTLS 通道实现
├── tls.go      # TLS 通道实现
├── tun.go      # TUN 设备数据处理
└── tunnel.go   # 隧道建立入口
```

## 核心流程

### 隧道建立 (tunnel.go)

#### SetupTunnel()

建立 VPN 隧道的主入口函数：

```go
func SetupTunnel() error
```

**流程：**
1. 初始化请求头（Cookie、Master Secret 等）
2. 发送 HTTP CONNECT 请求到 `/CSCOSSLC/tunnel`
3. 解析响应头创建 `ConnSession`
4. 创建并配置 TUN 设备
5. 设置路由规则
6. 启动 TLS 通道
7. 启动 DTLS 通道（可选）
8. 启动 DPD 和读超时定时器

**请求头：**

| Header | 说明 |
|--------|------|
| `Cookie` | `webvpn=<SessionToken>` |
| `X-CSTP-VPNAddress-Type` | `IPv4` |
| `X-CSTP-MTU` | `1399` |
| `X-CSTP-Base-MTU` | `1399` |
| `X-CSTP-Local-VPNAddress-IP4` | 本地 IP |
| `X-DTLS-Master-Secret` | DTLS 预主密钥（hex） |
| `X-DTLS12-CipherSuite` | 支持的 DTLS 加密套件 |

---

### TLS 通道 (tls.go)

#### tlsChannel()

TLS 数据通道，从服务器读取数据放入 `PayloadIn`：

```go
func tlsChannel(conn *tls.Conn, bufR *bufio.Reader, 
                cSess *session.ConnSession, resp *http.Response)
```

**数据包处理：**

| Type | 名称 | 处理 |
|------|------|------|
| `0x00` | DATA | 去除头部，放入 PayloadIn |
| `0x03` | DPD-REQ | 回复 DPD-RESP (0x04) |
| `0x04` | DPD-RESP | 记录日志 |

#### payloadOutTLSToServer()

从 `PayloadOutTLS` 读取数据发送到服务器：

```go
func payloadOutTLSToServer(conn *tls.Conn, cSess *session.ConnSession)
```

**数据包格式：**

```
+------+------+------+------+------+------+------+------+--------+
| 0x53 | 0x54 | 0x46 | 0x01 | Len(H)| Len(L)| Type | 0x00 | Data   |
+------+------+------+------+------+------+------+------+--------+
  'S'    'T'    'F'   固定    长度（大端）   类型   固定   数据
```

---

### DTLS 通道 (dtls.go)

#### dtlsChannel()

DTLS 数据通道，使用 UDP 提供低延迟传输：

```go
func dtlsChannel(cSess *session.ConnSession)
```

**DTLS 配置：**

```go
config := &dtls.Config{
    InsecureSkipVerify:   true,
    ExtendedMasterSecret: dtls.DisableExtendedMasterSecret,
    CipherSuites:         // 根据服务器协商选择
    SessionStore:         // 使用预主密钥
}
```

**支持的加密套件：**

| 套件 | DTLS ID |
|------|---------|
| ECDHE-ECDSA-AES128-GCM-SHA256 | `TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256` |
| ECDHE-RSA-AES128-GCM-SHA256 | `TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256` |
| ECDHE-ECDSA-AES256-GCM-SHA384 | `TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384` |
| ECDHE-RSA-AES256-GCM-SHA384 | `TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384` |

#### payloadOutDTLSToServer()

从 `PayloadOutDTLS` 读取数据发送到服务器：

```go
func payloadOutDTLSToServer(conn *dtls.Conn, dSess *session.DtlsSession, 
                            cSess *session.ConnSession)
```

**DTLS 数据包格式（简化版，仅 1 字节头）：**

```
+------+--------+
| Type | Data   |
+------+--------+
```

---

### TUN 设备处理 (tun.go)

#### setupTun()

创建和配置 TUN 虚拟网卡：

```go
func setupTun(cSess *session.ConnSession) error
```

**平台设备名：**

| 平台 | 设备名 |
|------|--------|
| Windows | `SSLCon` |
| macOS | `utun` |
| Linux | `sslcon` |

#### tunToPayloadOut()

从 TUN 设备读取数据，放入 `PayloadOutTLS` 或 `PayloadOutDTLS`：

```go
func tunToPayloadOut(dev tun.Device, cSess *session.ConnSession)
```

优先使用 DTLS 通道（如果已连接）。

#### payloadInToTun()

从 `PayloadIn` 读取数据，写入 TUN 设备：

```go
func payloadInToTun(dev tun.Device, cSess *session.ConnSession)
```

支持动态分流：检测 DNS 响应包，根据域名匹配规则动态添加路由。

#### dynamicSplitRoutes()

处理 DNS 响应实现动态路由分流：

```go
func dynamicSplitRoutes(data []byte, cSess *session.ConnSession)
```

---

### 缓冲池 (buffer.go)

使用 `sync.Pool` 复用数据包缓冲区：

```go
func getPayloadBuffer() *proto.Payload   // 获取缓冲区
func putPayloadBuffer(pl *proto.Payload) // 归还缓冲区
```

## 数据流图

```
应用程序
    │
    ▼
┌─────────┐    tunToPayloadOut     ┌─────────────────┐
│   TUN   │ ────────────────────► │  PayloadOutTLS  │ ──► TLS服务器
│  设备   │                        │  PayloadOutDTLS │ ──► DTLS服务器
└─────────┘                        └─────────────────┘
    ▲
    │  payloadInToTun
    │
┌─────────┐
│PayloadIn│ ◄── tlsChannel / dtlsChannel ◄── 服务器
└─────────┘
```

## 协议类型

| 值 | 名称 | 说明 |
|----|------|------|
| `0x00` | DATA | IPv4/IPv6 数据包 |
| `0x03` | DPD-REQ | 死亡对等检测请求 |
| `0x04` | DPD-RESP | 死亡对等检测响应 |
| `0x05` | DISCONNECT | 断开连接 |
| `0x07` | KEEPALIVE | 保活 |
| `0x08` | COMPRESSED DATA | 压缩数据 |
| `0x09` | TERMINATE | 服务器关闭 |

## 协议参考

- [CSTP Channel Protocol](https://datatracker.ietf.org/doc/html/draft-mavrogiannopoulos-openconnect-03#section-2.1.4)
- [DTLS Channel Protocol](https://datatracker.ietf.org/doc/html/draft-mavrogiannopoulos-openconnect-03#section-2.1.5)
