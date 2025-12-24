# CLI 命令

## 概述

`sslcon` 是项目提供的命令行 VPN 客户端工具，基于 [Cobra](https://github.com/spf13/cobra) 框架实现。

## 入口文件

```
sslcon.go  -> cmd/root.go
```

## 命令结构

```
sslcon
├── connect      # 连接 VPN
├── disconnect   # 断开连接
└── status       # 查看状态
```

## 全局帮助

```bash
$ ./sslcon
A CLI application that supports the OpenConnect SSL VPN protocol.
For more information, please visit https://github.com/tlslink/sslcon

Usage:
  sslcon [flags]
  sslcon [command]

Available Commands:
  connect     Connect to the VPN server
  disconnect  Disconnect from the VPN server
  status      Get VPN connection information

Flags:
  -h, --help   help for sslcon

Use "sslcon [command] --help" for more information about a command.
```

---

## connect 命令

连接到 VPN 服务器。

### 语法

```bash
sslcon connect [flags]
```

### 参数

| 短参数 | 长参数 | 类型 | 必需 | 说明 |
|--------|--------|------|------|------|
| `-s` | `--server` | string | ✓ | VPN 服务器地址 |
| `-u` | `--username` | string | ✓ | 用户名 |
| `-p` | `--password` | string | | 密码（不指定则交互输入） |
| `-g` | `--group` | string | | 用户组 |
| `-k` | `--key` | string | | 密钥 |
| `-l` | `--log_level` | string | | 日志级别（默认 info） |
| `-d` | `--log_path` | string | | 日志目录（默认系统临时目录） |

### 示例

```bash
# 基本连接（交互输入密码）
./sslcon connect -s vpn.example.com -u myuser

# 指定密码和用户组
./sslcon connect -s vpn.example.com -u myuser -p mypassword -g default

# 带端口号
./sslcon connect -s vpn.example.com:8443 -u myuser

# 指定密钥
./sslcon connect -s vpn.example.com -u myuser -g default -k secretkey

# 调试模式
./sslcon connect -s vpn.example.com -u myuser -l Debug -d /tmp/vpnlog
```

### 工作流程

1. 通过 WebSocket 连接到本地 vpnagent 服务 (6210 端口)
2. 发送 `config` RPC 调用设置配置
3. 发送 `connect` RPC 调用建立 VPN 连接
4. 输出连接结果

---

## disconnect 命令

断开当前 VPN 连接。

### 语法

```bash
sslcon disconnect
```

### 示例

```bash
./sslcon disconnect
```

### 工作流程

1. 连接到本地 vpnagent 服务
2. 发送 `disconnect` RPC 调用
3. 输出断开结果

---

## status 命令

获取当前 VPN 连接状态。

### 语法

```bash
sslcon status
```

### 示例

```bash
./sslcon status
```

### 输出示例（已连接）

```json
{
  "ServerAddress": "192.168.1.1",
  "LocalAddress": "192.168.1.100",
  "Hostname": "vpn.example.com",
  "TunName": "SSLCon",
  "VPNAddress": "10.0.0.2",
  "VPNMask": "255.255.255.0",
  "DNS": ["8.8.8.8", "8.8.4.4"],
  "MTU": 1399,
  "TLSCipherSuite": "TLS_AES_256_GCM_SHA384",
  "DTLSCipherSuite": "ECDHE-RSA-AES256-GCM-SHA384"
}
```

---

## RPC 通信

CLI 命令通过 JSON-RPC 与 vpnagent 通信：

```go
func rpcCall(method string, params map[string]string, 
             result *gson.Value, id int) error
```

通信地址：`ws://127.0.0.1:6210/rpc`

---

## 前提条件

使用 CLI 前需确保 vpnagent 服务已运行：

```bash
# 安装并启动服务
sudo ./vpnagent install

# 或直接前台运行
sudo ./vpnagent
```
