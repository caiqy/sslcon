# SSLCon 项目 Wiki

欢迎来到 SSLCon 项目文档中心。本 Wiki 提供了项目的完整技术文档，包括架构设计、模块说明、API 文档和编码规范等。

## 📚 文档目录

### 项目概述
- [项目简介](./01-overview/project-intro.md) - 项目背景、目标和主要功能
- [架构设计](./01-overview/architecture.md) - 系统整体架构和设计理念
- [技术栈](./01-overview/tech-stack.md) - 使用的技术和依赖库

### 核心模块
- [认证模块 (auth)](./02-modules/auth.md) - VPN 认证流程和实现
- [会话管理 (session)](./02-modules/session.md) - 连接会话状态管理
- [VPN 隧道 (vpn)](./02-modules/vpn.md) - TLS/DTLS 隧道实现
- [TUN 设备 (tun)](./02-modules/tun.md) - 虚拟网卡操作
- [RPC 服务 (rpc)](./02-modules/rpc.md) - WebSocket JSON-RPC 接口
- [系统服务 (svc)](./02-modules/svc.md) - 后台服务管理
- [协议定义 (proto)](./02-modules/proto.md) - OpenConnect 协议数据结构
- [工具函数 (utils)](./02-modules/utils.md) - 通用工具和辅助函数
- [基础设施 (base)](./02-modules/base.md) - 配置和日志管理

### 命令行工具
- [CLI 命令](./03-cli/commands.md) - sslcon 命令行使用说明
- [vpnagent 服务](./03-cli/vpnagent.md) - VPN 代理服务说明

### API 文档
- [JSON-RPC API](./04-api/jsonrpc.md) - WebSocket RPC 接口规范

### 开发指南
- [编码规范](./05-development/coding-style.md) - 代码风格和最佳实践
- [构建部署](./05-development/build-deploy.md) - 编译和部署说明
- [平台适配](./05-development/platform.md) - 多平台支持说明

## 🔗 相关链接

- [GitHub 仓库](https://github.com/tlslink/sslcon)
- [OpenConnect 协议规范](https://datatracker.ietf.org/doc/html/draft-mavrogiannopoulos-openconnect-04)
- [GUI 客户端示例](https://github.com/tlslink/anylink-client)

## 📝 文档维护

本文档基于项目源码自动生成，如有疑问请参考源代码或提交 Issue。
