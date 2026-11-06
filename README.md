# 🦉 OHoneyToken

<p align="center">
  <img src="https://img.shields.io/badge/Platform-Linux-0a0c10?style=flat-square"/>
  <img src="https://img.shields.io/badge/Category-Forensics%20/%20DFIR-blue?style=flat-square"/>
  <img src="https://img.shields.io/badge/License-MIT-green?style=flat-square"/>
  <img src="https://img.shields.io/badge/Part%20of-OwlSec%20Toolkit-7b5ea7?style=flat-square"/>
  <img src="https://img.shields.io/badge/Version-1.0-cyan?style=flat-square"/>
</p>

> **OHoneyToken** — Honeytoken/canary forge — decoy credentials, deployable files + manifest, trigger correlation ║

---

> ⚠️ **AUTHORISED USE ONLY** — Use only on systems you own or have explicit written permission to test. Unauthorised use is illegal.

---

## 📌 Overview

OHoneyToken is an offline-first OwlSec tool for Honeytoken/canary forge — decoy credentials, deployable files + manifest, trigger correlation. It provides an interactive cyan terminal UI and a full CLI with multi-format export (TXT, Markdown, JSON, CSV, SARIF), designed for authorized security work on Linux.

---

## ✨ Features

- Forge honeytokens and canary credentials for deception ops
- Generate deployable decoy files with a tracking manifest
- Correlate trigger events back to planted tokens
- Support multiple decoy formats suited to common lure paths
- Offline-first builder with clear operator guidance
- Export manifests for blue-team monitoring hooks

---

## ⚙️ Requirements

- **Linux**
- **Go 1.21+** (stdlib only; `CGO_ENABLED=0`)

---

## 🚀 Usage

```bash
go build -ldflags="-s -w" -o OHoneyToken OHoneyToken.go
./OHoneyToken --help
./OHoneyToken --forge --format json
./OHoneyToken --manifest tokens.json --correlate triggers.log
./OHoneyToken --demo
```

---

## 📦 Part of OwlSec Toolkit

This tool is part of the **OwlSec** suite — a collection of 300+ security and privacy tools.

🔗 [owlsec.org](https://owlsec.org)

---

## ©️ License

MIT License — © Khaled S. Haddad
