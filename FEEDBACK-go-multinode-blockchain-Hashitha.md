# Feedback on Your Networked Blockchain Submission — Hashitha

**Reviewer:** Dasun · **Date:** 2026-08-14

This is one of the strongest submissions I have tested this round, and I want to be
specific about why before I get to what to fix. The thing this whole round exists to
test — does the network actually converge after a real fork, with the ledger rebuilt
correctly — your code does, and it does it cleanly. **Verified:** I funded two
independent senders on a shared chain, partitioned the network, submitted a different
payment to each side (Alice → recipient 50, Bob → recipient 70), let each side mine
its own block so the chains genuinely diverged (heads `0000a7c11a14` at height 16 vs
`0000c169f117` at height 18), then reconnected them. The network converged on one head
in **0.5 seconds**, and afterwards the recipient's balance read **120 on both nodes** —
the 70 from the winning chain plus the orphaned 50, which was returned to the pool and
re-mined. That is fork resolution, ledger rebuild, and orphan requeue all working
together, and it is the hardest part of the assignment. I also watched three nodes mine
concurrently at the same difficulty for 35 seconds, forking constantly, and reconverge
on their own to a single head at height 226 with a consistent ledger — no restart, no
nudge. `go test -race ./...` passes in a real race detector with no data race, and your
signature enforcement is real rather than decorative: both an amount-tampered and a
sender-spoofed transaction were rejected.

**The one thing holding this back from a clean sweep is a consensus-soundness bug: your
heaviest-chain rule trusts a number the sender controls.** It does not break anything
between honest nodes, which is why everything above works — but a malicious peer can
exploit it, and this round is partly about exactly that threat model.

## What works (verified)

- **Fork convergence + ledger rebuild + orphan requeue (FR-6).** Described above.
  `ReorganiseTo` ([chain.go:125-194](go-multinode-blockchain/internal/blockchain/chain.go#L125-L194))
  finds the fork point, validates the new suffix, collects orphaned non-coinbase
  transactions, swaps the chain, and rebuilds the ledger from scratch with
  `RebuildFromChain` ([ledger.go:43-57](go-multinode-blockchain/internal/blockchain/ledger.go#L43-L57)).
  That is a full replay, not an in-place patch — the correct approach, and the reason your
  balances stayed consistent across every reorg I forced.
- **Automatic convergence with no manual intervention.** 3-node fork storm settled to one
  head `00008d9b6fcc` at height 226; controlled partition healed in 0.5s. Convergence is
  driven by gossip on block receipt plus a 15s periodic sync
  ([node.go:194-227](go-multinode-blockchain/internal/node/node.go#L194-L227)).
- **Race-clean under concurrent mining and gossip (FR-7).** **Verified:** `go test -race
  ./...` PASS in `golang:1.26`, no `WARNING: DATA RACE`. Shared state is guarded with
  per-component `sync.RWMutex` and gossip fans out in per-peer goroutines, so no lock is
  ever held across a network call.
- **Signatures bind to contents and are enforced everywhere (FR-2).** **Verified:** an
  amount-tampered transaction was rejected `400 Invalid transaction: invalid signature`;
  a spoofed sender (`from` changed to an address that doesn't match the signing key) was
  rejected `400 sender address does not match public key: got abc123…, want 271f…`. The
  address is re-derived from the public key during verification
  ([transaction.go:95-126](go-multinode-blockchain/internal/blockchain/transaction.go#L95-L126)),
  and `Verify()` is reached on all four entry paths: client submit, peer gossip (same
  `POST /tx` handler), block validation (`block.Validate` verifies every tx), and sync
  (downloaded blocks validated from the fork point before adoption).
- **Gossip propagates and de-duplicates (FR-3, FR-4).** Tx submitted to one node reached
  all three pools; re-submitting the same transaction 10× grew no pool. There is a real
  seen-set with a TTL ([gossip.go:15-72](go-multinode-blockchain/internal/network/gossip.go#L15-L72)),
  and gossip excludes the peer it came from via the `X-Source-Peer` header.
- **Fresh-node sync (FR-5) and a working cluster launcher.** A late-joining node reached
  the peer's height validating each block. **Verified:** `bash scripts/cluster.sh` built and
  started three connected nodes on `:3001-:3003`, each reporting 2 peers.

## Must fix

- **Your heaviest-chain rule trusts an attacker-controlled difficulty (consensus
  soundness).** `HasValidPoW` hardcodes the work check to `DefaultDifficulty` (4 leading
  zeros) and ignores the block's own `Difficulty` field
  ([block.go:87-90](go-multinode-blockchain/internal/blockchain/block.go#L87-L90)), while
  `calculateWork` sums `1 << b.Difficulty` from that same unvalidated field
  ([chain.go:89-97](go-multinode-blockchain/internal/blockchain/chain.go#L89-L97)). Nothing
  checks that `b.Difficulty` equals the network difficulty. **The consequence:** a
  malicious peer can mine a block that only needs 4 leading zeros (cheap) but declares
  `Difficulty: 60`, and your `calculateWork` will value it at `2^60`. It becomes the
  "heaviest" chain and every honest node reorgs onto it. This is the concrete answer to
  your own report's "what could a malicious peer still try" question, and right now the
  defence is missing. Fix: in `Validate`, reject any block whose `Difficulty` is not the
  expected network difficulty, and compute work from the validated value.
- **A related smaller cliff: difficulty below 4 silently breaks block production.** Because
  `HasValidPoW` is hardcoded to 4 while the miner mines to `b.Difficulty`, running with
  `--difficulty 3` produces blocks the node then rejects as invalid PoW — the node mines
  and cannot add its own blocks. It happens to work only because your cluster script uses
  4. Tie the mined difficulty and the validated difficulty to one source of truth.

## Smaller issues

- **Unbounded resource reads — no `http.MaxBytesReader` / `io.LimitReader`.** Handlers read
  request bodies with `io.ReadAll` and no cap
  ([handlers.go:54-59](go-multinode-blockchain/internal/network/handlers.go#L54-L59)); the
  sync client reads peer responses with no limit. A hostile peer can stream forever. The
  `sideChains` map ([chain.go:12](go-multinode-blockchain/internal/blockchain/chain.go#L12))
  and the peer set (`ExchangePeers` adds any address a peer sends) are also unbounded. Your
  node survived a 3 MB junk body in testing, so nothing is on fire — but the bounds the
  spec asks for aren't there. Add a request-size limit and cap the orphan/peer structures.
- **The seen-set is marked before the signature is verified.** `HandleSubmitTransaction`
  calls `IsSeenTx(tx.ID)` ([handlers.go:68](go-multinode-blockchain/internal/network/handlers.go#L68)),
  which *records* the id, before `tx.Verify()` at line 78. An invalid transaction's id is
  remembered as "seen" for the TTL. You are protected from the classic censorship version
  of this because your `tx.ID` covers the signature, so an attacker can't forge a different
  transaction with a victim's id — but the ordering is still verify-then-cache in intent,
  and it's cleaner to verify first.
- **No test mines while gossiping concurrently.** You have real concurrency in the suite —
  `TestPoolConcurrentAccess` runs 100 goroutines, `TestMultiNodeIntegration` drives two live
  HTTP nodes — so the race detector had something to find. But the multi-node test mines and
  *then* syncs sequentially; nothing drives mining and gossip at the same time, which is the
  scenario FR-7 is really about. Add a test that starts two nodes, has one mine on a timer
  while the other floods transactions, and runs it under `-race`.
- **Formatting: `gofmt -l` flags 21 files.** Twenty are flagged only for CRLF line endings;
  one (`node.go`) has trailing whitespace on three blank lines. Run `gofmt -w .` and add a
  `.gitattributes` with `*.go text eol=lf`. Minor, but `gofmt`-clean is an explicit NFR.
- **`ReturnOrphaned` re-checks the signature but not ledger validity.** Orphaned
  transactions come back to the pool on signature alone
  ([pool.go:89-107](go-multinode-blockchain/internal/mempool/pool.go#L89-L107)).
  A now-invalid double-spend can sit in the pool until mining filters it out via
  `FilterValidTransactions`. It never gets mined, so money is safe — but it's tidier to drop
  it at requeue time.

## What to fix first

1. **Validate `block.Difficulty` against the network difficulty and compute work from the
   validated value.** This closes the forgeable-heaviest-chain hole and also removes the
   difficulty-below-4 foot-gun. It is the single most important fix and it is a small one.
2. **Add a request-size limit** (`http.MaxBytesReader`) and bound the orphan and peer
   structures, so a hostile peer can't exhaust memory.
3. **Add a concurrent mine-while-gossip `-race` test**, then run `gofmt -w .` and commit the
   line-ending fix.

To place this honestly: operationally, this is excellent work. Every MUST requirement is
verified working, the network converges automatically after a real fork with a correct
ledger rebuild and orphan requeue, and it is race-clean — several things a lot of
submissions cannot claim. What separates it from a flawless one is that the consensus rule,
the thing this round is built around, trusts a number the sender controls, and the input
handling isn't hardened against a peer that isn't playing nicely. Those are narrow,
localised fixes, not a redesign. Fix the difficulty validation and this is at the very top
of the batch.

## Notes for the walkthrough call

1. Walk me through a live reorg on your own cluster: partition two nodes, pay a recipient on
   each side, heal, and show me the recipient ending with both payments. (I've verified this
   works — I want to hear you explain *why* the orphaned transaction comes back and where.)
2. In `calculateWork` you sum `1 << b.Difficulty`. Where is `b.Difficulty` validated? What
   stops me, as a peer, from sending you a block that claims `Difficulty: 60`? (This is the
   Must-fix; see how quickly they spot it.)
3. `HasValidPoW` checks `DefaultDifficulty`, but the miner mines to `b.Difficulty`. What
   happens if I start a node with `--difficulty 3`? (They should realise the node can't add
   its own blocks.)
4. Your `-race` suite is green. Which test would have caught a race between the miner
   appending a block and a handler reading the chain? Show me the goroutines. (Probes whether
   the green run actually exercised concurrent mining + gossip.)
5. You verify the signature after checking the seen-set. Talk me through the order of
   operations in `HandleSubmitTransaction` and whether it matters that `tx.ID` includes the
   signature.
