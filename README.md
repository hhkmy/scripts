# 🛠️ Development & Automation Scripts

A curated collection of developer installation, maintenance, and automation scripts by [@hhkmy](https://github.com/hhkmy).

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-Linux%20%7C%20macOS-lightgrey.svg)](#)
[![Shell](https://img.shields.io/badge/shell-bash-green.svg)](#)

---

## ⚡ Quick One-Line Execution

Run any script instantly directly from your terminal without cloning the repository:

```bash
# 🚀 Google Antigravity 2.0 (Agentic Desktop)
curl -fsSL https://raw.githubusercontent.com/hhkmy/scripts/main/antigravity | bash

# 🐹 Go Compiler & Version Manager
curl -fsSL https://raw.githubusercontent.com/hhkmy/scripts/main/go | bash

# ⚡ Hugo Extended Installer (with Sass/SCSS)
curl -fsSL https://raw.githubusercontent.com/hhkmy/scripts/main/gohugo | bash -s -- --extended
```

> **Tip:** You can pass command-line arguments to one-liners using `bash -s -- [OPTIONS]`  
> (e.g. `curl -fsSL https://raw.githubusercontent.com/hhkmy/scripts/main/antigravity | bash -s -- --check`).

---

## 📦 Included Scripts

- 🚀 **[`antigravity`](./antigravity)** &nbsp; [![Linux](https://img.shields.io/badge/Platform-Linux%20(x64%2Farm64)-1f2328?style=flat-square&logo=linux)](#)  
  Robust installer & auto-updater for Google Antigravity 2.0 (Agentic Desktop) with dynamic manifest discovery, zero-dependency ASAR parsing, process protection, multi-resolution icon theme integration, and automated rollback.  
  👉 [Usage Guide & Flags](#1-antigravity-20-updater-antigravity)

- 🐹 **[`go`](./go)** &nbsp; [![Linux | macOS](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS-1f2328?style=flat-square)](#)  
  Lightweight Go Version Manager. Fetches official stable releases directly from `go.dev`, manages user-space installs, and seamlessly configures shell environments (`bash`, `zsh`, `fish`).  
  👉 [Usage Guide](#2-go-version-manager-go)

- ⚡ **[`gohugo`](./gohugo)** &nbsp; [![Linux | macOS](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS-1f2328?style=flat-square)](#)  
  Fast installer and updater for Hugo static site generator supporting Standard, Extended (with Sass/SCSS), and Extended with Deploy releases.  
  👉 [Usage Guide & Flags](#3-hugo-extended-installer-gohugo)

---

## 🚀 Script Usage & Details

### 1. Antigravity 2.0 Updater (`antigravity`)

Automated installation and update tool for Google Antigravity 2.0 (Agentic Desktop) on Linux.

**One-line command:**
```bash
curl -fsSL https://raw.githubusercontent.com/hhkmy/scripts/main/antigravity | bash
```

**Local execution:**
```bash
# Check for updates only
./antigravity --check

# Install or update to latest version (User-space: ~/.local/share/antigravity)
./antigravity

# Install or update system-wide (Standard: /opt/Antigravity, requires sudo)
sudo ./antigravity --global

# Force reinstall current version
./antigravity --force

# Install specific version
./antigravity --version 2.19.1

# Restore previous backup (Rollback)
./antigravity --rollback
```

**Features:**
- **Dual-Mode Installation (FHS Standard /opt or User-space):** Supports system-wide standard installation (`/opt/Antigravity`) via `--global` or isolated non-root user installation (`~/.local/share/antigravity`).
- **External Tool & Manager Compatibility:** Automatically synchronizes `resources/app/package.json` loose metadata so tools like [Antigravity-Manager](https://github.com/lbjlaq/Antigravity-Manager) and CLI switchers detect the version instantly without hanging.
- **Zero-Dependency Icon & Version Extraction:** Directly parses `resources/app.asar` using standard library Python 3 to extract version info and high-res app icon without needing `npx` or `node`.
- **System Icon Theme Integration:** Auto-generates multi-resolution icons (16x16 up to 1024x1024) into `~/.local/share/icons/hicolor/` (or `/usr/share/icons/hicolor/`) for crisp display across GNOME/KDE docks, taskbars, and Alt+Tab switchers.
- **Dynamic Manifest Discovery:** Automatically queries active updater URL from local `app-update.yml` and official Google Cloud Run endpoints.
- **Process & Socket Safety:** Gracefully terminates running Antigravity processes (SIGTERM with clean timeout fallback) to prevent filesystem corruption and lockfile lingering.
- **Automated Backup & Rollback:** Archives existing installs to `antigravity-backups/` and rolls back on failure.
- **Desktop & CLI Integration:** Generates `.desktop` launcher with proper `StartupWMClass`, hardens `chrome-sandbox` (SUID in global mode), and registers CLI symlinks/wrappers.

---

### 2. Go Version Manager (`go`)

Quickly install or update the latest stable Go compiler.

**One-line command:**
```bash
curl -fsSL https://raw.githubusercontent.com/hhkmy/scripts/main/go | bash
```

**Local execution:**
```bash
chmod +x go
./go
```

**Features:**
- Automatic architecture detection (`amd64`, `arm64`, `386`, `armv6l`).
- Non-root user space installation to `~/go` (or `/usr/local` if run with sudo).
- Automatically configures `GOROOT` and `PATH` in `~/.bashrc`, `~/.zshrc`, or `~/.config/fish/config.fish`.

---

### 3. Hugo Extended Installer (`gohugo`)

Installs or updates Hugo static site generator directly from official GitHub releases.

**One-line command:**
```bash
curl -fsSL https://raw.githubusercontent.com/hhkmy/scripts/main/gohugo | bash -s -- --extended
```

**Local execution:**
```bash
chmod +x gohugo

# Interactive mode (Prompts for Normal, Extended, or Extended with Deploy)
./gohugo

# Non-interactive: Install Hugo Extended (recommended for Sass/SCSS)
./gohugo --extended

# Install Hugo Extended with Deploy features
./gohugo --withdeploy

# Force reinstall
./gohugo --force
```

---

## 📄 License

This repository is licensed under the [MIT License](LICENSE).
