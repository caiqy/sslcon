# Windows 7 32-bit (x86) Support Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

---

## Session Context (手动更新，2025-01-01)

### 背景
- `anylink-client` 当前处于 `win7-32bit` 分支；`sslcon` 已作为 git 子模块挂载于 `3rdparty/sslcon`，跟踪 `sslcon` 仓库的 `win7-32bit` 分支。
- `sslcon` 的 `win7-32bit` 分支已完成：Go 1.20 降级、依赖锁定、wintun-go 替换为官方 `golang.zx2c4.com/wintun`、amd64/arm64 DLL 内嵌。
- 当前 `GOARCH=386` 编译失败原因：`tun/tun_windows.go` 调用 `ExtractWintunDLL()`，但 x86 对应的嵌入文件缺失。

### 已完成工作
- [x] 将 `sslcon` 作为子模块添加到 `anylink-client`（commit `cb62201`）
- [x] 完成 brainstorming 全流程
- [x] 编写设计文档（commit `a23573c`）：`docs/superpowers/specs/2025-01-01-win7-32bit-support-design.md`
- [x] 编写本实现计划（commit `d38d9bc`）

### 待完成工作（本 plan 全部 4 个 Task 均未开始）
- Task 1：下载并验证 Wintun 0.14.1 x86 DLL → 放入 `tun/dll/wintun_386.dll`
- Task 2：创建 `tun/wintun_embed_windows_386.go` → 验证 x86 编译 → 提交
- Task 3：修改 `ci.yml` → 添加 386 编译验证（仅 `windows-latest`）→ 提交
- Task 4：修改 `release.yml` → 添加 `windows-386` 矩阵条目 → 提交

### 关键约束
- **工作目录：** 所有命令在 `3rdparty/sslcon`（sslcon 模块根目录）下以 **PowerShell** 执行
- **不修改任何业务逻辑文件**，只新增/修改文件如上所列
- Wintun 0.14.1 x86 DLL 信息：
  - 下载地址：`https://www.wintun.net/builds/wintun-0.14.1.zip`
  - ZIP SHA256：`07c256185d6ee3652e09fa55c0b673e2624b565e02c4b9091c79ca7d2f24ef51`
  - ZIP 内路径：`wintun/bin/x86/wintun.dll`
  - 文件大小：**184,664 字节**
- `ci.yml`：现有 `windows*` 分支需拆分为 `windows-latest`（含 386 验证）和 `windows*`（arm 等）两段
- `release.yml`：`Build archive` 步骤的 `windows*` glob 已自动匹配 `windows-386`，无需改动

---

**Goal:** Add `GOOS=windows GOARCH=386` compile support to sslcon by embedding the official Wintun x86 DLL and wiring it into CI and release workflows.

**Architecture:** Mirror the existing amd64/arm64 DLL-embedding pattern exactly. A new build-tag-guarded file `tun/wintun_embed_windows_386.go` provides `ExtractWintunDLL()` for the 386 target; no business logic files change. CI gains a 386 compile-verification step on `windows-latest`; release gains a `windows-386` matrix entry producing `sslcon-windows7-386.7z`.

**Tech Stack:** Go 1.20, `//go:embed`, GitHub Actions, 7-Zip

> **Shell convention:** All local commands in this plan are run in PowerShell from the `3rdparty/sslcon` directory (the sslcon module root), unless otherwise noted.

---

## Chunk 1: Add wintun x86 DLL and embed file

### Task 1: Download and verify the official Wintun 0.14.1 x86 DLL

**Files:**
- Create: `tun/dll/wintun_386.dll`

- [ ] **Step 1: Download the Wintun 0.14.1 release ZIP**

  Navigate to `https://www.wintun.net/builds/wintun-0.14.1.zip` and download it, or run:

  ```powershell
  Invoke-WebRequest -Uri https://www.wintun.net/builds/wintun-0.14.1.zip -OutFile wintun-0.14.1.zip
  ```

- [ ] **Step 2: Verify the ZIP SHA256**

  ```powershell
  Get-FileHash wintun-0.14.1.zip -Algorithm SHA256
  # Expected: 07c256185d6ee3652e09fa55c0b673e2624b565e02c4b9091c79ca7d2f24ef51
  ```

  If the hash does not match, do not proceed — re-download from the official site.

- [ ] **Step 3: Extract the x86 DLL**

  ```powershell
  Expand-Archive wintun-0.14.1.zip -DestinationPath wintun-extracted
  ```

  The x86 DLL is located at: `wintun-extracted\wintun\bin\x86\wintun.dll`

- [ ] **Step 4: Verify the extracted DLL file size**

  ```powershell
  (Get-Item "wintun-extracted\wintun\bin\x86\wintun.dll").Length
  # Expected: 184664
  ```

  If the size does not match 184,664 bytes, do not proceed.

- [ ] **Step 5: Copy DLL to repository**

  ```powershell
  Copy-Item "wintun-extracted\wintun\bin\x86\wintun.dll" "tun\dll\wintun_386.dll"
  ```

- [ ] **Step 6: Confirm all three DLLs are present**

  ```powershell
  Get-ChildItem tun\dll\ | Select-Object Name, Length
  # Expected:
  # Name             Length
  # ----             ------
  # wintun_386.dll   184664
  # wintun_amd64.dll 427552
  # wintun_arm64.dll 222488
  ```

---

### Task 2: Create the x86 embed file

**Files:**
- Create: `tun/wintun_embed_windows_386.go`
- Reference: `tun/wintun_embed_windows.go` (amd64 counterpart — identical logic, only 2 lines differ)

- [ ] **Step 1: Verify the amd64 reference file**

  Open `tun/wintun_embed_windows.go` and confirm it matches this structure exactly (45 lines):

  ```go
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

  func ExtractWintunDLL() error { ... }
  ```

- [ ] **Step 2: Create `tun/wintun_embed_windows_386.go`**

  Create the file with the following exact content (verbatim copy of the amd64 file with 2 lines changed):

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
  ```

  The two differences from `wintun_embed_windows.go`:
  - Line 1: `//go:build windows && 386` (was `windows && amd64`)
  - Line 12: `//go:embed dll/wintun_386.dll` (was `dll/wintun_amd64.dll`)

- [ ] **Step 3: Verify x86 compilation succeeds**

  Run from the sslcon module root:

  ```powershell
  $env:GOOS="windows"; $env:GOARCH="386"
  go build -trimpath -ldflags "-s -w" -o vpnagent_386.exe vpnagent.go
  go build -trimpath -ldflags "-s -w" -o sslcon_386.exe sslcon.go
  Remove-Item Env:GOOS, Env:GOARCH
  ```

  Expected: both commands exit with code 0, no errors.  
  If you see `undefined: ExtractWintunDLL`, the build tag or file name is wrong — verify line 1 of `tun/wintun_embed_windows_386.go` is exactly `//go:build windows && 386` and the file is named `wintun_embed_windows_386.go`.

- [ ] **Step 4: Verify the 386 binary embeds the DLL (size check)**

  ```powershell
  # The 386 binary must be larger than the DLL alone (184664 bytes)
  (Get-Item vpnagent_386.exe).Length
  # Expected: greater than 184664
  ```

- [ ] **Step 5: Clean up temp build artifacts**

  ```powershell
  Remove-Item vpnagent_386.exe, sslcon_386.exe -ErrorAction SilentlyContinue
  ```

- [ ] **Step 6: Commit**

  ```powershell
  git add tun/dll/wintun_386.dll tun/wintun_embed_windows_386.go
  git commit -m "add wintun x86 DLL and embed file for Windows 7 32-bit support"
  ```

---

## Chunk 2: Update CI and release workflows

### Task 3: Update `ci.yml` to verify x86 compilation on every push

**Files:**
- Modify: `.github/workflows/ci.yml` (lines 37–39, the `windows*` branch of the Build step)

- [ ] **Step 1: Open `.github/workflows/ci.yml` and locate the Build step**

  Find this block (around lines 37–39):

  ```yaml
        elif [[ "${{ matrix.os }}" == windows* ]]; then
            go build -trimpath -ldflags "-s -w" -o vpnagent.exe vpnagent.go
            go build -trimpath -ldflags "-s -w" -o sslcon.exe sslcon.go
  ```

- [ ] **Step 2: Split the `windows*` branch to add 386 verification only on `windows-latest`**

  Replace the `windows*` block with:

  ```yaml
        elif [[ "${{ matrix.os }}" == windows-latest ]]; then
            go build -trimpath -ldflags "-s -w" -o vpnagent.exe vpnagent.go
            go build -trimpath -ldflags "-s -w" -o sslcon.exe sslcon.go
            # Verify Windows x86 cross-compilation
            GOOS=windows GOARCH=386 go build -trimpath -ldflags "-s -w" -o vpnagent_386.exe vpnagent.go
            GOOS=windows GOARCH=386 go build -trimpath -ldflags "-s -w" -o sslcon_386.exe sslcon.go
        elif [[ "${{ matrix.os }}" == windows* ]]; then
            go build -trimpath -ldflags "-s -w" -o vpnagent.exe vpnagent.go
            go build -trimpath -ldflags "-s -w" -o sslcon.exe sslcon.go
  ```

  Note: No artifact upload for `_386.exe` files. The `windows-11-arm` runner hits the second branch and builds arm64 only.

- [ ] **Step 3: Verify the full `ci.yml` Build step looks correct after the edit**

  The complete `Build` step `run:` block should now be:

  ```yaml
      run: |
        if [[ "${{ matrix.os }}" == ubuntu* ]]; then
          go build -trimpath -ldflags "-s -w" -o vpnagent vpnagent.go
          go build -trimpath -ldflags "-s -w" -o sslcon sslcon.go
        elif [[ "${{ matrix.os }}" == windows-latest ]]; then
          go build -trimpath -ldflags "-s -w" -o vpnagent.exe vpnagent.go
          go build -trimpath -ldflags "-s -w" -o sslcon.exe sslcon.go
          # Verify Windows x86 cross-compilation
          GOOS=windows GOARCH=386 go build -trimpath -ldflags "-s -w" -o vpnagent_386.exe vpnagent.go
          GOOS=windows GOARCH=386 go build -trimpath -ldflags "-s -w" -o sslcon_386.exe sslcon.go
        elif [[ "${{ matrix.os }}" == windows* ]]; then
          go build -trimpath -ldflags "-s -w" -o vpnagent.exe vpnagent.go
          go build -trimpath -ldflags "-s -w" -o sslcon.exe sslcon.go
        elif [[ "${{ matrix.os }}" == macos* ]]; then
          go build -trimpath -ldflags "-s -w" -o vpnagent vpnagent.go
          go build -trimpath -ldflags "-s -w" -o sslcon sslcon.go
        fi
  ```

- [ ] **Step 4: Commit**

  ```powershell
  git add .github/workflows/ci.yml
  git commit -m "add GOOS=windows GOARCH=386 compile verification to CI"
  ```

---

### Task 4: Update `release.yml` to produce `sslcon-windows7-386.7z`

**Files:**
- Modify: `.github/workflows/release.yml` (matrix `build` list, `matrix.include`, and `Build` step)

- [ ] **Step 1: Add `windows-386` to the `build` matrix list**

  Find line 18:

  ```yaml
        build: [linux, linux-arm64, linux-armv7, linux-riscv64, linux-mipsle, windows, windows-arm64, macos-arm64]
  ```

  Replace with:

  ```yaml
        build: [linux, linux-arm64, linux-armv7, linux-riscv64, linux-mipsle, windows, windows-386, windows-arm64, macos-arm64]
  ```

- [ ] **Step 2: Add the `windows-386` entry to `matrix.include`**

  Find the existing `windows-arm64` entry (around lines 44–47):

  ```yaml
          - build: windows-arm64
            os: windows-11-arm
            go: '1.20'
            archive-name: sslcon-windows10-arm64.7z
  ```

  Insert the new `windows-386` entry **before** `windows-arm64`:

  ```yaml
          - build: windows-386
            os: windows-2022
            go: '1.20'
            archive-name: sslcon-windows7-386.7z
          - build: windows-arm64
            os: windows-11-arm
            go: '1.20'
            archive-name: sslcon-windows10-arm64.7z
  ```

- [ ] **Step 3: Add the `windows-386` build command to the `Build` step**

  Find the existing `windows` branch (around lines 80–82):

  ```yaml
          elif [ "${{ matrix.build }}" = "windows" ]; then
            go build -trimpath -ldflags "-s -w" -o vpnagent.exe vpnagent.go
            go build -trimpath -ldflags "-s -w" -o sslcon.exe sslcon.go
  ```

  Insert the `windows-386` branch **before** the `windows` branch:

  ```yaml
          elif [ "${{ matrix.build }}" = "windows-386" ]; then
            GOOS=windows GOARCH=386 go build -trimpath -ldflags "-s -w" -o vpnagent.exe vpnagent.go
            GOOS=windows GOARCH=386 go build -trimpath -ldflags "-s -w" -o sslcon.exe sslcon.go
          elif [ "${{ matrix.build }}" = "windows" ]; then
            go build -trimpath -ldflags "-s -w" -o vpnagent.exe vpnagent.go
            go build -trimpath -ldflags "-s -w" -o sslcon.exe sslcon.go
  ```

  Note: The `Build archive` step uses `if [[ "${{ matrix.build }}" = windows* ]]` which already matches `windows-386`, so it requires no changes.

- [ ] **Step 4: Verify the complete `release.yml` matrix and Build step are correct**

  Confirm:
  1. `build:` list contains `windows-386`
  2. `matrix.include` contains the `windows-386` entry with `os: windows-2022`, `go: '1.20'`, `archive-name: sslcon-windows7-386.7z`
  3. Build step has `elif [ "${{ matrix.build }}" = "windows-386" ]` with `GOOS=windows GOARCH=386`
  4. No existing entries have been modified

- [ ] **Step 5: Commit**

  ```powershell
  git add .github/workflows/release.yml
  git commit -m "add windows-386 release target producing sslcon-windows7-386.7z"
  ```

---

## Acceptance Verification

- [ ] **Verify all three DLLs exist with correct sizes**

  ```powershell
  Get-ChildItem tun\dll\ | Select-Object Name, Length
  # Expected:
  # Name             Length
  # ----             ------
  # wintun_386.dll   184664
  # wintun_amd64.dll 427552
  # wintun_arm64.dll 222488
  ```

- [ ] **Verify x86 compilation passes locally**

  ```powershell
  $env:GOOS="windows"; $env:GOARCH="386"
  go build -trimpath -ldflags "-s -w" -o vpnagent_386.exe vpnagent.go
  go build -trimpath -ldflags "-s -w" -o sslcon_386.exe sslcon.go
  # Both must exit with code 0
  Remove-Item -ErrorAction SilentlyContinue vpnagent_386.exe, sslcon_386.exe
  Remove-Item Env:GOOS, Env:GOARCH
  ```

- [ ] **Verify CI passes on `windows-latest` with 386 steps (Acceptance Criterion 3)**

  Push the branch to remote and observe GitHub Actions:
  1. Go to **Actions → CI** in the GitHub repository
  2. Find the run triggered by the push
  3. Open the **"Build on windows-latest"** job
  4. Confirm the `Build` step completes successfully — it should show output for both the default `vpnagent.exe` build and the `GOOS=windows GOARCH=386` builds

- [ ] **Verify `release.yml` produces `sslcon-windows7-386.7z` (Acceptance Criterion 4)**

  Trigger the release workflow manually:
  1. Go to **Actions → Release** in the GitHub repository
  2. Click **"Run workflow"** → select the branch → confirm
  3. Once the `windows-386` job completes, go to **Releases**
  4. Confirm an asset named `sslcon-windows7-386.7z` is attached to the release
