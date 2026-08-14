package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/blockchain2/internal/blockchain"
	"github.com/blockchain2/internal/crypto"
	"github.com/blockchain2/internal/node"
)

func main() {
	// Parse command-line flags
	addr := flag.String("addr", "localhost:3001", "Listen address for the HTTP API")
	peers := flag.String("peers", "", "Comma-separated list of peer addresses")
	minerAddr := flag.String("miner", "", "Miner address for receiving rewards (auto-generated if empty)")
	difficulty := flag.Int("difficulty", 4, "Mining difficulty (number of leading zeros)")
	enableMining := flag.Bool("mine", true, "Enable mining")

	flag.Parse()

	// Parse peer list
	var peerList []string
	if *peers != "" {
		peerList = strings.Split(*peers, ",")
		for i, p := range peerList {
			peerList[i] = strings.TrimSpace(p)
		}
	}

	// Extract port for data isolation
	parts := strings.Split(*addr, ":")
	port := "default"
	if len(parts) == 2 {
		port = parts[1]
	}
	dataDir := ".data/node_" + port

	// Generate miner address if not provided or load from disk (FR-11)
	mAddr := *minerAddr
	if mAddr == "" {
		keyFile := dataDir + "/key.json"
		keyData, err := os.ReadFile(keyFile)
		if err == nil {
			mAddr = string(keyData)
			log.Printf("Loaded miner address from disk: %s", mAddr)
		} else {
			pub, _, err := crypto.GenerateKeyPair()
			if err != nil {
				log.Fatalf("Failed to generate miner key: %v", err)
			}
			mAddr = crypto.AddressFromPublicKey(pub)
			os.MkdirAll(dataDir, 0755)
			os.WriteFile(keyFile, []byte(mAddr), 0644)
			log.Printf("Generated new miner address: %s", mAddr)
		}
	}

	// Validate miner address (must be valid hex and at least 16 chars)
	if len(mAddr) < 16 {
		log.Fatalf("Invalid miner address: must be at least 16 characters")
	}

	// Tie mined difficulty and validated difficulty to one source of truth
	if *difficulty != blockchain.DefaultDifficulty {
		log.Fatalf("Invalid difficulty %d: must match network DefaultDifficulty (%d)", *difficulty, blockchain.DefaultDifficulty)
	}

	// Create node config
	config := node.Config{
		ListenAddr:   *addr,
		Peers:        peerList,
		MinerAddress: mAddr,
		Difficulty:   *difficulty,
		EnableMining: *enableMining,
	}

	// Create and start the node
	n := node.NewNode(config)

	// Set up context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		fmt.Printf("\n[MAIN] Received signal: %v, shutting down...\n", sig)
		cancel()
	}()

	// Start the node (blocks until shutdown)
	if err := n.Start(ctx); err != nil {
		log.Fatalf("Node error: %v", err)
	}

	log.Println("[MAIN] Node stopped. Goodbye!")
}
