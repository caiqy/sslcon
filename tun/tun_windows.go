package tun

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"

	"golang.org/x/sys/windows"
	"golang.zx2c4.com/wintun"
	"golang.zx2c4.com/wireguard/windows/tunnel/winipcfg"
)

type NativeTun struct {
	wt      *wintun.Adapter
	session wintun.Session
	name    string
	mtu     int

	closeOnce sync.Once
	close     atomic.Bool
}

var (
	WintunTunnelType          = "TLSLink Secure"
	WintunStaticRequestedGUID = &windows.GUID{
		0x0000000,
		0xFFFF,
		0xFFFF,
		[8]byte{0xFF, 0xe9, 0x76, 0xe5, 0x8c, 0x74, 0x06, 0x3e},
	}
)

func CreateTUN(ifname string, mtu int) (Device, error) {
	// Extract embedded wintun.dll before use
	if err := ExtractWintunDLL(); err != nil {
		return nil, fmt.Errorf("failed to extract wintun.dll: %w", err)
	}

	wt, err := wintun.CreateAdapter(ifname, WintunTunnelType, WintunStaticRequestedGUID)
	if err != nil {
		return nil, fmt.Errorf("failed to create adapter: %w", err)
	}

	// 启动会话，容量 8 MiB (0x800000)
	session, err := wt.StartSession(0x800000)
	if err != nil {
		wt.Close()
		return nil, fmt.Errorf("failed to start session: %w", err)
	}

	tun := &NativeTun{
		wt:      wt,
		session: session,
		name:    ifname,
		mtu:     mtu,
	}

	return tun, nil
}

func (tun *NativeTun) File() *os.File {
	return nil
}

func (tun *NativeTun) Read(buff []byte, offset int) (int, error) {
	if tun.close.Load() {
		return 0, os.ErrClosed
	}

retry:
	// 检查关闭状态
	if tun.close.Load() {
		return 0, os.ErrClosed
	}

	packet, err := tun.session.ReceivePacket()
	switch err {
	case nil:
		packetSize := len(packet)
		copy(buff[offset:], packet)
		tun.session.ReleaseReceivePacket(packet)
		return packetSize, nil
	case windows.ERROR_NO_MORE_ITEMS:
		// 没有数据包，等待 500ms 后重试，避免无限阻塞导致无法响应 Close
		windows.WaitForSingleObject(tun.session.ReadWaitEvent(), 500)
		goto retry
	case windows.ERROR_HANDLE_EOF:
		return 0, os.ErrClosed
	case windows.ERROR_INVALID_DATA:
		return 0, errors.New("recv ring corrupt")
	}
	return 0, fmt.Errorf("read failed: %w", err)
}

func (tun *NativeTun) Write(buff []byte, offset int) (int, error) {
	if tun.close.Load() {
		return 0, os.ErrClosed
	}

	packetSize := len(buff) - offset

	packet, err := tun.session.AllocateSendPacket(packetSize)
	if err == nil {
		copy(packet, buff[offset:])
		tun.session.SendPacket(packet)
		return packetSize, nil
	}
	switch err {
	case windows.ERROR_HANDLE_EOF:
		return 0, os.ErrClosed
	case windows.ERROR_BUFFER_OVERFLOW:
		return 0, nil // Dropping when ring is full.
	}
	return 0, fmt.Errorf("write failed: %w", err)
}

func (tun *NativeTun) Flush() error {
	return nil
}

func (tun *NativeTun) MTU() (int, error) {
	return tun.mtu, nil
}

func (tun *NativeTun) Name() (string, error) {
	return tun.name, nil
}

func (tun *NativeTun) Events() <-chan Event {
	return nil
}

func (tun *NativeTun) Close() error {
	tun.closeOnce.Do(func() {
		tun.close.Store(true)

		tun.session.End()
		if tun.wt != nil {
			tun.wt.Close()
		}
	})

	return nil
}

func (tun *NativeTun) LUID() winipcfg.LUID {
	return winipcfg.LUID(tun.wt.LUID())
}
