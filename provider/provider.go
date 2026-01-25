package provider

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jkaninda/goma-http-provider/config"
	"github.com/jkaninda/goma-http-provider/models"
	"github.com/jkaninda/logger"
	"gopkg.in/yaml.v3"
)

type HTTPProvider struct {
	config     *config.ProviderConfig
	client     *http.Client
	cache      map[string]*CachedConfig
	cacheMu    sync.RWMutex
	defaultID  string
	reloadMu   sync.Mutex
	lastReload time.Time
	startTime  time.Time
	metadata   map[string]string
}

type CachedConfig struct {
	Bundle    *config.ConfigBundle
	ExpiresAt time.Time
	ETag      string
}

type ProviderStats struct {
	ConfigsLoaded int       `json:"configsLoaded"`
	LastReload    time.Time `json:"lastReload"`
	Uptime        string    `json:"uptime"`
	CacheHits     int64     `json:"cacheHits"`
	CacheMisses   int64     `json:"cacheMisses"`
}

// NewHTTPProvider creates a new HTTP configuration provider
func NewHTTPProvider(config *config.ProviderConfig) (*HTTPProvider, error) {
	provider := &HTTPProvider{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache:     make(map[string]*CachedConfig),
		startTime: time.Now(),
		metadata:  map[string]string{},
	}

	// Load and cache all configurations at startup
	if err := provider.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize provider: %w", err)
	}

	return provider, nil
}

// initialize loads all configurations and identifies the default
func (p *HTTPProvider) initialize() error {
	p.reloadMu.Lock()
	defer p.reloadMu.Unlock()

	p.cacheMu.Lock()
	p.cache = make(map[string]*CachedConfig)
	p.cacheMu.Unlock()

	seenIDs := map[string]struct{}{}

	for _, cfg := range p.config.Configurations {
		cfg.ID = p.BuildCacheKey(cfg.Metadata)
		if cfg.ID == "" {
			return fmt.Errorf("configuration id is required")
		}
		if _, ok := seenIDs[cfg.ID]; ok {
			return fmt.Errorf("duplicate configuration id: %s", cfg.ID)
		}
		seenIDs[cfg.ID] = struct{}{}

		bundle, err := p.loadConfigFromDirectory(cfg.Directory)
		if err != nil {
			return fmt.Errorf("failed to load config %s: %w", cfg.ID, err)
		}

		// merge metadata
		for k, v := range cfg.Metadata {
			bundle.Metadata[k] = v
			p.metadata[k] = v
		}

		bundle.Checksum = p.calculateChecksum(bundle)
		bundle.Timestamp = time.Now()

		p.cacheMu.Lock()
		p.cache[cfg.ID] = &CachedConfig{
			Bundle:    bundle,
			ExpiresAt: time.Now().Add(5 * time.Minute),
			ETag:      bundle.Checksum,
		}
		p.cacheMu.Unlock()

		if cfg.Default {
			p.defaultID = cfg.ID
		}
	}

	p.lastReload = time.Now()
	return nil
}

// GetConfig retrieves configuration based on metadata filters
func (p *HTTPProvider) GetConfig(
	ctx context.Context,
	metadata map[string]string,
) (*config.ConfigBundle, *config.Configuration, error) {

	cfg := p.matchConfiguration(metadata)
	if cfg == nil {
		logger.Debug("no configuration matched metadata")

		return nil, nil, fmt.Errorf("no configuration matched metadata")
	}

	p.cacheMu.RLock()
	cached := p.cache[cfg.ID]
	p.cacheMu.RUnlock()

	if cached == nil {
		return nil, nil, fmt.Errorf("config %s not loaded", cfg.ID)
	}
	logger.Debug("cached configuration matched metadata")
	return cached.Bundle, cfg, nil
}

func (p *HTTPProvider) loadConfigFromDirectory(directory string) (*config.ConfigBundle, error) {
	bundle := &config.ConfigBundle{
		Version:     "1.0",
		Routes:      make([]models.Route, 0),
		Middlewares: make([]models.Middleware, 0),
		Metadata:    make(map[string]string),
	}

	// Walk through directory
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		// Only process YAML/JSON files
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".yaml" && ext != ".yml" && ext != ".json" {
			return nil
		}

		// Read file
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		// Parse based on file type
		var tempBundle config.ConfigBundle
		if ext == ".json" {
			if err := json.Unmarshal(data, &tempBundle); err != nil {
				return fmt.Errorf("failed to parse JSON %s: %w", path, err)
			}
		} else {
			if err := yaml.Unmarshal(data, &tempBundle); err != nil {
				return fmt.Errorf("failed to parse YAML %s: %w", path, err)
			}
		}

		// Merge into main bundle
		bundle.Routes = append(bundle.Routes, tempBundle.Routes...)
		bundle.Middlewares = append(bundle.Middlewares, tempBundle.Middlewares...)

		// Merge metadata
		for k, v := range tempBundle.Metadata {
			bundle.Metadata[k] = v
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return bundle, nil
}

// ExtractMetadata extracts metadata from request
func (p *HTTPProvider) ExtractMetadata(r *http.Request) map[string]string {
	metadata := make(map[string]string)

	// From query parameters
	for key, values := range r.URL.Query() {
		if len(values) > 0 {
			metadata[key] = values[0]
		}
	}

	// From headers (with X-Goma-Meta- prefix)
	for key, values := range r.Header {
		if strings.HasPrefix(key, "X-Goma-Meta-") && len(values) > 0 {
			metaKey := strings.ToLower(strings.TrimPrefix(key, "X-Goma-Meta-"))
			metadata[metaKey] = values[0]
		}
	}
	return metadata
}

// Authenticate validates the request based on configured auth
func (p *HTTPProvider) Authenticate(
	r *http.Request,
	cfg *config.Configuration,
) error {

	if cfg.Auth == nil {
		return nil
	}
	if cfg.Auth.BasicAuth == nil {
		return nil
	}
	if cfg.Auth.APIKey != "" {
		if r.Header.Get("X-API-Key") == cfg.Auth.APIKey {
			return nil
		}
	}

	if cfg.Auth.BasicAuth.Username != "" {
		u, p, ok := r.BasicAuth()
		if ok &&
			u == cfg.Auth.BasicAuth.Username &&
			p == cfg.Auth.BasicAuth.Password {
			return nil
		}
	}

	return fmt.Errorf("authentication failed for config %s", cfg.ID)
}

func (p *HTTPProvider) GetMetadata() map[string]string {
	if p.cache[p.defaultID] != nil {
		return p.cache[p.defaultID].Bundle.Metadata

	}
	return p.metadata
}
func (p *HTTPProvider) calculateChecksum(bundle *config.ConfigBundle) string {
	temp := *bundle
	temp.Checksum = ""
	temp.Timestamp = time.Time{}

	data, _ := json.Marshal(temp)
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}
func (p *HTTPProvider) matchConfiguration(
	metadata map[string]string,
) *config.Configuration {

	var best *config.Configuration
	bestScore := 0

	for _, cfg := range p.config.Configurations {
		score := 0
		for k, v := range metadata {
			if cfg.Metadata[k] == v {
				logger.Info("matchConfiguration", "key", k, "v", v)
				score++
			}
		}
		if score > bestScore {
			bestScore = score
			best = cfg
		}
	}

	if best != nil {
		return best
	}

	// fallback to default
	if p.defaultID != "" {
		for _, cfg := range p.config.Configurations {
			if cfg.ID == p.defaultID {
				return cfg
			}
		}
	}
	return nil
}

// Reload refreshes all configurations
func (p *HTTPProvider) Reload() error {
	return p.initialize()
}

// getReloadTimestamp returns the last reload timestamp
func (p *HTTPProvider) GetReloadTimestamp() time.Time {
	p.reloadMu.Lock()
	defer p.reloadMu.Unlock()
	return p.lastReload
}

// GetStats returns provider statistics
func (p *HTTPProvider) GetStats() ProviderStats {
	p.cacheMu.RLock()
	configCount := len(p.cache)
	p.cacheMu.RUnlock()

	return ProviderStats{
		ConfigsLoaded: configCount,
		LastReload:    p.GetReloadTimestamp(),
		Uptime:        time.Since(p.startTime).String(),
	}
}

// Close cleanup resources
func (p *HTTPProvider) Close() error {
	p.client.CloseIdleConnections()
	return nil
}

// buildCacheKey creates a consistent cache key from metadata
func (p *HTTPProvider) BuildCacheKey(metadata map[string]string) string {
	if len(metadata) == 0 {
		return "default"
	}

	keys := make([]string, 0, len(metadata))
	for k, v := range metadata {
		keys = append(keys, fmt.Sprintf("%s=%s", k, v))
	}

	return strings.ToLower(strings.Join(keys, "&"))
}
