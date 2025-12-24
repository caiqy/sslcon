# vpnagent 服务

## 概述

`vpnagent` 是 VPN 代理服务，需要以管理员/root 权限运行。它提供网络配置、隧道管理和 WebSocket RPC 服务。

## 入口文件

```
vpnagent.go -> svc/service.go
             -> rpc/rpc.go
             -> base/setup.go
```

## 运行模式

### 1. 交互模式

在终端前台运行：

```bash
# Linux/macOS
sudo ./vpnagent

# Windows (管理员权限)
.\vpnagent.exe
```

输出示例：
```
2024/01/15 10:30:45 setup.go:15: [Info] Server pid: 12345
```

按 `Ctrl+C` 退出。

### 2. 服务模式

作为系统服务在后台运行。

---

## 服务管理

### 安装服务

```bash
# Linux/macOS
sudo ./vpnagent install

# Windows (管理员权限)
.\vpnagent.exe install
```

安装后服务会自动启动。

### 卸载服务

```bash
# Linux/macOS
sudo ./vpnagent uninstall

# Windows (管理员权限)
.\vpnagent.exe uninstall
```

---

## 平台服务命令

### Linux (systemd)

```bash
# 查看状态
sudo systemctl status sslcon.service

# 启动服务
sudo systemctl start sslcon.service

# 停止服务
sudo systemctl stop sslcon.service

# 重启服务
sudo systemctl restart sslcon.service

# 开机自启
sudo systemctl enable sslcon.service

# 禁止自启
sudo systemctl disable sslcon.service

# 查看日志
journalctl -u sslcon.service -f
```

### Linux (OpenWrt)

```bash
# 启动
/etc/init.d/sslcon start

# 停止
/etc/init.d/sslcon stop

# 重启
/etc/init.d/sslcon restart

# 状态
/etc/init.d/sslcon status
```

### Windows

```powershell
# 使用 sc 命令
sc query SSLCon
sc start SSLCon
sc stop SSLCon

# 或使用 services.msc 图形界面
```

### macOS

```bash
# 使用 launchctl
sudo launchctl list | grep sslcon
sudo launchctl start sslcon
sudo launchctl stop sslcon
```

---

## 服务端口

| 端口 | 协议 | 用途 |
|------|------|------|
| 6210 | WebSocket | JSON-RPC API |

---

## 信号处理

交互模式下支持以下信号：

| 信号 | 行为 |
|------|------|
| `SIGINT` (Ctrl+C) | 优雅退出 |
| `SIGTERM` | 优雅退出 |
| `SIGQUIT` | 优雅退出 |

退出时会：
1. 断开 VPN 连接
2. 重置路由规则
3. 关闭 TUN 设备

---

## 启动流程

```
main()
  │
  ├─► 交互模式
  │     │
  │     ├── base.Setup()      # 初始化配置和日志
  │     ├── rpc.Setup()       # 启动 RPC 服务
  │     └── watchSignal()     # 监听退出信号
  │
  └─► 服务模式
        │
        └── svc.RunSvc()      # 运行系统服务
              │
              ├── program.Start()
              │     └── program.run()
              │           ├── base.Setup()
              │           └── rpc.Setup()
              │
              └── program.Stop()
                    └── rpc.DisConnect()
```

---

## 权限要求

vpnagent 需要管理员权限来：

1. **创建 TUN 设备** - 虚拟网卡操作
2. **配置 IP 地址** - 设置 VPN IP
3. **修改路由表** - 设置 VPN 路由
4. **配置 DNS** - 设置 DNS 服务器

---

## 故障排查

### 服务无法启动

```bash
# 检查端口占用
netstat -tlnp | grep 6210  # Linux
netstat -an | findstr 6210 # Windows

# 检查服务状态
sudo systemctl status sslcon.service
```

### 连接失败

```bash
# 检查 vpnagent 是否运行
ps aux | grep vpnagent  # Linux/macOS
tasklist | findstr vpnagent  # Windows

# 查看日志
cat /var/log/vpn/vpnagent.log
```

### 权限问题

确保以 root/管理员权限运行：

```bash
# Linux/macOS
sudo ./vpnagent

# Windows - 右键以管理员身份运行
```
