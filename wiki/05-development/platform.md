# 平台适配

## 概述

SSLCon 支持 Linux、macOS 和 Windows 三大平台，通过构建标签和平台特定实现文件实现兼容。

## 支持平台

| 平台 | 架构 | 状态 |
|------|------|------|
| Linux | amd64, arm64, mips | ✓ |
| macOS | amd64, arm64 | ✓ |
| Windows | amd64 | ✓ |

---

## 平台差异

### TUN 设备

| 平台 | 实现 | 设备名 |
|------|------|--------|
| Windows | WinTun 驱动 | `SSLCon` |
| Linux | 内核 TUN 驱动 | `sslcon` |
| macOS | utun 接口 | `utunX` |

### 数据包偏移

| 平台 | 偏移量 | 原因 |
|------|--------|------|
| Windows | 0 | 无额外头 |
| Linux | 0 | 无额外头 |
| macOS | 4 | utun 协议族头 |

### 网络配置

| 平台 | 方法 |
|------|------|
| Windows | `winipcfg` API |
| Linux | `netlink` API |
| macOS | 系统命令 (`ifconfig`, `route`) |

---

## 平台特定文件

### TUN 模块

```
tun/
├── tun.go           # 公共接口
├── tun_windows.go   # Windows WinTun
├── tun_linux.go     # Linux TUN
└── tun_darwin.go    # macOS utun
```

### 网络配置

```
utils/vpnc/
├── vpnc_windows.go  # Windows 网络配置
├── vpnc_linux.go    # Linux 网络配置
└── vpnc_darwin.go   # macOS 网络配置
```

---

## Windows 适配

### WinTun 驱动

项目使用 [WinTun](https://www.wintun.net/) 作为 Windows TUN 驱动：

```go
import "github.com/lysShub/wintun-go"
```

WinTun 驱动会在首次使用时自动安装。

### 网络配置

使用 `winipcfg` 配置网络：

```go
import "golang.zx2c4.com/wireguard/windows/tunnel/winipcfg"

// 获取接口 LUID
luid := tun.NativeTunDevice.LUID()

// 设置 IP 地址
luid.SetIPAddressesForFamily(windows.AF_INET, []netip.Prefix{prefix})

// 添加路由
luid.AddRoute(dst, nextHop, metric)

// 设置 DNS
luid.SetDNS(windows.AF_INET, servers, []string{})
```

### MTU 设置

通过 netsh 命令：

```go
cmdStr := fmt.Sprintf("netsh interface ipv4 set subinterface \"%s\" MTU=%d", ifname, mtu)
```

### 服务名称

Windows 服务名使用 `SSLCon`（首字母大写）。

---

## Linux 适配

### TUN 设备

使用内核 TUN 驱动：

```go
// 打开 /dev/net/tun
tunFile, err := os.OpenFile("/dev/net/tun", os.O_RDWR|syscall.O_CLOEXEC, 0)

// ioctl 配置
syscall.SYS_IOCTL, TUNSETIFF, IFF_TUN|IFF_NO_PI
```

### 网络配置

使用 `netlink` 库：

```go
import "github.com/vishvananda/netlink"

// 获取接口
link, _ := netlink.LinkByName(ifname)

// 设置 IP 地址
addr, _ := netlink.ParseAddr(ipCIDR)
netlink.AddrAdd(link, addr)

// 启动接口
netlink.LinkSetUp(link)

// 添加路由
route := &netlink.Route{
    LinkIndex: link.Attrs().Index,
    Dst:       dst,
    Gw:        gateway,
}
netlink.RouteAdd(route)
```

### DNS 配置

修改 `/etc/resolv.conf`：

```go
content := fmt.Sprintf("nameserver %s\n", dns)
os.WriteFile("/etc/resolv.conf", []byte(content), 0644)
```

### 服务管理

支持 systemd 和 OpenWrt init.d。

---

## macOS 适配

### utun 接口

使用系统 utun 接口：

```go
// 创建 socket
fd, _ := syscall.Socket(syscall.AF_SYSTEM, syscall.SOCK_DGRAM, 2) // SYSPROTO_CONTROL

// 连接到 utun
var ctlInfo = &struct {
    ctl_id   uint32
    ctl_name [96]byte
}{}
copy(ctlInfo.ctl_name[:], "com.apple.net.utun_control")
syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), CTLIOCGINFO, uintptr(unsafe.Pointer(ctlInfo)))
```

### 数据包偏移

macOS utun 需要 4 字节头（协议族）：

```go
offset = 4  // macOS 特有

// 读取时跳过头部
n, err = dev.Read(buf, offset)

// 写入时添加头部
expand := make([]byte, offset+len(data))
copy(expand[offset:], data)
dev.Write(expand, offset)
```

### 网络配置

使用系统命令：

```bash
# 配置 IP
ifconfig utun0 10.0.0.2 10.0.0.1 netmask 255.255.255.0

# 添加路由
route add -net 0.0.0.0/0 10.0.0.1

# 设置 DNS
networksetup -setdnsservers "Wi-Fi" 8.8.8.8
```

### 设备名称

utun 设备名由系统自动分配（utun0, utun1, ...），需要连接后查询实际名称：

```go
name, _ := dev.Name()
cSess.TunName = name  // 例如 "utun3"
```

---

## 构建标签

所有主要源文件使用构建标签限制平台：

```go
//go:build linux || darwin || windows
```

---

## 测试建议

### 单元测试

工具函数可以跨平台测试：

```bash
go test ./utils/...
```

### 集成测试

需要在目标平台上进行：

1. 创建 TUN 设备
2. 配置网络
3. 建立 VPN 连接
4. 验证路由

### CI/CD

建议使用 GitHub Actions 进行多平台构建测试。
