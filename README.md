<div align="center">

```
╔══════════════════════════════════════════════════════════════════════╗
║                                                                      ║
║    ██████╗ ██╗      ██████╗  ██████╗██╗  ██╗                        ║
║    ██╔══██╗██║     ██╔═══██╗██╔════╝██║ ██╔╝                        ║
║    ██████╔╝██║     ██║   ██║██║     █████╔╝                         ║
║    ██╔══██╗██║     ██║   ██║██║     ██╔═██╗                         ║
║    ██████╔╝███████╗╚██████╔╝╚██████╗██║  ██╗                        ║
║    ╚═════╝ ╚══════╝ ╚═════╝  ╚═════╝╚═╝  ╚═╝                        ║
║                                                                      ║
║     ██████╗██╗  ██╗ █████╗ ██╗███╗   ██╗                            ║
║    ██╔════╝██║  ██║██╔══██╗██║████╗  ██║                            ║
║    ██║     ███████║███████║██║██╔██╗ ██║                            ║
║    ██║     ██╔══██║██╔══██║██║██║╚██╗██║                            ║
║    ╚██████╗██║  ██║██║  ██║██║██║ ╚████║                            ║
║     ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝╚═╝  ╚═══╝                            ║
║                                                                      ║
║          🔗  NETWORKED MULTI-NODE BLOCKCHAIN  🔗                     ║
║                                                                      ║
╚══════════════════════════════════════════════════════════════════════╝
```

<br>

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-yellow?style=for-the-badge)](LICENSE)
[![Build](https://img.shields.io/badge/Build-Passing-brightgreen?style=for-the-badge&logo=github-actions&logoColor=white)]()
[![Tests](https://img.shields.io/badge/Tests-40%20Passing-brightgreen?style=for-the-badge&logo=testcafe&logoColor=white)]()
[![Race Free](https://img.shields.io/badge/Race%20Free-✓-success?style=for-the-badge)]()

<br>

**A peer-to-peer blockchain network built from scratch in Go.**  
*Multiple independent nodes • Ed25519 signed transactions • Gossip protocol • Fork resolution*

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
### Multi-Node Network
Multiple independent nodes running on different ports, communicating via HTTP

</td>
<td align="center" width="25%">

### 🔐
### Ed25519 Crypto
Every transaction is signed and verified using Ed25519 digital signatures

</td>
<td align="center" width="25%">

### 📡
### Gossip Protocol
Transactions and blocks propagate across the network with smart de-duplication

</td>
<td align="center" width="25%">

### ⛓️
### Fork Resolution
Longest-chain rule with automatic reorganisation and orphan recovery

</td>
</tr>
</table>

<br>

## 🎯 Features

```
 ┌─────────────────────────────────────────────────────────────┐
 │                    CORE FEATURES                            │
 ├─────────────────────────────────────────────────────────────┤
 │                                                             │
 │  ✅ Networked Service     HTTP API on configurable ports    │
 │  ✅ Signed Transactions   Ed25519 sign + verify             │
 │  ✅ Transaction Gossip    Forward with de-duplication        │
 │  ✅ Block Gossip          Broadcast mined/received blocks    │
 │  ✅ Chain Sync            Initial + periodic synchronisation │
 │  ✅ Fork Resolution       Longest chain + reorg + orphans    │
 │  ✅ Concurrency Safety    sync.RWMutex, race-free            │
 │  ✅ Introspection API     Status, balances, mempool, peers   │
 │  ✅ Mempool Consistency   Remove mined, return orphaned      │
 │  ✅ Peer Exchange         Dynamic peer discovery              │
 │                                                             │
 └─────────────────────────────────────────────────────────────┘
```

<br>

## 🚀 Quick Start

### Prerequisites

> **Go 1.22+** required • No third-party dependencies • Standard library only

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
# All tests
go test ./...

# With race detector
go test -race ./...
```

<br>

## 🔧 CLI Configuration

| Flag | Default | Description |
|:-----|:--------|:------------|
| `--addr` | `localhost:3001` | 🌐 Listen address for the HTTP API |
| `--peers` | *(empty)* | 🤝 Comma-separated list of peer addresses |
| `--miner` | *(auto-generated)* | ⛏️ Miner address for rewards |
| `--difficulty` | `4` | 💪 Mining difficulty (leading zeros) |
| `--mine` | `true` | ⚡ Enable/disable mining |

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
| `GET` | `/peers` | 🤝 List known peers | ✅ |
| `POST` | `/peers` | 🔄 Exchange peer lists | ✅ |
| `GET` | `/mempool` | 🏊 Pending transactions | ✅ |
| `GET` | `/balances` | 💰 All account balances | ✅ |
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
  "height": 42,
  "head_hash": "0000abc123def456...",
  "peer_count": 2,
  "mempool_size": 3,
  "chain_length": 43
}
```

<br>

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        NODE ARCHITECTURE                         │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│   ┌──────────┐    ┌──────────┐    ┌──────────┐                  │
│   │  Node 1  │◄──►│  Node 2  │◄──►│  Node 3  │    Network      │
│   │  :3001   │    │  :3002   │    │  :3003   │                  │
│   └────┬─────┘    └────┬─────┘    └────┬─────┘                  │
│        │               │               │                         │
│   ┌────▼───────────────▼───────────────▼─────┐                  │
│   │              HTTP API Layer               │  ← handlers.go  │
│   ├───────────────────────────────────────────┤                  │
│   │    Gossip    │    Sync    │    Peers      │  ← network/     │
│   ├───────────────────────────────────────────┤                  │
│   │         Miner          │     Mempool      │  ← miner/       │
│   ├───────────────────────────────────────────┤                  │
│   │   Chain   │  Ledger  │  Block  │   Tx    │  ← blockchain/  │
│   ├───────────────────────────────────────────┤                  │
│   │        Ed25519       │      SHA-256       │  ← crypto/      │
│   └───────────────────────────────────────────┘                  │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 📁 Project Structure

```
📦 go-multinode-blockchain
 ┣━━ 📂 cmd
 ┃   ┗━━ 📂 node
 ┃       ┗━━ 📜 main.go              ← CLI entry point
 ┣━━ 📂 internal
 ┃   ┣━━ 📂 blockchain
 ┃   ┃   ┣━━ 📜 block.go             ← Block struct, PoW mining
 ┃   ┃   ┣━━ 📜 blockchain_test.go   ← 12 tests
 ┃   ┃   ┣━━ 📜 chain.go             ← Chain + fork resolution
 ┃   ┃   ┣━━ 📜 genesis.go           ← Deterministic genesis
 ┃   ┃   ┣━━ 📜 ledger.go            ← Account balances
 ┃   ┃   ┗━━ 📜 transaction.go       ← Tx signing & validation
 ┃   ┣━━ 📂 crypto
 ┃   ┃   ┣━━ 📜 crypto_test.go       ← 7 tests
 ┃   ┃   ┣━━ 📜 hash.go              ← SHA-256 utilities
 ┃   ┃   ┗━━ 📜 keys.go              ← Ed25519 key management
 ┃   ┣━━ 📂 mempool
 ┃   ┃   ┣━━ 📜 mempool_test.go      ← 10 tests
 ┃   ┃   ┗━━ 📜 pool.go              ← Thread-safe tx pool
 ┃   ┣━━ 📂 miner
 ┃   ┃   ┗━━ 📜 miner.go             ← Background PoW mining
 ┃   ┣━━ 📂 network
 ┃   ┃   ┣━━ 📜 gossip.go            ← Gossip protocol
 ┃   ┃   ┣━━ 📜 handlers.go          ← HTTP API (10 endpoints)
 ┃   ┃   ┣━━ 📜 network_test.go      ← 11 tests
 ┃   ┃   ┣━━ 📜 peer.go              ← Peer management
 ┃   ┃   ┣━━ 📜 server.go            ← HTTP server setup
 ┃   ┃   ┗━━ 📜 sync.go              ← Chain synchronisation
 ┃   ┗━━ 📂 node
 ┃       ┗━━ 📜 node.go              ← Node orchestrator
 ┣━━ 📂 scripts
 ┃   ┣━━ 📜 cluster.ps1              ← Windows launcher
 ┃   ┗━━ 📜 cluster.sh               ← Linux/Mac launcher
 ┣━━ 📜 .gitignore
 ┣━━ 📜 go.mod
 ┣━━ 📜 Makefile
 ┣━━ 📜 README.md
 ┗━━ 📜 REPORT.md                    ← Research report
```

<br>

## 🔒 Concurrency Design

All shared state is protected by `sync.RWMutex` — **zero data races** under concurrent mining, gossip, and request handling.

| Component | Lock Type | Strategy |
|:----------|:----------|:---------|
| ⛓️ **Chain** | `sync.RWMutex` | RLock for reads, Lock for adds & reorgs |
| 🏊 **Mempool** | `sync.RWMutex` | RLock for queries, Lock for add/remove |
| 🤝 **PeerManager** | `sync.RWMutex` | RLock for peer list, Lock for mutations |
| 👁️ **SeenTracker** | `sync.Mutex` | Atomic check-and-mark operations |

<br>

## 📡 Gossip Protocol

```
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
    │                    ✅ Total: 2 messages (not ∞)   │
    │                                                    │
```

**De-duplication** prevents message loops:
- `SeenTracker` with 5-minute TTL
- `X-Source-Peer` header excludes the sender
- Cost: **O(N²)** for N nodes (not exponential!)

<br>

## ⛓️ Fork Resolution

```
  Before Reorg:                    After Reorg:
  
  ┌───┐   ┌───┐   ┌───┐          ┌───┐   ┌───┐   ┌───┐   ┌───┐
  │ 0 │──►│ 1 │──►│2A │          │ 0 │──►│ 1 │──►│2B │──►│3B │
  └───┘   └───┘   └─┬─┘          └───┘   └───┘   └───┘   └───┘
                     │                              ▲
                     │ Fork!                        │ Winner!
                     │                              │
                   ┌─▼─┐   ┌───┐                    │
                   │2B │──►│3B │  ← Longer chain ───┘
                   └───┘   └───┘
  
  Orphaned txs from 2A → returned to mempool ♻️
```

<br>

## 🧪 Test Results

<div align="center">

| Package | Tests | Status |
|:--------|:-----:|:------:|
| `internal/crypto` | 7 | ✅ PASS |
| `internal/blockchain` | 12 | ✅ PASS |
| `internal/mempool` | 10 | ✅ PASS |
| `internal/network` | 11 | ✅ PASS |
| **Total** | **40** | **✅ ALL PASS** |

</div>

```
ok   github.com/blockchain2/internal/blockchain   0.602s
ok   github.com/blockchain2/internal/crypto       0.579s
ok   github.com/blockchain2/internal/mempool      0.593s
ok   github.com/blockchain2/internal/network      1.409s
```

<br>

## 📊 Research Report

The full research report is available in [`REPORT.md`](REPORT.md), covering:

| Section | Topic |
|:--------|:------|
| 🏗️ **Design Write-up** | Wire format, concurrency model, reorg process |
| 🔀 **Convergence Experiment** | Fork creation, disconnection, and resolution |
| 📡 **Gossip Cost Analysis** | Message propagation with de-duplication |
| 💬 **Discussion** | Longest-chain rule, finality, 51% attacks, trust |

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

In this Round 2 iteration, the implementation was substantially hardened based on feedback:
- **Cumulative Work vs Length**: Fork resolution now correctly uses cumulative work (`sum(2^difficulty)`) instead of simple chain length, properly adhering to the heaviest-chain rule.
- **Mempool Validation**: Transactions are strictly verified against the *ledger state* before entering the mempool, preventing unspendable or replay transactions from clogging the network.
- **True Concurrency in Mining**: The Miner now correctly listens to `Restart()` signals to abort current proof-of-work immediately when a new block is accepted, eliminating wasted hashing.
- **Equal-Height Fork Retention**: Replaced the simplistic ignore-equal-height logic with a robust `sideChains` structure that preserves competing blocks, preventing chain stalls.
- **Atomic Ledger Updates**: Block application is now fully atomic (clone-then-apply) preventing corrupted state on failure.
- **Persistence & Peer Health**: Implemented saving/loading of the chain state per-node (FR-11) and background peer pinging to evict dead nodes (FR-10).

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

**Built with 💻 and ☕ using only the Go standard library**

⭐ Star this repo if you found it helpful!

</div>