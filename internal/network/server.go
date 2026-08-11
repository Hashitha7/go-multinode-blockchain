package network

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/blockchain2/internal/blockchain"
	"github.com/blockchain2/internal/mempool"
)

// Server wraps the HTTP server and all its dependencies.
type Server struct {
	httpServer  *http.Server
	handlers    *Handlers
	listenAddr  string
}

// NewServer creates a new HTTP server with all routes registered.
func NewServer(
	listenAddr string,
	chain *blockchain.Chain,
	pool *mempool.Pool,
	peerManager *PeerManager,
	gossiper *Gossiper,
	chainSync *ChainSync,
) *Server {
	handlers := NewHandlers(chain, pool, peerManager, gossiper, chainSync, listenAddr)

	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("/tx", handlers.HandleSubmitTransaction)
	mux.HandleFunc("/block", handlers.HandleReceiveBlock)
	mux.HandleFunc("/chain", handlers.HandleGetChain)
	mux.HandleFunc("/chain/height", handlers.HandleGetChainHeight)
	mux.HandleFunc("/chain/blocks", handlers.HandleGetBlocks)
	mux.HandleFunc("/peers", handlers.HandleGetPeers)
	mux.HandleFunc("/mempool", handlers.HandleGetMempool)
	mux.HandleFunc("/balances", handlers.HandleGetBalances)
	mux.HandleFunc("/status", handlers.HandleGetStatus)

	// Wrap with logging middleware
	loggedMux := loggingMiddleware(mux)

	httpServer := &http.Server{
		Addr:         listenAddr,
		Handler:      loggedMux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer:  httpServer,
		handlers:    handlers,
		listenAddr:  listenAddr,
	}
}

// Start starts the HTTP server (blocking).
func (s *Server) Start() error {
	log.Printf("[SERVER] Listening on %s", s.listenAddr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	log.Printf("[SERVER] Shutting down...")
	return s.httpServer.Shutdown(ctx)
}

// loggingMiddleware logs incoming HTTP requests.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("[HTTP] %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
		log.Printf("[HTTP] %s %s completed in %v", r.Method, r.URL.Path, time.Since(start))
	})
}
