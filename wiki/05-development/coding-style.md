# 编码规范

## 概述

本文档总结了 SSLCon 项目的代码风格和最佳实践，供开发者参考。

## 项目结构规范

### 目录组织

```
sslcon/
├── auth/           # 功能模块：认证
├── base/           # 基础设施：配置、日志
├── cmd/            # CLI 命令
├── proto/          # 协议定义
├── rpc/            # RPC 服务
├── session/        # 会话管理
├── svc/            # 系统服务
├── tun/            # TUN 设备
├── utils/          # 工具函数
├── vpn/            # VPN 隧道
├── sslcon.go       # CLI 入口
└── vpnagent.go     # 服务入口
```

### 文件命名

- 使用小写字母和下划线
- 平台特定代码使用 `_<os>.go` 后缀
  - `tun_windows.go`
  - `tun_linux.go`
  - `tun_darwin.go`

---

## Go 代码风格

### 包命名

- 使用简短、小写的包名
- 避免使用下划线或混合大小写

```go
package auth      // ✓
package vpn       // ✓
package vpnAuth   // ✗
package vpn_auth  // ✗
```

### 导入分组

按标准库、第三方库、本地包顺序分组：

```go
import (
    // 标准库
    "bufio"
    "bytes"
    "crypto/tls"

    // 第三方库
    "github.com/elastic/go-sysinfo"
    "github.com/gorilla/websocket"

    // 本地包
    "sslcon/base"
    "sslcon/proto"
    "sslcon/session"
)
```

### 变量命名

**驼峰命名法：**
```go
var SessionToken string      // 导出变量，首字母大写
var reqHeaders map[string]string  // 私有变量，首字母小写
```

**常量命名：**
```go
const (
    _Debug = iota  // 私有常量
    _Info
    _Warn
)

const (
    STATUS = iota  // 导出常量
    CONFIG
    CONNECT
)
```

### 结构体定义

使用 JSON 标签进行序列化：

```go
type ClientConfig struct {
    LogLevel           string `json:"log_level"`
    LogPath            string `json:"log_path"`
    InsecureSkipVerify bool   `json:"skip_verify"`
}
```

---

## 错误处理

### 错误返回

函数返回错误作为最后一个返回值：

```go
func InitAuth() error {
    // ...
    if err != nil {
        return err
    }
    return nil
}
```

### 错误包装

使用 `fmt.Errorf` 添加上下文：

```go
return fmt.Errorf("auth error %s", resp.Status)
return fmt.Errorf("routing error: %s %s", dst.String(), err)
```

### 错误日志

```go
if err != nil {
    base.Error("tls server to payloadIn error:", err)
    return
}
```

---

## 并发编程

### Channel 使用

定义带缓冲的 channel：

```go
PayloadIn:      make(chan *proto.Payload, 64),
PayloadOutTLS:  make(chan *proto.Payload, 64),
CloseChan:      make(chan struct{}),
```

### 协程安全关闭

使用 `sync.Once` 确保只关闭一次：

```go
func (cSess *ConnSession) Close() {
    cSess.closeOnce.Do(func() {
        close(cSess.CloseChan)
        // cleanup...
    })
}
```

### select 模式

```go
for {
    select {
    case pl = <-cSess.PayloadOutTLS:
        // 处理数据
    case <-cSess.CloseChan:
        return
    }
}
```

### 原子操作

使用 `atomic.Bool` 进行并发安全的状态检查：

```go
DtlsConnected: atomic.NewBool(false)

if cSess.DtlsConnected.Load() {
    // DTLS 已连接
}

cSess.DtlsConnected.Store(true)
```

---

## 资源管理

### defer 使用

确保资源正确释放：

```go
func tlsChannel(conn *tls.Conn, ...) {
    defer func() {
        base.Info("tls channel exit")
        resp.Body.Close()
        _ = conn.Close()
        cSess.Close()
    }()
    // ...
}
```

### 对象池

使用 `sync.Pool` 复用对象：

```go
var payloadBufferPool = sync.Pool{
    New: func() interface{} {
        return &proto.Payload{
            Data: make([]byte, BufferSize),
        }
    },
}

func getPayloadBuffer() *proto.Payload {
    return payloadBufferPool.Get().(*proto.Payload)
}

func putPayloadBuffer(pl *proto.Payload) {
    payloadBufferPool.Put(pl)
}
```

---

## 平台兼容

### 构建标签

```go
//go:build linux || darwin || windows
```

### 平台特定实现

每个平台一个文件，实现相同的接口：

```go
// vpnc_windows.go
func ConfigInterface(cSess *session.ConnSession) error { ... }

// vpnc_linux.go
func ConfigInterface(cSess *session.ConnSession) error { ... }

// vpnc_darwin.go
func ConfigInterface(cSess *session.ConnSession) error { ... }
```

---

## 注释规范

### 包注释

```go
// Package tun copy from https://git.zx2c4.com/wireguard-go/tree/tun/tun.go
package tun
```

### 函数注释

```go
// InitAuth 确定用户组和服务端认证地址 AuthPath
func InitAuth() error {
```

### 行内注释

解释复杂逻辑或协议细节：

```go
// https://datatracker.ietf.org/doc/html/draft-mavrogiannopoulos-openconnect-03#section-2.1.3
reqHeaders["Cookie"] = "webvpn=" + session.Sess.SessionToken
```

---

## 日志规范

### 日志级别使用

| 级别 | 用途 |
|------|------|
| `Debug` | 详细调试信息 |
| `Info` | 正常运行信息 |
| `Warn` | 警告信息 |
| `Error` | 错误信息 |
| `Fatal` | 致命错误，程序退出 |

### 日志示例

```go
base.Debug("tun device:", cSess.TunName)
base.Info("tls channel negotiation succeeded")
base.Warn("证书即将过期")
base.Error("tls payloadOut to server error:", err)
base.Fatal(http.ListenAndServe(":6210", nil))
```

---

## 测试规范

- 测试文件以 `_test.go` 结尾
- 测试函数以 `Test` 开头
- 使用表驱动测试

```go
func TestIpMask2CIDR(t *testing.T) {
    tests := []struct {
        ip, mask, want string
    }{
        {"192.168.1.0", "255.255.255.0", "192.168.1.0/24"},
        {"10.0.0.0", "255.0.0.0", "10.0.0.0/8"},
    }
    for _, tt := range tests {
        got := IpMask2CIDR(tt.ip, tt.mask)
        if got != tt.want {
            t.Errorf("IpMask2CIDR(%s, %s) = %s; want %s",
                tt.ip, tt.mask, got, tt.want)
        }
    }
}
```
