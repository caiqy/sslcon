# Design: sslcon Windows 7 32-bit (x86) Support

**Date:** 2025-01-01  
**Branch:** `win7-32bit`  
**Status:** Approved

---

## Background

The `win7-32bit` branch has already completed the foundational work for Windows 7 compatibility:

- Downgraded to Go 1.20 (last version with official Windows 7 support)
- Pinned all transitive dependencies to Go 1.20-compatible versions
- Replaced `wintun-go` with the official `golang.zx2c4.com/wintun` library
- Embedded `wintun_amd64.dll` and `wintun_arm64.dll` into the binary with automatic extraction at runtime
- Fixed Windows TUN read-blocking and goroutine panic stability issues

However, `GOARCH=386` compilation currently fails:

```
# sslcon/tun
tun\tun_windows.go:37:12: undefined: ExtractWintunDLL
```

The root cause: `ExtractWintunDLL()` is only defined in `wintun_embed_windows.go` (amd64) and `wintun_embed_windows_arm64.go` (arm64). There is no x86 counterpart, and no `wintun_386.dll` in the repository.

---

## Goal

Enable `GOOS=windows GOARCH=386` compilation and release packaging for binaries intended for Windows 7 32-bit. Add CI validation to ensure x86 buildability is maintained on every push.

**Delivery scope:** This spec covers compilation correctness and CI verification only. Runtime validation on a physical or virtual Windows 7 SP1 x86 environment is out of scope for CI (GitHub-hosted runners do not support Win7) and must be performed manually as a separate post-implementation acceptance activity.

---

## Design

### Principle

Mirror the existing amd64 and arm64 patterns exactly. No business logic changes. No changes to `tun_windows.go`. The Go build system selects the correct embed file via build tags.

### Architecture

```
tun/
  dll/
    wintun_amd64.dll              ✅ existing (427,552 bytes)
    wintun_arm64.dll              ✅ existing (222,488 bytes)
    wintun_386.dll                🆕 new — x86 DLL from Wintun 0.14.1 official release

  wintun_embed_windows.go        ✅ existing  build tag: windows && amd64
  wintun_embed_windows_arm64.go  ✅ existing  build tag: windows && arm64
  wintun_embed_windows_386.go    🆕 new       build tag: windows && 386

.github/workflows/
  release.yml    ✏️  add windows-386 matrix entry
  ci.yml         ✏️  add GOARCH=386 build verification step
```

### Runtime flow (x86, unchanged from amd64)

1. `CreateTUN()` in `tun_windows.go` calls `ExtractWintunDLL()` (now defined for 386 via build tag)
2. `ExtractWintunDLL()` uses `sync.Once` to write the embedded `wintunDLL` bytes to `<exe-dir>/wintun.dll`
3. Size-check skips re-extraction if DLL already exists with matching size
4. `wintun.CreateAdapter()` loads `wintun.dll` from the exe directory

---

## Changes

### 1. `tun/dll/wintun_386.dll` (new binary file)

Source: official Wintun 0.14.1 release ZIP (`https://www.wintun.net/builds/wintun-0.14.1.zip`)  
ZIP SHA2-256: `07c256185d6ee3652e09fa55c0b673e2624b565e02c4b9091c79ca7d2f24ef51`  
Path inside ZIP: `wintun/bin/x86/wintun.dll`

This is the only signed and officially supported distribution of Wintun x86. The version must match the existing `wintun_amd64.dll` and `wintun_arm64.dll` (both from Wintun 0.14.1).

The implementer must verify the extracted `x86/wintun.dll` matches the known file size before committing: the Wintun 0.14.1 x86 DLL is **550,928 bytes**.

---

### 2. `tun/wintun_embed_windows_386.go` (new file)

```go
//go:build windows && 386

package tun

import (
	_ "embed"
	"os"
	"path/filepath"
	"sync"
)

//go:embed dll/wintun_386.dll
var wintunDLL []byte

var extractOnce sync.Once
var extractErr error

// ExtractWintunDLL extracts the embedded wintun.dll to the executable directory.
// This must be called before any wintun operations.
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
```

This file is a verbatim copy of `wintun_embed_windows.go` with two changes:
- Line 1: build tag changed from `windows && amd64` to `windows && 386`
- Line 12: embed directive changed from `dll/wintun_amd64.dll` to `dll/wintun_386.dll`

---

### 3. `.github/workflows/release.yml` — add `windows-386` matrix entry

Add to the `matrix.include` list (alongside the existing `windows` and `windows-arm64` entries):

```yaml
- build: windows-386
  os: windows-2022
  go: '1.20'
  archive-name: sslcon-windows7-386.7z
```

Add to the `Build` step shell script:

```bash
elif [ "${{ matrix.build }}" = "windows-386" ]; then
  GOARCH=386 go build -trimpath -ldflags "-s -w" -o vpnagent.exe vpnagent.go
  GOARCH=386 go build -trimpath -ldflags "-s -w" -o sslcon.exe sslcon.go
```

The `Build archive` step already handles all `windows*` builds with the same 7z logic, so no changes are needed there. Do not modify any existing matrix entries.

---

### 4. `.github/workflows/ci.yml` — add GOARCH=386 build verification

Only in the `windows-latest` runner (not `windows-11-arm`). Inside the existing `windows*` branch of the `Build` step, append x86 cross-compilation after the default amd64 build:

```bash
elif [[ "${{ matrix.os }}" == windows-latest* ]]; then
  go build -trimpath -ldflags "-s -w" -o vpnagent.exe vpnagent.go
  go build -trimpath -ldflags "-s -w" -o sslcon.exe sslcon.go
  # Verify Windows x86 cross-compilation (must set both GOOS and GOARCH to target _windows.go paths)
  GOOS=windows GOARCH=386 go build -trimpath -ldflags "-s -w" -o vpnagent_386.exe vpnagent.go
  GOOS=windows GOARCH=386 go build -trimpath -ldflags "-s -w" -o sslcon_386.exe sslcon.go
elif [[ "${{ matrix.os }}" == windows* ]]; then
  go build -trimpath -ldflags "-s -w" -o vpnagent.exe vpnagent.go
  go build -trimpath -ldflags "-s -w" -o sslcon.exe sslcon.go
```

`GOOS=windows` is required even when running on a Windows runner to ensure the `_windows.go` build constraint paths (including `tun_windows.go` and `wintun_embed_windows_386.go`) are correctly activated for the 386 target. No artifact upload — this step is compile-only verification. The `windows-11-arm` runner builds its native arm64 target only; 386 cross-compilation is not required there.

---

## Constraints

- **Go version stays at 1.20.** No toolchain directive added. Go 1.21+ dropped Windows 7 support.
- **No changes to `tun_windows.go`.** The `CreateTUN()` call to `ExtractWintunDLL()` is architecture-agnostic; the build tag system resolves the correct implementation.
- **No changes to any other business logic files.**
- **Wintun DLL version must be consistent** across amd64, arm64, and x86 (all 0.14.1).
- **CI cannot verify runtime behavior on Windows 7 32-bit.** GitHub-hosted runners do not provide Win7 environments. Runtime validation requires a dedicated Win7 SP1 x86 VM.

---

## Acceptance Criteria

| # | Criterion | How to verify |
|---|-----------|---------------|
| 1 | `GOOS=windows GOARCH=386 go build … vpnagent.go` exits with code 0, triggering `tun_windows.go` and `wintun_embed_windows_386.go` | Run on Windows runner or observe CI green |
| 2 | `GOOS=windows GOARCH=386 go build … sslcon.go` exits with code 0 | Run on Windows runner or observe CI green |
| 3 | CI `ci.yml` passes on `windows-latest` with both amd64 and 386 build steps | GitHub Actions check |
| 4 | `release.yml` produces `sslcon-windows7-386.7z` on workflow dispatch | Check Release assets |
| 5 | `tun/dll/wintun_386.dll` is 550,928 bytes (Wintun 0.14.1 x86) | `ls -la tun/dll/` |

## Post-implementation Manual Validation (out of scope for this spec)

The following validation is recommended after delivery but is not a blocking acceptance condition for this spec:

- `vpnagent.exe` (386) starts on Win7 SP1 x86 and successfully extracts `wintun.dll` to its directory

---

## File Change Summary

| File | Action | Lines changed |
|------|--------|---------------|
| `tun/dll/wintun_386.dll` | Add (binary) | — |
| `tun/wintun_embed_windows_386.go` | Add | 45 lines |
| `.github/workflows/release.yml` | Modify | ~8 lines |
| `.github/workflows/ci.yml` | Modify | ~4 lines |

Total source line changes: ~57 lines. Zero business logic changes.
