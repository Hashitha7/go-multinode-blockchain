package blockchain

// MiningReward is the reward given to the miner for each block.
const MiningReward uint64 = 50

// GenesisBlock creates the deterministic genesis block.
// All nodes produce the same genesis block so they share a common starting point.
func GenesisBlock() *Block {
	genesis := &Block{
		Index:        0,
		Timestamp:    0, // Fixed timestamp for determinism
		Transactions: []*Transaction{},
		PrevHash:     "",
		Nonce:        0,
		MinerAddress: "genesis",
		Difficulty:   1, // Easy difficulty for genesis
	}
	genesis.Hash = genesis.CalculateHash()
	return genesis
}
