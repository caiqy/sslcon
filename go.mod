module sslcon

go 1.20

require (
	github.com/elastic/go-sysinfo v1.11.2
	github.com/gopacket/gopacket v1.2.0
	github.com/gorilla/websocket v1.5.3
	github.com/jackpal/gateway v1.0.10
	github.com/kardianos/service v1.2.2
	github.com/pion/dtls/v2 v2.2.8
	github.com/sourcegraph/jsonrpc2 v0.2.1
	github.com/spf13/cobra v1.8.0
	github.com/vishvananda/netlink v1.2.1-beta.2
	go.uber.org/atomic v1.11.0
	golang.org/x/crypto v0.17.0
	golang.org/x/net v0.25.0
	golang.org/x/sys v0.18.0
	golang.zx2c4.com/wintun v0.0.0-20230126152724-0fa3db229ce2
	golang.zx2c4.com/wireguard/windows v0.5.3
)

require (
	github.com/elastic/go-windows v1.0.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/joeshaw/multierror v0.0.0-20140124173710-69b34d4ec901 // indirect
	github.com/pion/logging v0.2.4 // indirect
	github.com/pion/transport/v2 v2.2.4 // indirect
	github.com/prometheus/procfs v0.12.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/vishvananda/netns v0.0.4 // indirect
	golang.org/x/term v0.17.0 // indirect
	howett.net/plist v1.0.1 // indirect
)

replace github.com/kardianos/service v1.2.2 => github.com/cuonglm/service v0.0.0-20230322120818-ee0647d95905

// 强制使用Go 1.20兼容版本，避免传递依赖拉取新版本
replace (
	github.com/jackpal/gateway => github.com/jackpal/gateway v1.0.10
	golang.org/x/crypto => golang.org/x/crypto v0.17.0
	golang.org/x/net => golang.org/x/net v0.19.0
	golang.org/x/sys => golang.org/x/sys v0.15.0
	golang.org/x/term => golang.org/x/term v0.15.0
)
