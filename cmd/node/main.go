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

	// Generate miner address if not provided
	mAddr := *minerAddr
	if mAddr == "" {
		pub, _, err := crypto.GenerateKeyPair()
		if err != nil {
			log.Fatalf("Failed to generate miner key: %v", err)
		}
		mAddr = crypto.AddressFromPublicKey(pub)
		log.Printf("Generated miner address: %s", mAddr)
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
