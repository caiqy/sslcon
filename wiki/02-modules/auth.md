# 认证模块 (auth)

## 概述

`auth/` 模块负责 VPN 连接的认证流程，包括初始化认证和密码认证两个阶段。

## 文件结构

```
auth/
└── auth.go    # 认证逻辑实现
```

## 核心数据结构

### Profile

用户配置和设备信息，用于认证请求：

```go
type Profile struct {
    Host      string `json:"host"`      // VPN 服务器地址
    Username  string `json:"username"`  // 用户名
    Password  string `json:"password"`  // 密码
    Group     string `json:"group"`     // 用户组
    SecretKey string `json:"secret"`    // 密钥

    Initialized bool   // 是否已初始化
    AppVersion  string // 客户端版本号

    HostWithPort string // 带端口的主机地址
    Scheme       string // URL 协议 (https://)
    AuthPath     string // 认证路径

    MacAddress  string // 网卡 MAC 地址
    TunnelGroup string // 隧道组
    GroupAlias  string // 组别名
    ConfigHash  string // 配置哈希

    ComputerName    string // 计算机名
    DeviceType      string // 设备类型
    PlatformVersion string // 平台版本
    UniqueId        string // 设备唯一标识
}
```

## 全局变量

| 变量 | 类型 | 说明 |
|------|------|------|
| `Prof` | `*Profile` | 全局配置实例 |
| `Conn` | `*tls.Conn` | TLS 连接 |
| `BufR` | `*bufio.Reader` | 缓冲读取器 |
| `WebVpnCookie` | `string` | OpenConnect 兼容 Cookie |

## 核心函数

### InitAuth()

初始化认证，建立 TLS 连接并获取认证路径和用户组信息。

```go
func InitAuth() error
```

**流程：**
1. 建立 TLS 连接到 VPN 服务器
2. 发送 `init` 类型的 XML 请求
3. 解析响应获取 `AuthPath`、`TunnelGroup` 等信息
4. 验证用户组是否在允许列表中

**返回错误：**
- TLS 连接失败
- 用户组不在允许列表

### PasswordAuth()

密码认证，发送用户凭证获取 SessionToken。

```go
func PasswordAuth() error
```

**流程：**
1. 发送 `auth-reply` 类型的 XML 请求（包含用户名和密码）
2. 如果需要两步认证，再次发送请求
3. 解析响应获取 `SessionToken`
4. 兼容 OpenConnect 服务器的 `webvpn` Cookie

**返回错误：**
- 用户名/密码错误
- 认证消息错误

### tplPost()

内部函数，渲染 XML 模板并发送 POST 请求。

```go
func tplPost(typ int, path string, dtd *proto.DTD) error
```

## XML 模板

### 初始化请求 (templateInit)

```xml
<?xml version="1.0" encoding="UTF-8"?>
<config-auth client="vpn" type="init" aggregate-auth-version="2">
    <version who="vpn">{{.AppVersion}}</version>
    <device-id computer-name="{{.ComputerName}}" 
               device-type="{{.DeviceType}}" 
               platform-version="{{.PlatformVersion}}" 
               unique-id="{{.UniqueId}}">
    </device-id>
</config-auth>
```

### 认证回复 (templateAuthReply)

```xml
<?xml version="1.0" encoding="UTF-8"?>
<config-auth client="vpn" type="auth-reply" aggregate-auth-version="2">
    <version who="vpn">{{.AppVersion}}</version>
    <device-id ...></device-id>
    <opaque is-for="sg">
        <tunnel-group>{{.TunnelGroup}}</tunnel-group>
        <group-alias>{{.GroupAlias}}</group-alias>
        <config-hash>{{.ConfigHash}}</config-hash>
    </opaque>
    <mac-address-list>
        <mac-address public-interface="true">{{.MacAddress}}</mac-address>
    </mac-address-list>
    <auth>
        <username>{{.Username}}</username>
        <password>{{.Password}}</password>
    </auth>
    <group-select>{{.Group}}</group-select>
</config-auth>
```

## 请求头

认证请求包含以下自定义头：

| 头字段 | 值 | 说明 |
|--------|-----|------|
| `X-Transcend-Version` | `1` | 协议版本 |
| `X-Aggregate-Auth` | `1` | 聚合认证标识 |

## 初始化流程

模块初始化时自动获取系统信息：

```go
func init() {
    reqHeaders["X-Transcend-Version"] = "1"
    reqHeaders["X-Aggregate-Auth"] = "1"
    Prof.Scheme = "https://"
    
    // 获取系统信息
    host, _ := sysinfo.Host()
    info := host.Info()
    Prof.ComputerName = info.Hostname
    Prof.UniqueId = info.UniqueID
    // ...
}
```

## 协议参考

- [OpenConnect Protocol - Authentication](https://datatracker.ietf.org/doc/html/draft-mavrogiannopoulos-openconnect-03#section-2.1.2)
