# 系统服务模块 (svc)

## 概述

`svc/` 模块使用 [kardianos/service](https://github.com/kardianos/service) 库实现跨平台系统服务管理。

## 文件结构

```
svc/
└── service.go   # 服务管理实现
```

## 服务配置

```go
serviceConfig = &service.Config{
    Name:        svcName,           // "sslcon" (Linux/macOS) 或 "SSLCon" (Windows)
    DisplayName: "SSLCon VPN Agent",
    Description: "SSLCon SSL VPN service Agent",
}
```

## 服务接口

### program 结构

实现 `service.Interface` 接口：

```go
type program struct{}

func (p program) Start(s service.Service) error  // 服务启动
func (p program) Stop(s service.Service) error   // 服务停止
func (p program) run()                           // 主运行逻辑
```

### Start()

服务启动回调，异步执行 `run()`：

```go
func (p program) Start(s service.Service) error {
    if service.Interactive() {
        logger.Info("Running in terminal.")
    } else {
        logger.Info("Running under service manager.")
    }
    go p.run()
    return nil
}
```

### Stop()

服务停止回调，断开 VPN 连接：

```go
func (p program) Stop(s service.Service) error {
    logger.Info("I'm Stopping!")
    base.Info("Stop")
    rpc.DisConnect()
    return nil
}
```

### run()

主运行逻辑，初始化基础设施和 RPC 服务：

```go
func (p program) run() {
    base.Setup()
    rpc.Setup()
}
```

## 公开函数

### RunSvc()

运行服务（前台或后台模式）：

```go
func RunSvc()
```

### InstallSvc()

安装并启动系统服务：

```go
func InstallSvc()
```

**流程：**
1. 创建服务实例
2. 安装服务
3. 自动启动服务

### UninstallSvc()

停止并卸载系统服务：

```go
func UninstallSvc()
```

**流程：**
1. 创建服务实例
2. 停止服务
3. 卸载服务

## 平台服务管理

### Linux (systemd)

```bash
# 安装服务
sudo ./vpnagent install

# 卸载服务
sudo ./vpnagent uninstall

# 服务管理
sudo systemctl start/stop/restart sslcon.service
sudo systemctl enable/disable sslcon.service
```

### Linux (OpenWrt)

```bash
# 服务管理
/etc/init.d/sslcon start/stop/restart/status
```

### Windows

```powershell
# 安装服务
.\vpnagent.exe install

# 卸载服务
.\vpnagent.exe uninstall

# 服务管理（使用 services.msc 或 sc 命令）
sc start SSLCon
sc stop SSLCon
```

### macOS

```bash
# 安装服务
sudo ./vpnagent install

# 卸载服务
sudo ./vpnagent uninstall

# 服务管理
sudo launchctl start/stop sslcon
```

## 运行模式检测

```go
if service.Interactive() {
    // 终端交互模式
} else {
    // 系统服务模式
}
```

## 日志记录

```go
var logger service.Logger

// 使用系统日志
logger, err = svc.Logger(errs)
logger.Info("message")
```

服务模式下日志输出到系统日志（Windows Event Log / syslog）。
