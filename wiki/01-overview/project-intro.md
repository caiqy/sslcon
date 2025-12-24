# 项目简介

## 概述

**SSLCon** 是一个使用 Go 语言实现的 [OpenConnect VPN 协议](https://datatracker.ietf.org/doc/html/draft-mavrogiannopoulos-openconnect-04) 客户端。项目专注于客户端开发，提供命令行工具和后台服务两种运行模式。

## 项目目标

1. **跨平台支持** - 支持 Linux、macOS 和 Windows 操作系统
2. **协议兼容** - 兼容 AnyLink 和 OpenConnect VPN Server (ocserv)
3. **API 驱动** - 通过 WebSocket 和 JSON-RPC 2.0 暴露 API，便于 GUI 开发
4. **权限分离** - 将需要管理员权限的 VPN 代理服务与前端 UI 分离

## 项目组件

### 1. sslcon (CLI 工具)
命令行 VPN 客户端，提供以下功能：
- `connect` - 连接 VPN 服务器
- `disconnect` - 断开 VPN 连接
- `status` - 查看连接状态

### 2. vpnagent (后台服务)
VPN 代理服务，以管理员权限运行的后台进程：
- 管理网络接口和路由
- 提供 WebSocket JSON-RPC API (端口 6210)
- 支持系统服务安装/卸载

## 支持的服务器

| 服务器 | 链接 |
|--------|------|
| AnyLink | https://github.com/bjdgyc/anylink |
| OpenConnect VPN Server | https://gitlab.com/openconnect/ocserv |

## 核心特性

### 双通道传输
- **TLS (CSTP)** - 基于 TCP 的加密隧道，作为主通道
- **DTLS** - 基于 UDP 的加密隧道，提供更低延迟

### 动态分流
支持基于域名的动态路由分流：
- `DynamicSplitIncludeDomains` - 指定域名走 VPN
- `DynamicSplitExcludeDomains` - 指定域名不走 VPN

### 断线重连
支持网络切换后的自动重连机制。

## 项目结构

```
sslcon/
├── auth/           # VPN 认证模块
├── base/           # 基础配置和日志
├── cmd/            # CLI 命令定义
├── proto/          # 协议数据结构
├── rpc/            # WebSocket RPC 服务
├── session/        # 会话状态管理
├── svc/            # 系统服务管理
├── tun/            # TUN 虚拟网卡
├── utils/          # 工具函数
├── vpn/            # VPN 隧道实现
├── sslcon.go       # CLI 入口
└── vpnagent.go     # 服务入口
```

## 使用示例

### 安装服务
```bash
sudo ./vpnagent install
```

### 连接 VPN
```bash
./sslcon connect -s vpn.example.com -u username -g group
```

### WebSocket API
```
ws://127.0.0.1:6210/rpc
```
