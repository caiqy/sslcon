# 会话管理模块 (session)

## 概述

`session/` 模块管理 VPN 连接的会话状态，包括主会话、连接会话和 DTLS 会话。

## 文件结构

```
session/
└── session.go    # 会话管理实现
```

## 核心数据结构

### Session

主会话，管理整个 VPN 连接生命周期：

```go
type Session struct {
    SessionToken    string          // 服务器返回的会话令牌
    PreMasterSecret []byte          // DTLS 预主密钥

    ActiveClose bool                // 是否主动关闭
    CloseChan   chan struct{}       // 关闭通知通道
    CSess       *ConnSession        // 当前连接会话
}
```

### ConnSession

连接会话，存储具体连接的配置和状态：

```go
type ConnSession struct {
    Sess *Session `json:"-"`

    // 网络配置
    ServerAddress string   // 服务器 IP 地址
    LocalAddress  string   // 本地 IP 地址
    Hostname      string   // 服务器主机名
    TunName       string   // TUN 设备名称
    VPNAddress    string   // VPN 分配的 IPv4 地址
    VPNMask       string   // VPN 子网掩码
    DNS           []string // DNS 服务器列表
    MTU           int      // 最大传输单元

    // 路由配置
    SplitInclude []string // 走 VPN 的路由
    SplitExclude []string // 不走 VPN 的路由

    // 动态分流
    DynamicSplitTunneling       bool
    DynamicSplitIncludeDomains  []string
    DynamicSplitIncludeResolved sync.Map
    DynamicSplitExcludeDomains  []string
    DynamicSplitExcludeResolved sync.Map

    // TLS 配置
    TLSCipherSuite   string // TLS 加密套件
    TLSDpdTime       int    // TLS DPD 间隔（秒）
    TLSKeepaliveTime int    // TLS 保活间隔（秒）

    // DTLS 配置
    DTLSPort          string // DTLS 端口
    DTLSDpdTime       int    // DTLS DPD 间隔（秒）
    DTLSKeepaliveTime int    // DTLS 保活间隔（秒）
    DTLSId            string // DTLS 会话 ID
    DTLSCipherSuite   string // DTLS 加密套件

    // 流量统计
    Stat *stat

    // 并发控制
    closeOnce      sync.Once
    CloseChan      chan struct{}
    PayloadIn      chan *proto.Payload  // 入站数据
    PayloadOutTLS  chan *proto.Payload  // TLS 出站数据
    PayloadOutDTLS chan *proto.Payload  // DTLS 出站数据

    // DTLS 状态
    DtlsConnected *atomic.Bool
    DtlsSetupChan chan struct{}
    DSess         *DtlsSession

    // 读超时控制
    ResetTLSReadDead  *atomic.Bool
    ResetDTLSReadDead *atomic.Bool
}
```

### DtlsSession

DTLS 会话：

```go
type DtlsSession struct {
    closeOnce sync.Once
    CloseChan chan struct{}
}
```

### stat

流量统计：

```go
type stat struct {
    BytesSent     uint64 `json:"bytesSent"`     // 发送字节数
    BytesReceived uint64 `json:"bytesReceived"` // 接收字节数
}
```

## 全局变量

```go
var Sess = &Session{}  // 全局会话实例
```

## 核心方法

### NewConnSession()

从 HTTP 响应头创建连接会话：

```go
func (sess *Session) NewConnSession(header *http.Header) *ConnSession
```

**解析的响应头：**

| Header | 说明 |
|--------|------|
| `X-CSTP-Address` | VPN IP 地址 |
| `X-CSTP-Netmask` | 子网掩码 |
| `X-CSTP-MTU` | MTU 值 |
| `X-CSTP-DNS` | DNS 服务器 |
| `X-CSTP-Split-Include` | 包含路由 |
| `X-CSTP-Split-Exclude` | 排除路由 |
| `X-CSTP-DPD` | TLS DPD 时间 |
| `X-CSTP-Keepalive` | TLS 保活时间 |
| `X-DTLS-Session-ID` | DTLS 会话 ID |
| `X-DTLS-App-ID` | DTLS 应用 ID (ocserv) |
| `X-DTLS-Port` | DTLS 端口 |
| `X-DTLS-DPD` | DTLS DPD 时间 |
| `X-DTLS-Keepalive` | DTLS 保活时间 |
| `X-DTLS12-CipherSuite` | DTLS 加密套件 |
| `X-CSTP-Post-Auth-XML` | 动态分流配置 |

### DPDTimer()

启动 Dead Peer Detection 定时器：

```go
func (cSess *ConnSession) DPDTimer()
```

定期向 TLS 和 DTLS 通道发送 DPD 请求（Type=0x03），检测连接存活状态。

### ReadDeadTimer()

启动读超时重置定时器：

```go
func (cSess *ConnSession) ReadDeadTimer()
```

每 4 秒重置一次读超时标志，避免频繁的超时时间计算。

### Close()

关闭连接会话：

```go
func (cSess *ConnSession) Close()
```

使用 `sync.Once` 确保只关闭一次，关闭所有通道并清理资源。

### DtlsSession.Close()

关闭 DTLS 会话：

```go
func (dSess *DtlsSession) Close()
```

## 数据通道

```
                    ┌─────────────────────────────────┐
                    │         ConnSession             │
                    │                                 │
   服务器数据 ──────►│  PayloadIn (cap=64)            │──────► TUN 设备
                    │                                 │
   TUN 数据 ────────►│  PayloadOutTLS (cap=64)        │──────► TLS 服务器
                    │                                 │
   TUN 数据 ────────►│  PayloadOutDTLS (cap=64)       │──────► DTLS 服务器
                    │                                 │
   关闭信号 ◄────────│  CloseChan                     │
                    │                                 │
   DTLS就绪 ◄────────│  DtlsSetupChan                 │
                    └─────────────────────────────────┘
```

## 动态分流

支持基于域名的动态路由分流，通过解析 `X-CSTP-Post-Auth-XML` 响应头：

```xml
<config>
    <opaque>
        <custom-attr>
            <dynamic-split-include-domains>example.com,test.com</dynamic-split-include-domains>
            <dynamic-split-exclude-domains>bypass.com</dynamic-split-exclude-domains>
        </custom-attr>
    </opaque>
</config>
```

## 协议参考

- [OpenConnect Protocol - Tunnel Establishment](https://datatracker.ietf.org/doc/html/draft-mavrogiannopoulos-openconnect-03#section-2.1.3)
