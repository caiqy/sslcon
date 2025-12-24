# 工具函数模块 (utils)

## 概述

`utils/` 模块提供通用工具函数和平台特定的网络配置功能。

## 文件结构

```
utils/
├── record.go           # 数据记录工具
├── utils.go            # 通用工具函数
├── vpnc/               # VPN 网络配置
│   ├── vpnc_darwin.go  # macOS 实现
│   ├── vpnc_linux.go   # Linux 实现
│   └── vpnc_windows.go # Windows 实现
└── waterutil/          # IP 数据包解析
    ├── ip_protocols.go # IP 协议常量
    └── utils_ipv4.go   # IPv4 工具
```

## 通用工具 (utils.go)

### InArray()

检查字符串是否在数组中：

```go
func InArray(arr []string, str string) bool
```

### InArrayGeneric()

检查字符串是否以数组中任一元素结尾（域名匹配）：

```go
func InArrayGeneric(arr []string, str string) bool
```

### SetCommonHeader()

设置 HTTP 请求公共头：

```go
func SetCommonHeader(req *http.Request)
```

设置 `User-Agent` 和 `Content-Type` 头。

### IpMask2CIDR()

将 IP 和掩码转换为 CIDR 格式：

```go
func IpMask2CIDR(ip, mask string) string
// 例: IpMask2CIDR("192.168.1.0", "255.255.255.0") => "192.168.1.0/24"
```

### IpMaskToCIDR()

将 IP/掩码字符串转换为 CIDR：

```go
func IpMaskToCIDR(ipMask string) string
// 例: IpMaskToCIDR("192.168.1.10/255.255.255.255") => "192.168.1.10/32"
```

### ResolvePacket()

解析 IP 数据包获取源/目标地址和端口：

```go
func ResolvePacket(packet []byte) (src string, srcPort uint16, dst string, dstPort uint16)
```

### MakeMasterSecret()

生成 DTLS 预主密钥：

```go
func MakeMasterSecret() ([]byte, error)
```

生成 48 字节随机密钥，前两字节为 DTLS 版本号。

### Min() / Max()

整数最小/最大值：

```go
func Min(init int, other ...int) int
func Max(init int, other ...int) int
```

### CopyFile()

复制文件：

```go
func CopyFile(dstName, srcName string) error
```

### FirstUpper()

首字母大写：

```go
func FirstUpper(s string) string
```

### RemoveBetween()

移除两个标记之间的内容：

```go
func RemoveBetween(input, start, end string) string
```

---

## VPN 网络配置 (vpnc/)

平台特定的网络接口配置和路由管理。

### 公共接口

```go
func ConfigInterface(cSess *session.ConnSession) error  // 配置网络接口
func SetRoutes(cSess *session.ConnSession) error        // 设置路由
func ResetRoutes(cSess *session.ConnSession)            // 重置路由
func GetLocalInterface() error                          // 获取本地接口信息
func DynamicAddIncludeRoutes(ips []string)              // 动态添加包含路由
func DynamicAddExcludeRoutes(ips []string)              // 动态添加排除路由
```

### Windows 实现

使用 `winipcfg` 库配置网络：

```go
// 配置 IP 地址
iface.SetIPAddressesForFamily(windows.AF_INET, []netip.Prefix{prefixVPN})

// 添加路由
iface.AddRoute(dst, nextHop, metric)

// 设置 DNS
iface.SetDNS(windows.AF_INET, servers, []string{})

// 设置 MTU
netsh interface ipv4 set subinterface "SSLCon" MTU=1399
```

### Linux 实现

使用 `netlink` 库配置网络：

```go
// 配置 IP 地址
netlink.AddrAdd(link, addr)

// 添加路由
netlink.RouteAdd(route)

// 设置 DNS (修改 /etc/resolv.conf)
```

### macOS 实现

使用系统命令配置网络：

```go
// 配置 IP 地址
ifconfig utun0 10.0.0.2 10.0.0.1 netmask 255.255.255.0

// 添加路由
route add -net 0.0.0.0/0 10.0.0.1

// 设置 DNS
networksetup -setdnsservers "Wi-Fi" 8.8.8.8
```

---

## IP 数据包工具 (waterutil/)

### ip_protocols.go

IP 协议号常量定义：

```go
const (
    ICMP   Protocol = 1
    TCP    Protocol = 6
    UDP    Protocol = 17
    ICMPv6 Protocol = 58
)
```

### utils_ipv4.go

IPv4 数据包解析函数：

```go
func IPv4Source(packet []byte) net.IP           // 源 IP
func IPv4Destination(packet []byte) net.IP       // 目标 IP
func IPv4SourcePort(packet []byte) uint16        // 源端口
func IPv4DestinationPort(packet []byte) uint16   // 目标端口
func IPv4Protocol(packet []byte) Protocol        // 协议类型
```

## 使用示例

### 解析数据包

```go
packet := []byte{...} // IP 数据包
src, srcPort, dst, dstPort := utils.ResolvePacket(packet)
fmt.Printf("From %s:%d to %s:%d\n", src, srcPort, dst, dstPort)
```

### 配置路由

```go
// 设置 VPN 路由
err := vpnc.SetRoutes(cSess)

// 动态添加排除路由
vpnc.DynamicAddExcludeRoutes([]string{"8.8.8.8", "8.8.4.4"})
```
