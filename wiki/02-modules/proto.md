# 协议定义模块 (proto)

## 概述

`proto/` 模块定义 OpenConnect VPN 协议的数据结构，包括数据包格式和 XML DTD 定义。

## 文件结构

```
proto/
├── dtd.go       # XML 数据类型定义
└── protocol.go  # 数据包协议定义
```

## 数据包协议 (protocol.go)

### 协议头格式

CSTP (TLS) 数据包使用 8 字节固定头：

```
+------+------+------+------+------+------+------+------+
| Byte |  0   |  1   |  2   |  3   |  4   |  5   |  6   |  7   |
+------+------+------+------+------+------+------+------+------+
| 值   | 0x53 | 0x54 | 0x46 | 0x01 | LenH | LenL | Type | 0x00 |
+------+------+------+------+------+------+------+------+------+
| 说明 |  'S' |  'T' |  'F' | 固定 |  长度（大端）  | 类型 | 固定 |
+------+------+------+------+------+------+------+------+------+
```

### Header 定义

```go
var Header = []byte{
    0x53, 0x54, 0x46, 0x01,  // 魔数 "STF\x01"
    0x00, 0x00,               // 数据长度（大端序）
    0x00,                     // 负载类型
    0x00,                     // 固定为 0
}
```

### Payload 结构

```go
type Payload struct {
    Type byte   // 负载类型
    Data []byte // 负载数据
}
```

### 负载类型

| 值 | 常量 | 说明 |
|----|------|------|
| `0x00` | DATA | IPv4/IPv6 数据包 |
| `0x03` | DPD-REQ | Dead Peer Detection 请求 |
| `0x04` | DPD-RESP | Dead Peer Detection 响应 |
| `0x05` | DISCONNECT | 断开连接请求 |
| `0x07` | KEEPALIVE | 保活包 |
| `0x08` | COMPRESSED DATA | 压缩数据包 |
| `0x09` | TERMINATE | 服务器关闭通知 |

### DTLS 数据包格式

DTLS (UDP) 使用简化的 1 字节头：

```
+------+--------+
| Type | Data   |
+------+--------+
```

## XML DTD 定义 (dtd.go)

### DTD 主结构

```go
type DTD struct {
    XMLName              xml.Name       `xml:"config-auth"`
    Client               string         `xml:"client,attr"`
    Type                 string         `xml:"type,attr"`
    AggregateAuthVersion string         `xml:"aggregate-auth-version,attr"`
    Version              string         `xml:"version"`
    GroupAccess          string         `xml:"group-access"`
    GroupSelect          string         `xml:"group-select"`
    SessionToken         string         `xml:"session-token"`
    Auth                 auth           `xml:"auth"`
    DeviceId             deviceId       `xml:"device-id"`
    Opaque               opaque         `xml:"opaque"`
    MacAddressList       macAddressList `xml:"mac-address-list"`
    Config               config         `xml:"config"`
}
```

### 请求类型 (Type)

| 类型 | 说明 |
|------|------|
| `init` | 初始化请求 |
| `auth-reply` | 认证回复 |
| `auth-request` | 认证请求（服务端） |
| `complete` | 认证完成 |
| `logout` | 登出 |

### auth 结构

```go
type auth struct {
    Username string    `xml:"username"`
    Password string    `xml:"password"`
    Message  string    `xml:"message"`
    Banner   string    `xml:"banner"`
    Error    authError `xml:"error"`
    Form     form      `xml:"form"`
}
```

### authError 结构

```go
type authError struct {
    Param1 string `xml:"param1,attr"`
    Value  string `xml:",chardata"`
}
```

### form 结构

```go
type form struct {
    Action string   `xml:"action,attr"`
    Groups []string `xml:"select>option"`
}
```

### deviceId 结构

```go
type deviceId struct {
    ComputerName    string `xml:"computer-name,attr"`
    DeviceType      string `xml:"device-type,attr"`
    PlatformVersion string `xml:"platform-version,attr"`
    UniqueId        string `xml:"unique-id,attr"`
    UniqueIdGlobal  string `xml:"unique-id-global,attr"`
}
```

### opaque 结构

```go
type opaque struct {
    TunnelGroup string `xml:"tunnel-group"`
    GroupAlias  string `xml:"group-alias"`
    ConfigHash  string `xml:"config-hash"`
}
```

### 动态分流配置

```go
type config struct {
    Opaque opaque2 `xml:"opaque"`
}

type opaque2 struct {
    CustomAttr customAttr `xml:"custom-attr"`
}

type customAttr struct {
    DynamicSplitExcludeDomains string `xml:"dynamic-split-exclude-domains"`
    DynamicSplitIncludeDomains string `xml:"dynamic-split-include-domains"`
}
```

## XML 示例

### 初始化请求

```xml
<?xml version="1.0" encoding="UTF-8"?>
<config-auth client="vpn" type="init" aggregate-auth-version="2">
    <version who="vpn">4.10.07062</version>
    <device-id computer-name="MyPC" 
               device-type="Windows" 
               platform-version="10.0.19041" 
               unique-id="xxx">
    </device-id>
</config-auth>
```

### 初始化响应

```xml
<?xml version="1.0" encoding="UTF-8"?>
<config-auth client="vpn" type="auth-request">
    <auth>
        <form action="/auth">
            <select>
                <option>default</option>
                <option>admin</option>
            </select>
        </form>
    </auth>
    <opaque>
        <tunnel-group>default</tunnel-group>
        <group-alias>default</group-alias>
        <config-hash>abc123</config-hash>
    </opaque>
</config-auth>
```

### 认证完成响应

```xml
<?xml version="1.0" encoding="UTF-8"?>
<config-auth client="vpn" type="complete">
    <session-token>TOKEN_STRING</session-token>
    <config>
        <opaque>
            <custom-attr>
                <dynamic-split-include-domains>example.com,test.com</dynamic-split-include-domains>
            </custom-attr>
        </opaque>
    </config>
</config-auth>
```

## 协议参考

- [OpenConnect Protocol DTD](https://datatracker.ietf.org/doc/html/draft-mavrogiannopoulos-openconnect-03#appendix-C.1)
- [CSTP Channel Protocol](https://datatracker.ietf.org/doc/html/draft-mavrogiannopoulos-openconnect-04#name-the-cstp-channel-protocol)
