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

1. Start two nodes (A and B) connected to each other
2. Both nodes mine blocks and stay synchronised
3. Disconnect the nodes (remove from each other's peer lists)
4. Mine one block on each node independently, creating a fork
5. Reconnect the nodes
6. Observe which chain wins and what happens to orphaned transactions

### 3.2 Expected Behaviour

- After reconnection, the node with the shorter chain should detect the longer chain via periodic sync
- The shorter-chain node reorganises to adopt the longer chain
- Both nodes converge on the same chain
- Transactions from the orphaned block return to the losing node's mempool

### 3.3 Observations

When running this experiment:

- **Both nodes start with the same genesis block** (deterministic genesis ensures agreement)
- **During disconnection**, each node mines independently, creating two competing chains of equal or near-equal length
- **On reconnection**, the periodic sync (every 15s) detects the height difference
- **The node with the shorter chain downloads the full chain** from the taller peer and performs a reorganisation
- **Convergence time**: Depends on the sync interval (15s by default). In practice, convergence occurs within one sync cycle after reconnection
- **Orphaned transactions**: Non-coinbase transactions from orphaned blocks are verified and returned to the mempool

### 3.4 Edge Case: Equal Height

When both chains are the same height, neither node reorganises (the longest-chain rule requires strictly longer). Convergence only occurs when one node mines the next block, making its chain longer. This is a known limitation of the simple longest-chain rule.

---

## 4. Experiment: Gossip Cost

### 4.1 Setup

Run a cluster of 3 nodes (A, B, C) where:
- A knows B and C
- B knows A and C  
- C knows A and B

Submit one transaction to node A and count the total HTTP requests generated.

### 4.2 Analysis

**Without de-duplication** (naive flooding):
- A receives the transaction → forwards to B and C (2 messages)
- B receives from A → forwards to A and C (2 messages)
- C receives from A → forwards to A and B (2 messages)
- This continues indefinitely: **O(n² × ∞)** messages

**With de-duplication** (our implementation):
- A receives the transaction → marks it as seen → forwards to B and C (2 messages)
- B receives from A → marks as seen → forwards to C only (1 message, A excluded as source)
- C receives from A → marks as seen → forwards to B only (1 message, A excluded as source)
- B receives from C → **already seen, drops it** (0 messages)
- C receives from B → **already seen, drops it** (0 messages)
- **Total: 4 messages** for 3 nodes

**Scaling**: For N fully-connected nodes, the gossip cost with de-duplication is **O(N²)** in the worst case (each node forwards to N-1 peers, but duplicates are dropped). In practice with the `X-Source-Peer` header excluding the sender, the cost is approximately **N × (N-1) / 2** messages, which is manageable for small networks.

The `SeenTracker` with a 5-minute TTL ensures entries are eventually cleaned up, preventing unbounded memory growth.

---

## 5. Discussion Questions

### 5.1 Why does the longest-chain rule work without a coordinator?

The longest-chain rule provides **probabilistic consensus** because:

1. **Proof of work is costly**: Producing a longer chain requires proportionally more computational work
2. **Honest majority assumption**: If >50% of mining power is honest, the honest chain will outpace any attacking chain over time
3. **Self-reinforcing**: Each new block on the longest chain increases the cost of overtaking it

It works **only probabilistically** because an attacker with sufficient mining power could temporarily produce a competing chain. A **51% attack** occurs when an attacker controls more than half the network's mining power and can:
- Consistently outpace the honest chain
- Perform double-spend attacks by mining a secret chain and releasing it later
- Censor transactions by refusing to include them

### 5.2 Finality in Proof-of-Work

**Finality** means a transaction can never be reversed. A proof-of-work chain **never offers hard finality** because:

- There is always a non-zero probability that a longer competing chain could appear
- The probability decreases exponentially with each confirmation (new block added on top)
- Bitcoin's "6 confirmations" rule reduces the reversal probability to approximately 0.0002% against an attacker with 10% mining power

**Real networks reduce risk by**:
- Waiting for multiple confirmations (Bitcoin: 6, Ethereum PoW: 12)
- Increasing mining difficulty to make attacks more expensive
- Large, distributed mining networks that make 51% attacks economically infeasible

### 5.3 Trust and Signatures

Our nodes "trust each other very little yet still cooperate" because:

- **Signatures prevent forgery**: Every transaction carries the sender's Ed25519 signature, so no node can create transactions on behalf of another user
- **Proof of work prevents spam**: Mining a block requires real computation, preventing block flooding
- **Validation on receipt**: Every node independently validates all transactions and blocks it receives

**What signatures prevent**: A malicious peer cannot forge a transaction from an address it doesn't hold the private key for.

**What a malicious peer could still try**:
- **Withholding blocks**: Mining but not broadcasting (selfish mining)
- **Double spending**: Mining a secret chain with a conflicting transaction
- **Eclipse attacks**: Isolating a node by being its only peer

**How our node defends**:
- Periodic sync with multiple peers reduces eclipse attack risk
- Longest-chain rule penalises withholding (the honest chain overtakes)
- Signature verification on every transaction prevents forgery

---

## 6. Limitations and Future Work

### Current Limitations
1. **No persistence**: Chain and keys are lost on restart (FR-11 is a COULD)
2. **Simple peer discovery**: Peers must be specified at startup
3. **Equal-height forks**: Resolution requires waiting for the next block
4. **No difficulty adjustment**: Fixed difficulty across all nodes

### Potential Improvements
1. **Heaviest chain**: Use cumulative proof-of-work instead of simple height comparison
2. **Difficulty retargeting**: Adjust difficulty based on block times to maintain consistent intervals
3. **Merkle trees**: Add transaction Merkle roots for efficient verification
4. **Peer discovery**: Automatic peer exchange to build the network from a seed node
5. **Persistence**: Save chain and keys to disk for crash recovery

---

## 7. Conclusion

This implementation demonstrates the core principles of a distributed blockchain network: cryptographic identity (Ed25519), consensus (longest-chain PoW), and coordination (gossip + sync). The system successfully handles the key challenges of multi-node operation: transaction propagation, block broadcasting, chain synchronisation, and fork resolution — all while maintaining thread safety under concurrent access.

The gossip de-duplication strategy keeps message costs quadratic rather than exponential, and the reorganisation protocol ensures no valid transactions are lost during chain switches. While the system has limitations (no persistence, no difficulty adjustment), it provides a solid foundation that demonstrates understanding of distributed systems concepts.
