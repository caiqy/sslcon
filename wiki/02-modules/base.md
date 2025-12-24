# 基础设施模块 (base)

## 概述

`base/` 模块提供配置管理、日志记录和基础初始化功能。

## 文件结构

```
base/
├── config.go   # 配置定义
├── log.go      # 日志系统
└── setup.go    # 初始化设置
```

## 配置管理 (config.go)

### ClientConfig

客户端配置结构：

```go
type ClientConfig struct {
    LogLevel           string `json:"log_level"`      // 日志级别
    LogPath            string `json:"log_path"`       // 日志目录
    InsecureSkipVerify bool   `json:"skip_verify"`    // 跳过证书验证
    CiscoCompat        bool   `json:"cisco_compat"`   // Cisco 兼容模式
    NoDTLS             bool   `json:"no_dtls"`        // 禁用 DTLS
    AgentName          string `json:"agent_name"`     // 客户端名称
    AgentVersion       string `json:"agent_version"`  // 客户端版本
}
```

### Interface

本地网络接口信息：

```go
type Interface struct {
    Name    string `json:"name"`    // 接口名称
    Ip4     string `json:"ip4"`     // IPv4 地址
    Mac     string `json:"mac"`     // MAC 地址
    Gateway string `json:"gateway"` // 默认网关
}
```

### 全局变量

```go
var Cfg = &ClientConfig{}           // 全局配置
var LocalInterface = &Interface{}   // 本地接口信息
```

### 默认配置

```go
func initCfg() {
    Cfg.LogLevel = "Debug"
    Cfg.InsecureSkipVerify = true
    Cfg.CiscoCompat = true
    Cfg.AgentName = ""
    Cfg.AgentVersion = "4.10.07062"
}
```

---

## 日志系统 (log.go)

### 日志级别

```go
const (
    _Debug = iota  // 0 - 调试信息
    _Info          // 1 - 一般信息
    _Warn          // 2 - 警告
    _Error         // 3 - 错误
    _Fatal         // 4 - 致命错误
)
```

级别说明：只有 **≥ 设置级别** 的日志才会输出。

### logWriter

自定义日志写入器：

```go
type logWriter struct {
    UseStdout bool      // 是否输出到 stdout
    FileName  string    // 日志文件名
    File      *os.File  // 文件句柄
    NowDate   string    // 当前日期
}
```

### 日志函数

```go
func Debug(v ...interface{})  // 调试日志
func Info(v ...interface{})   // 信息日志
func Warn(v ...interface{})   // 警告日志
func Error(v ...interface{})  // 错误日志
func Fatal(v ...interface{})  // 致命错误（会退出程序）
```

### 日志格式

```
2024/01/15 10:30:45 auth.go:86: [Info] message content
```

包含：时间戳、源文件:行号、日志级别、消息内容

### InitLog()

初始化日志系统：

```go
func InitLog()
```

- 如果 `LogPath` 为空，输出到 stdout
- 否则创建日志文件 `{LogPath}/vpnagent.log`
- 每次重启客户端会清空日志文件

### GetBaseLogger()

获取底层 logger 实例（用于 RPC 库）：

```go
func GetBaseLogger() *log.Logger
```

---

## 初始化 (setup.go)

### Setup()

基础设施初始化：

```go
func Setup()
```

调用 `initCfg()` 和 `InitLog()` 进行初始化。

---

## 配置参数说明

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `log_level` | string | `Debug` | 日志级别：Debug/Info/Warn/Error/Fatal |
| `log_path` | string | `""` | 日志目录，空则输出到控制台 |
| `skip_verify` | bool | `true` | 跳过 TLS 证书验证 |
| `cisco_compat` | bool | `true` | 使用 AnyConnect 作为 User-Agent |
| `no_dtls` | bool | `false` | 禁用 DTLS 通道 |
| `agent_name` | string | `""` | 自定义客户端名称 |
| `agent_version` | string | `4.10.07062` | 客户端版本号 |

## 使用示例

### 配置日志

```go
base.Cfg.LogLevel = "Info"
base.Cfg.LogPath = "/var/log/vpn"
base.InitLog()
```

### 记录日志

```go
base.Debug("详细调试信息:", someVar)
base.Info("连接成功")
base.Warn("证书即将过期")
base.Error("连接失败:", err)
base.Fatal("无法启动服务")  // 会调用 os.Exit(1)
```

### 获取本地接口

```go
fmt.Println("接口:", base.LocalInterface.Name)
fmt.Println("IP:", base.LocalInterface.Ip4)
fmt.Println("MAC:", base.LocalInterface.Mac)
fmt.Println("网关:", base.LocalInterface.Gateway)
```
