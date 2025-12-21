package stratum

import (
	"github.com/dominant-strategies/go-quai/internal/quaiapi"
)

// Manager manages multiple algorithm-specific stratum servers
type Manager struct {
	servers map[Algorithm]*Server
	stats   *PoolStats
	api     *API
	backend quaiapi.Backend
}

// ManagerConfig holds configuration for the stratum manager
type ManagerConfig struct {
	SHAAddr    string
	ScryptAddr string
	KawPoWAddr string
	APIAddr    string
}

// NewManager creates a new stratum manager with servers for each algorithm
func NewManager(cfg ManagerConfig, backend quaiapi.Backend) *Manager {
	// Create shared stats for all servers
	stats := NewPoolStats()

	m := &Manager{
		servers: make(map[Algorithm]*Server),
		stats:   stats,
		backend: backend,
	}

	// Create algorithm-specific servers sharing the same stats
	if cfg.SHAAddr != "" {
		m.servers[AlgoSHA] = NewServerWithAlgorithm(cfg.SHAAddr, AlgoSHA, backend, stats)
	}
	if cfg.ScryptAddr != "" {
		m.servers[AlgoScrypt] = NewServerWithAlgorithm(cfg.ScryptAddr, AlgoScrypt, backend, stats)
	}
	if cfg.KawPoWAddr != "" {
		m.servers[AlgoKawPoW] = NewServerWithAlgorithm(cfg.KawPoWAddr, AlgoKawPoW, backend, stats)
	}

	// Create API server
	if cfg.APIAddr != "" {
		m.api = NewAPI(cfg.APIAddr, stats, backend)
	}

	return m
}

// Start starts all stratum servers and the API
func (m *Manager) Start() error {
	for algo, server := range m.servers {
		if err := server.Start(); err != nil {
			return err
		}
		server.logger.WithField("algorithm", algo).WithField("addr", server.addr).Info("Stratum server started")
	}

	if m.api != nil {
		if err := m.api.Start(); err != nil {
			return err
		}
	}

	return nil
}

// Stop stops all stratum servers and the API
func (m *Manager) Stop() error {
	var lastErr error

	for _, server := range m.servers {
		if err := server.Stop(); err != nil {
			lastErr = err
		}
	}

	if m.api != nil {
		if err := m.api.Stop(); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// Stats returns the shared pool statistics
func (m *Manager) Stats() *PoolStats {
	return m.stats
}

// GetServer returns the server for a specific algorithm
func (m *Manager) GetServer(algo Algorithm) *Server {
	return m.servers[algo]
}
