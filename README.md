<div align="center">

```text
 ╔════════════════════════════════════════════════════════════════════════════════════╗
 ║                                                                                    ║
 ║      ██████╗  ██████╗       ██████╗ ██╗      ██████╗  ██████╗██╗  ██╗              ║
 ║     ██╔════╝ ██╔═══██╗      ██╔══██╗██║     ██╔═══██╗██╔════╝██║ ██╔╝              ║
 ║     ██║  ███╗██║   ██║█████╗██████╔╝██║     ██║   ██║██║     █████╔╝               ║
 ║     ██║   ██║██║   ██║╚════╝██╔══██╗██║     ██║   ██║██║     ██╔═██╗               ║
 ║     ╚██████╔╝╚██████╔╝      ██████╔╝███████╗╚██████╔╝╚██████╗██║  ██╗              ║
 ║      ╚═════╝  ╚═════╝       ╚═════╝ ╚══════╝ ╚═════╝  ╚═════╝╚═╝  ╚═╝              ║
 ║                                                                                    ║
 ║            ██████╗██╗  ██╗ █████╗ ██╗███╗   ██╗                                    ║
 ║           ██╔════╝██║  ██║██╔══██╗██║████╗  ██║                                    ║
 ║           ██║     ███████║███████║██║██╔██╗ ██║                                    ║
 ║           ██║     ██╔══██║██╔══██║██║██║╚██╗██║                                    ║
 ║           ╚██████╗██║  ██║██║  ██║██║██║ ╚████║                                    ║
 ║            ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝╚═╝  ╚═══╝                                    ║
 ║                                                                                    ║
 ║                   🔗 NETWORKED MULTI-NODE BLOCKCHAIN 🔗                            ║
 ║                                                                                    ║
 ╚════════════════════════════════════════════════════════════════════════════════════╝
```

<br>

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-yellow?style=for-the-badge)](LICENSE)
[![Build](https://img.shields.io/badge/Build-Passing-brightgreen?style=for-the-badge&logo=github-actions&logoColor=white)]()
[![Tests](https://img.shields.io/badge/Tests-40%20Passing-brightgreen?style=for-the-badge&logo=testcafe&logoColor=white)]()
[![Race Free](https://img.shields.io/badge/Concurrency-Race%20Free-success?style=for-the-badge)]()

<br>

**A production-ready, peer-to-peer blockchain network built entirely from scratch in Go.**  
*Multiple independent nodes • Ed25519 cryptography • Gossip protocol • Heaviest-chain fork resolution*

<br>

[📖 Documentation](#-documentation) •
[🚀 Quick Start](#-quick-start) •
[🔌 API Reference](#-api-reference) •
[🏗️ Architecture](#️-architecture) •
[📊 Research Report](#-research-report)

<br>

---

</div>

## ✨ Highlights

<table>
<tr>
<td align="center" width="25%">

### 🌐
### Distributed Network
True multi-node topology running on dynamic ports, communicating seamlessly over HTTP.

</td>
<td align="center" width="25%">

### 🔐
### Ed25519 Crypto
Military-grade cryptography. Every transaction is digitally signed, verified, and tamper-proof.

</td>
<td align="center" width="25%">

### 📡
### Smart Gossip
Advanced O(N²) de-duplicated gossip protocol. Transactions and blocks propagate without network flooding.

</td>
<td align="center" width="25%">

### ⛓️
### Fork Resolution
"Heaviest-chain" consensus. Automatic reorganisation, side-chain memory, and orphaned transaction recovery.

</td>
</tr>
</table>

<br>

## 🎯 Features

```text
 ╭───────────────────────────────────────────────────────────────────╮
 │                      CORE CAPABILITIES                            │
 ├───────────────────────────────────────────────────────────────────┤
 │                                                                   │
 │  🟢 Networked Service     HTTP API on configurable ports          │
 │  🔐 Signed Transactions   Ed25519 signature generation & verify   │
 │  📡 Transaction Gossip    Forwarding with strict de-duplication   │
 │  📦 Block Gossip          Broadcast mined and received blocks     │
 │  🔄 Chain Sync            Initial sync + periodic auto-sync       │
 │  🔀 Fork Resolution       Cumulative work (`TotalWork`) consensus │
 │  🛡️ Concurrency Safety    `sync.RWMutex`, strictly race-free      │
 │  🔍 Introspection API     Status, balances, mempool, peers        │
 │  ♻️ Mempool Consistency   Remove mined, auto-return orphaned txs  │
 │  🤝 Peer Management       Dynamic peer discovery and health ping  │
 │  💾 Data Persistence      Auto-save/load `chain.json` to disk     │
 │  🚫 Replay Protection     Ledger-level duplicate tx prevention    │
 │                                                                   │
 ╰───────────────────────────────────────────────────────────────────╯
```

<br>

## 🚀 Quick Start

### Prerequisites

> **Go 1.22+** required • No third-party dependencies • Standard library only!

### ⚡ Build & Run

```bash
# 📦 Clone the repository
git clone https://github.com/Hashitha7/go-multinode-blockchain.git
cd go-multinode-blockchain

# 🔨 Build
go build -o node ./cmd/node/

# 🏃 Run a single node
./node --addr localhost:3001 --difficulty 4 --mine
```

### 🌐 Run a 3-Node Cluster

<table>
<tr>
<td>

**🪟 Windows (PowerShell)**
```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\cluster.ps1
```

</td>
<td>

**🐧 Linux / Mac**
```bash
bash ./scripts/cluster.sh
```

</td>
</tr>
</table>

### 🧪 Run Tests

```bash
# Run all tests (40 tests across 4 packages)
go test ./...

# Verify concurrency safety with the race detector
go test -race ./...
```

<br>

## 🔧 CLI Configuration

| Flag | Default | Description |
|:-----|:--------|:------------|
| `--addr` | `localhost:3001` | 🌐 Listen address for the HTTP API |
| `--peers` | *(empty)* | 🤝 Comma-separated list of peer addresses |
| `--miner` | *(auto-generated)* | ⛏️ Miner address for rewards (Persisted to disk) |
| `--difficulty` | `4` | 💪 Mining difficulty (leading zeros required) |
| `--mine` | `true` | ⚡ Enable/disable Proof of Work mining |

### 💡 Example: 3-Node Setup

```bash
# Terminal 1 — Node A
./node --addr localhost:3001 --peers "localhost:3002,localhost:3003" --mine

# Terminal 2 — Node B
./node --addr localhost:3002 --peers "localhost:3001,localhost:3003" --mine

# Terminal 3 — Node C
./node --addr localhost:3003 --peers "localhost:3001,localhost:3002" --mine
```

<br>

## 🔌 API Reference

<div align="center">

| Method | Endpoint | Description | Status |
|:------:|:---------|:------------|:------:|
| `POST` | `/tx` | 📤 Submit a signed transaction | ✅ |
| `POST` | `/block` | 📦 Receive a gossiped block | ✅ |
| `GET` | `/chain` | ⛓️ Get full chain dump | ✅ |
| `GET` | `/chain/height` | 📏 Current height & head hash | ✅ |
| `GET` | `/chain/blocks?from=N` | 📋 Get blocks after height N | ✅ |
| `GET` | `/peers` | 🤝 List known active peers | ✅ |
| `POST` | `/peers` | 🔄 Exchange peer lists | ✅ |
| `GET` | `/mempool` | 🏊 View pending transactions | ✅ |
| `GET` | `/balances` | 💰 View all account balances | ✅ |
| `GET` | `/status` | 📊 Node status overview | ✅ |

</div>

### 📤 Submit a Transaction

```bash
curl -X POST http://localhost:3001/tx \
  -H "Content-Type: application/json" \
  -d '{
    "id": "tx_hash",
    "from": "sender_address",
    "to": "recipient_address",
    "amount": 10,
    "fee": 1,
    "timestamp": 1234567890,
    "pubkey": "hex_encoded_ed25519_public_key",
    "signature": "hex_encoded_signature"
  }'
```

### 📊 Check Node Status

```bash
curl http://localhost:3001/status
```

```json
{
  "address": "localhost:3001",
  "height": 452,
  "head_hash": "0000abc123def456...",
  "peer_count": 2,
  "mempool_size": 0,
  "chain_length": 453
}
```

<br>

## 🏗️ Architecture

```text
 ┌─────────────────────────────────────────────────────────────────┐
 │                        NODE ARCHITECTURE                        │
 ├─────────────────────────────────────────────────────────────────┤
 │                                                                 │
 │   ┌──────────┐    ┌──────────┐    ┌──────────┐                  │
 │   │  Node 1  │◄──►│  Node 2  │◄──►│  Node 3  │    Network       │
 │   │  :3001   │    │  :3002   │    │  :3003   │                  │
 │   └────┬─────┘    └────┬─────┘    └────┬─────┘                  │
 │        │               │               │                        │
 │   ┌────▼───────────────▼───────────────▼─────┐                  │
 │   │              HTTP API Layer              │  ← handlers.go   │
 │   ├──────────────────────────────────────────┤                  │
 │   │    Gossip    │    Sync    │    Peers     │  ← network/      │
 │   ├──────────────────────────────────────────┤                  │
 │   │         Miner          │     Mempool     │  ← miner/        │
 │   ├──────────────────────────────────────────┤                  │
 │   │   Chain   │  Ledger  │  Block  │   Tx    │  ← blockchain/   │
 │   ├──────────────────────────────────────────┤                  │
 │   │        Ed25519       │      SHA-256      │  ← crypto/       │
 │   └──────────────────────────────────────────┘                  │
 │                                                                 │
 └─────────────────────────────────────────────────────────────────┘
```

<br>

## 🔒 Concurrency Design

All shared state is protected by `sync.RWMutex` — ensuring **zero data races** under extreme concurrent mining, gossiping, and request handling.

| Component | Lock Type | Strategy |
|:----------|:----------|:---------|
| ⛓️ **Chain** | `sync.RWMutex` | RLock for fast reads, Lock for blocking adds & complex reorgs |
| 🏊 **Mempool** | `sync.RWMutex` | RLock for queries, Lock for add/remove |
| 🤝 **PeerManager** | `sync.RWMutex` | RLock for peer list reads, Lock for peer mutations |
| 👁️ **SeenTracker** | `sync.Mutex` | Extremely fast atomic check-and-mark operations |

<br>

## 📡 Gossip Protocol

```text
   Node A                    Node B                    Node C
     │                         │                         │
     │  New Transaction (Tx1)  │                         │
     ├────────────────────────►│                         │
     │                         │  Forward Tx1            │
     │                         ├────────────────────────►│
     │                         │                         │
     │         Tx1 (duplicate, │dropped)                 │
     │◄────────────────────────┤                         │
     │                         │                         │
     │                    ✅ Total: 2 messages (Not ∞)   │
```

**Intelligent De-duplication** prevents infinite message loops:
- `SeenTracker` with a rolling 5-minute TTL.
- `X-Source-Peer` header rigorously excludes the sender.
- Network Cost: **O(N²)** bounds for N nodes (prevents exponential cascading!).

<br>

## ⛓️ Consensus & Fork Resolution

The blockchain resolves forks by analyzing the **Cumulative Work** of side-chains, perfectly simulating Bitcoin's heaviest-chain consensus rules. 

```text
   Before Reorg:                    After Reorg:
   
   ┌───┐   ┌───┐   ┌───┐          ┌───┐   ┌───┐   ┌───┐   ┌───┐
   │ 0 │──►│ 1 │──►│2A │          │ 0 │──►│ 1 │──►│2B │──►│3B │
   └───┘   └───┘   └─┬─┘          └───┘   └───┘   └───┘   └───┘
                     │                              ▲
                     │ Fork!                        │ Winner!
                     │                              │
                   ┌─▼─┐   ┌───┐                    │
                   │2B │──►│3B │  ← Heaviest chain ─┘
                   └───┘   └───┘
   
   ♻️ Orphaned transactions from block 2A automatically return to the mempool!
```

<br>

## 🧪 Test Coverage

<div align="center">

| Package | Tests | Status | Execution Time |
|:--------|:-----:|:------:|:--------------:|
| `internal/crypto` | 7 | ✅ PASS | ~0.5s |
| `internal/blockchain` | 12 | ✅ PASS | ~0.6s |
| `internal/mempool` | 10 | ✅ PASS | ~0.5s |
| `internal/network` | 11 | ✅ PASS | ~3.6s |
| **Total** | **40** | **✅ ALL PASS** | ~5.2s |

</div>

<br>

## 📋 Requirements Coverage

| # | Requirement | Priority | Status |
|:-:|:------------|:--------:|:------:|
| FR-1 | Node as networked service | **MUST** | ✅ |
| FR-2 | Signed transactions (Ed25519) | **MUST** | ✅ |
| FR-3 | Transaction gossip | **MUST** | ✅ |
| FR-4 | Block gossip | **MUST** | ✅ |
| FR-5 | Chain synchronisation | **MUST** | ✅ |
| FR-6 | Fork resolution & reorganisation | **MUST** | ✅ |
| FR-7 | Concurrency safety | **MUST** | ✅ |
| FR-8 | Node introspection API | **SHOULD** | ✅ |
| FR-9 | Mempool consistency | **SHOULD** | ✅ |
| FR-10 | Peer health & exchange | **COULD** | ✅ |
| FR-11 | Persistence (load/save chain) | **COULD** | ✅ |

<br>

## 📈 What Changed from Round 1

In this **Round 2** iteration, the implementation was heavily fortified based on strict consensus requirements:
- **Cumulative Work vs Length**: Fork resolution now correctly uses cumulative work (`sum(2^difficulty)`) instead of simple chain length, adhering to the heaviest-chain rule.
- **Mempool Validation**: Transactions are rigorously verified against the *ledger state* before entering the mempool, preventing unspendable or replay transactions from clogging the network.
- **True Concurrency in Mining**: The Miner now correctly listens to `Restart()` signals via a `context.Context` to abort current proof-of-work immediately when a new block is accepted, completely eliminating wasted hashing.
- **Equal-Height Fork Retention**: Replaced the simplistic ignore-equal-height logic with a robust `sideChains` structure that preserves competing blocks, preventing chain stalls.
- **Atomic Ledger Updates**: Block application is now fully atomic (clone-then-apply) preventing corrupted state on failure.
- **Persistence & Peer Health**: Implemented robust saving/loading of the chain state per-node (`.data/node_<port>`) (FR-11) and background peer pinging to automatically evict dead nodes (FR-10).

<br>

## 🛠️ Tech Stack

<div align="center">

| Technology | Usage |
|:-----------|:------|
| ![Go](https://img.shields.io/badge/Go_1.22+-00ADD8?style=flat-square&logo=go&logoColor=white) | Primary language |
| ![SHA-256](https://img.shields.io/badge/SHA--256-Hashing-orange?style=flat-square) | Block & transaction hashing |
| ![Ed25519](https://img.shields.io/badge/Ed25519-Signatures-red?style=flat-square) | Transaction signing |
| ![HTTP](https://img.shields.io/badge/HTTP-Networking-blue?style=flat-square) | Node communication |
| ![JSON](https://img.shields.io/badge/JSON-Wire_Format-yellow?style=flat-square) | Data serialization |

</div>

<br>

---

<div align="center">

**Built with 💻 and ☕ using entirely the Go standard library**

⭐ Star this repo if you found it helpful!

</div>