# TUN 设备模块 (tun)

## 概述

`tun/` 模块提供跨平台的 TUN 虚拟网卡操作，代码主要来源于 [wireguard-go](https://git.zx2c4.com/wireguard-go/tree/tun/tun.go)。

## 文件结构

```
tun/
├── rwcancel/        # 读写取消机制 (Linux/Darwin)
├── tun.go           # 设备接口定义
├── tun_darwin.go    # macOS 实现
├── tun_linux.go     # Linux 实现
└── tun_windows.go   # Windows 实现
```

## 核心接口

### Device

TUN 设备统一接口：

```go
type Device interface {
    File() *os.File                 // 返回设备文件描述符
    Read([]byte, int) (int, error)  // 读取数据包
    Write([]byte, int) (int, error) // 写入数据包
    Flush() error                   // 刷新写缓冲
    MTU() (int, error)              // 获取 MTU
    Name() (string, error)          // 获取设备名称
    Events() <-chan Event           // 事件通道
    Close() error                   // 关闭设备
}
```

### Event

设备事件类型：

```go
type Event int

const (
    EventUp = 1 << iota  // 设备启动
    EventDown            // 设备关闭
    EventMTUUpdate       // MTU 更新
)
```

## 全局变量

```go
var NativeTunDevice *NativeTun  // 全局 TUN 设备实例
```

## 平台实现

### Windows (tun_windows.go)

使用 [WinTun](https://www.wintun.net/) 驱动：

```go
func CreateTUN(ifname string, mtu int) (Device, error)
```

**特点：**
- 设备名固定为 `SSLCon`
- 使用 Ring Buffer 高效传输数据
- 通过 `winipcfg` 配置 IP 和路由

### Linux (tun_linux.go)

使用内核 TUN 驱动：

```go
func CreateTUN(name string, mtu int) (Device, error)
```

**特点：**
- 设备名为 `sslcon`
- 使用 `ioctl` 系统调用
- 支持多队列 (IFF_MULTI_QUEUE)

### macOS (tun_darwin.go)

使用 `utun` 系统接口：

```go
func CreateTUN(name string, mtu int) (Device, error)
```

**特点：**
- 设备名自动分配 `utunX`
- 使用 `SYSPROTO_CONTROL` socket
- 数据包需要 4 字节头偏移

## NativeTun 结构

### Windows

```go
type NativeTun struct {
    wt        *wintun.Adapter
    name      string
    handle    windows.Handle
    close     int32
    running   sync.WaitGroup
    forcedMTU int
    rate      rateJuggler
    session   wintun.Session
    readWait  windows.Handle
    events    chan Event
    closeOnce sync.Once
    tunLUID   winipcfg.LUID
}
```

### Linux

```go
type NativeTun struct {
    tunFile                 *os.File
    index                   int32
    errors                  chan error
    events                  chan Event
    netlinkSock             int
    netlinkCancel           *rwcancel.RWCancel
    hackListenerClosed      sync.Mutex
    statusListenersShutdown chan struct{}
    closeOnce               sync.Once
}
```

### macOS

```go
type NativeTun struct {
    name        string
    tunFile     *os.File
    events      chan Event
    errors      chan error
    routeSocket int
    closeOnce   sync.Once
}
```

## 数据包偏移

| 平台 | 偏移量 | 原因 |
|------|--------|------|
| Windows | 0 | 无额外头 |
| Linux | 0 | 无额外头 |
| macOS | 4 | utun 协议族头 |

## 使用示例

```go
// 创建 TUN 设备
dev, err := tun.CreateTUN("sslcon", 1399)
if err != nil {
    return err
}
defer dev.Close()

// 保存全局引用
tun.NativeTunDevice = dev.(*tun.NativeTun)

// 获取设备名称
name, _ := dev.Name()

// 读取数据
buf := make([]byte, 1500)
n, err := dev.Read(buf, offset)

// 写入数据
_, err = dev.Write(data, offset)
```

## 平台特定 API

### Windows

```go
func (tun *NativeTun) LUID() winipcfg.LUID
```

获取 Windows 网络接口 LUID，用于配置 IP 和路由。

### Linux

```go
func (tun *NativeTun) routineRouteListener(tunIfindex int)
```

监听路由变化事件。

## 协议参考

- [WireguardGo TUN](https://git.zx2c4.com/wireguard-go/tree/tun/)
- [WinTun Driver](https://www.wintun.net/)
