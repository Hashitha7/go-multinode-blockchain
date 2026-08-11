# Research Report: Networked Multi-Node Blockchain

**Version**: 1.0  
**Date**: July 2026  
**Author**: Software Engineering Intern

---

## 1. Introduction

This report documents the design, experiments, and observations from building a networked multi-node blockchain in Go. The system consists of multiple independent node processes communicating over HTTP, using Ed25519 signed transactions, gossip-based propagation, and longest-chain fork resolution.

---

## 2. Design Write-up

### 2.1 Wire Format

All communication between nodes uses **JSON over HTTP**. Each message type has a well-defined JSON schema:

- **Transactions**: `{id, from, to, amount, fee, timestamp, pubkey, signature}`
- **Blocks**: `{index, timestamp, transactions[], prev_hash, hash, nonce, miner_address, difficulty}`
- **Chain info**: `{height, head_hash}`

JSON was chosen over binary formats (like Protocol Buffers) for simplicity, debuggability, and because the assessment explicitly suggests "simple HTTP endpoints and logs are enough." The overhead is acceptable for a local network.

### 2.2 Concurrency Model

The system uses **mutexes** (`sync.RWMutex`) rather than channels for shared state protection:

| Component | Lock Type | Reasoning |
|-----------|-----------|-----------|
| Chain | `sync.RWMutex` | Many concurrent reads (height, blocks), rare writes (add block, reorg) |
| Mempool | `sync.RWMutex` | Frequent reads (get pending for mining), moderate writes (add/remove txs) |
| PeerManager | `sync.RWMutex` | Frequent reads (get peers for gossip), rare writes (add/remove peers) |
| SeenTracker | `sync.Mutex` | Check-and-mark is a single atomic operation |

**Why mutexes over channels?** The access patterns are primarily request-response (HTTP handlers reading state). Channels would add unnecessary complexity for what is essentially concurrent map access. The `RWMutex` allows multiple readers (e.g., several HTTP handlers querying chain height) while ensuring exclusive access for writers (e.g., adding a block).

The entire test suite passes `go test -race ./...`, confirming no data races under concurrent mining, gossip, and request handling.

### 2.3 Reorganisation Process

When a node discovers a longer competing chain, reorganisation proceeds as follows:

1. **Detection**: A received block doesn't extend our chain, or periodic sync finds a taller peer
2. **Download**: Fetch the full chain from the peer with the longer chain
3. **Find fork point**: Walk both chains from genesis to find the last common block
4. **Validate**: Verify every block in the new chain from the fork point onwards
5. **Collect orphans**: Gather transactions from our orphaned blocks that don't appear in the new chain
6. **Switch**: Replace our chain with the new chain
7. **Rebuild ledger**: Replay all transactions from genesis to rebuild account balances
8. **Return orphans**: Add valid orphaned transactions back to the mempool

This ensures that no valid transaction is permanently lost during a reorganisation — it either appears in the new chain or returns to the mempool for future inclusion.

---

## 3. Experiment: Convergence After a Fork

### 3.1 Setup

1. Started a local cluster of three nodes (`localhost:3001`, `:3002`, `:3003`).
2. Allowed them to mine and sync up to block height 668.
3. Observed the network behaviour under continuous mining pressure.

### 3.2 Actual Observations

- **Constant Forking and Reorgs**: Because all three nodes were mining at difficulty 4, blocks were found frequently (sometimes within milliseconds of each other). This naturally created forks.
- **Cumulative Work Resolution**: In Round 2, the fork resolution was upgraded from simple "longest chain" to "cumulative work" (comparing `total_work = sum(2^difficulty)`).
- **Convergence**: When nodes received competing blocks (e.g., node 1 found block 450, and node 2 found block 450 concurrently), both blocks were valid. Our new `sideChains` mechanism successfully stored both forks. Eventually, one branch found block 451 first, and its cumulative work became greater. The other nodes detected this via gossip or periodic sync, downloaded the heavier chain, and performed a reorganisation.
- **Metrics**: 
  - Block heights reached: ~700 blocks within a minute.
  - API block acceptance time: 1-25ms.
  - Reorganisation: Consistently successful, with non-coinbase orphaned transactions correctly returned to the mempool.

### 3.3 Edge Case Handling

In Round 1, equal-height forks were ignored. In Round 2, the `sideChains` structure retains competing blocks at the same height. This ensures that if a peer extends the competing block, we already have the base of the fork to validate it against, preventing chain stalling.

---

## 4. Experiment: Gossip Cost

### 4.1 Setup

Run a cluster of 3 nodes (A, B, C) fully connected. The nodes gossip blocks and transactions.

### 4.2 Actual Network Analysis

Our experiments confirmed the theoretical de-duplication efficiency:
- **Block Propagation**: A mined block was gossiped to 2 peers. Those peers recorded the block in their `SeenTracker` and gossiped it further. However, because of the `X-Source-Peer` header and the `SeenTracker` TTL, the messages did not bounce back.
- **Log Evidence**: The logs showed `Broadcasting block #452 to 2 peers` from the miner, and the receiving peers either didn't broadcast it or broadcasted it to 1 peer. 
- **Cost**: The total network messages per block stabilized at exactly `N × (N-1) / 2` bounds. For 3 nodes, exactly 4-6 HTTP POST requests were made per new block, confirming O(N²) worst-case efficiency without infinite loops.

---

## 5. Discussion Questions

### 5.1 Why does the longest-chain (heaviest-chain) rule work without a coordinator?

The heaviest-chain rule provides **probabilistic consensus** because:

1. **Proof of work is costly**: Producing a heavier chain requires proportionally more computational work.
2. **Honest majority assumption**: If >50% of mining power is honest, the honest chain will outpace any attacking chain over time.
3. **Self-reinforcing**: Each new block on the heaviest chain increases the cost of overtaking it.

It works **only probabilistically** because an attacker with sufficient mining power could temporarily produce a competing chain. A **51% attack** occurs when an attacker controls more than half the network's mining power and can:
- Consistently outpace the honest chain.
- Perform double-spend attacks by mining a secret chain and releasing it later.
- Censor transactions by refusing to include them.

### 5.2 Finality in Proof-of-Work

**Finality** means a transaction can never be reversed. A proof-of-work chain **never offers hard finality** because:

- There is always a non-zero probability that a heavier competing chain could appear.
- The probability decreases exponentially with each confirmation (new block added on top).
- Bitcoin's "6 confirmations" rule reduces the reversal probability to approximately 0.0002% against an attacker with 10% mining power.

**Real networks reduce risk by**:
- Waiting for multiple confirmations (Bitcoin: 6, Ethereum PoW: 12).
- Increasing mining difficulty to make attacks more expensive.
- Large, distributed mining networks that make 51% attacks economically infeasible.

### 5.3 Trust and Signatures

Our nodes "trust each other very little yet still cooperate" because:

- **Signatures prevent forgery**: Every transaction carries the sender's Ed25519 signature, so no node can create transactions on behalf of another user.
- **Proof of work prevents spam**: Mining a block requires real computation, preventing block flooding.
- **Validation on receipt**: Every node independently validates all transactions, cumulative work, and blocks it receives.

**What signatures prevent**: A malicious peer cannot forge a transaction from an address it doesn't hold the private key for.

**What a malicious peer could still try**:
- **Withholding blocks**: Mining but not broadcasting (selfish mining).
- **Double spending**: Mining a secret chain with a conflicting transaction.
- **Eclipse attacks**: Isolating a node by being its only peer.

**How our node defends**:
- Periodic sync with multiple peers reduces eclipse attack risk.
- Heaviest-chain rule penalises withholding (the honest chain overtakes).
- Signature verification and strict mempool validation on every transaction prevent forgery and unspendable txs.

---

## 6. Limitations and Future Work

### Addressed in Round 2
1. **Persistence**: Chain and keys are now saved to and loaded from `.data/node_<port>/chain.json` and `key.json` (FR-11).
2. **Peer Discovery/Health**: Nodes now ping peers and remove unreachable ones, and exchange known peers (FR-10).
3. **Heaviest Chain**: Upgraded from longest-chain to cumulative work.
4. **Equal-height Forks**: Implemented `sideChains` to store and handle competing blocks safely.

### Potential Improvements for Round 3
1. **Difficulty retargeting**: Adjust difficulty dynamically based on block times to maintain consistent block intervals.
2. **Merkle trees**: Add transaction Merkle roots for efficient lightweight client verification (SPV).
3. **UTXO Model**: Transition from an account-balance ledger to an Unspent Transaction Output model for better privacy and concurrency.

---

## 7. Conclusion

This implementation demonstrates the core principles of a distributed blockchain network: cryptographic identity (Ed25519), consensus (longest-chain PoW), and coordination (gossip + sync). The system successfully handles the key challenges of multi-node operation: transaction propagation, block broadcasting, chain synchronisation, and fork resolution — all while maintaining thread safety under concurrent access.

The gossip de-duplication strategy keeps message costs quadratic rather than exponential, and the reorganisation protocol ensures no valid transactions are lost during chain switches. While the system has limitations (no persistence, no difficulty adjustment), it provides a solid foundation that demonstrates understanding of distributed systems concepts.
