//go:build windows && amd64

package tun

import (
	_ "embed"
	"os"
	"path/filepath"
	"sync"
)

//go:embed dll/wintun_amd64.dll
var wintunDLL []byte

var extractOnce sync.Once
var extractErr error

// ExtractWintunDLL extracts the embedded wintun.dll to the executable directory
// This must be called before any wintun operations
func ExtractWintunDLL() error {
	extractOnce.Do(func() {
		exePath, err := os.Executable()
		if err != nil {
			extractErr = err
			return
		}
		exeDir := filepath.Dir(exePath)
		dllPath := filepath.Join(exeDir, "wintun.dll")

		// Check if DLL already exists and has correct size
		if info, err := os.Stat(dllPath); err == nil {
			if info.Size() == int64(len(wintunDLL)) {
				return // DLL already exists with correct size
			}
		}

		// Extract the DLL
		if len(wintunDLL) == 0 {
			extractErr = os.ErrNotExist
			return
		}
		extractErr = os.WriteFile(dllPath, wintunDLL, 0644)
	})
	return extractErr
}
