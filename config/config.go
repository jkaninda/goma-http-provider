package config

import (
	"fmt"
	"os"
	"time"

	goutils "github.com/jkaninda/go-utils"
	"github.com/jkaninda/goma-http-provider/models"
	"github.com/jkaninda/goma-http-provider/utils"
	"github.com/jkaninda/logger"
	"github.com/jkaninda/okapi"
	"github.com/jkaninda/okapi/okapicli"
	"github.com/joho/godotenv"
)

type Config struct {
	app           *okapi.Okapi
	path          string
	ProviderConf  *ProviderConfig
	server        ServerConfig
	hasBasicAuth  bool
	hasApiKeyAuth bool
	Secutity      []map[string][]string
}
type ServerConfig struct {
	port       int
	enableDocs bool
	tls        Tls
}
type Tls struct {
	Cert string
	Key  string
}
type (
	ProviderConfig struct {
		Version        string           `json:"version" yaml:"version"`
		Configurations []*Configuration `yaml:"configurations"`
	}
	Configuration struct {
		ID string `yaml:"id"`

		Directory string    `yaml:"directory"`
		Auth      *HTTPAuth `yaml:"auth,omitempty" json:"auth,omitempty"`
		// If the config in this path is default
		Default  bool              `yaml:"default"`
		Metadata map[string]string `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	}
	ConfigBundle struct {
		Version     string              `json:"version" yaml:"version"`
		Routes      []models.Route      `json:"routes" yaml:"routes"`
		Middlewares []models.Middleware `json:"middlewares" yaml:"middlewares"`
		Metadata    map[string]string   `json:"metadata,omitempty" yaml:"metadata,omitempty"`
		Checksum    string              `json:"checksum,omitempty" yaml:"checksum,omitempty"`
		Timestamp   time.Time           `json:"timestamp" yaml:"timestamp"`
	}

	HTTPAuth struct {
		APIKey    string     `yaml:"apiKey,omitempty"`
		BasicAuth *BasicAuth `yaml:"basicAuth,omitempty" `
	}
	BasicAuth struct {
		Username string `yaml:"username,omitempty" json:"username,omitempty"`
		Password string `yaml:"password,omitempty" json:"password,omitempty"`
	}
)

func (c *Config) validate() error {
	if len(c.ProviderConf.Configurations) == 0 {
		return fmt.Errorf("at least one configuration is required")
	}

	defaultCount := 0
	for i, cfg := range c.ProviderConf.Configurations {
		if cfg.Directory == "" {
			return fmt.Errorf("configuration[%d]: directory is required", i)
		}
		if len(cfg.Metadata) == 0 {
			logger.Warn("Empty metadata", "config", i)
		}
		// Check if directory exists
		if _, err := os.Stat(cfg.Directory); os.IsNotExist(err) {
			return fmt.Errorf("configuration[%d]: directory does not exist: %s", i, cfg.Directory)
		}
		if cfg.Auth != nil {
			if cfg.Auth.APIKey != "" {
				c.hasApiKeyAuth = true
			}
			if cfg.Auth.BasicAuth != nil {
				if cfg.Auth.BasicAuth.Username == "" || cfg.Auth.BasicAuth.Password == "" {
					return fmt.Errorf("error, basic auth, username or password missing")
				}
				c.hasBasicAuth = true
			}

		}

		if cfg.Default {
			defaultCount++
		}
	}

	if defaultCount > 1 {
		return fmt.Errorf("only one configuration can be marked as default")
	}

	return nil
}
func New(app *okapi.Okapi, cli *okapicli.CLI) (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()
	// Parse flags
	if err := cli.ParseFlags(); err != nil {
		return nil, err
	}
	configFile := cli.GetString("config")
	port := cli.GetInt("port")
	cfg := &Config{

		app:  app,
		path: configFile,
		server: ServerConfig{
			enableDocs: goutils.EnvBool("ENABLE_DOCS", true),
			port:       goutils.EnvInt("PORT", port),
			tls: Tls{
				Cert: goutils.Env("TLS_CERT_PATH", ""),
				Key:  goutils.Env("TLS_KEY_PATH", ""),
			},
		},
		Secutity:     []map[string][]string{},
		ProviderConf: &ProviderConfig{},
	}
	err := cli.LoadConfig(cfg.path, cfg.ProviderConf)
	if err != nil {
		return cfg, fmt.Errorf("failed to load provider config file, error=%v", err)
	}
	if err := cfg.initialize(); err != nil {
		return nil, err
	}
	cfg.enableDocs()
	return cfg, nil
}
func (c *Config) initialize() error {

	// Init TLS
	if len(c.server.tls.Cert) > 0 && len(c.server.tls.Key) > 0 {
		tls, err := okapi.LoadTLSConfig(c.server.tls.Cert, c.server.tls.Key, "", false)
		if err != nil {
			return fmt.Errorf("failed to load tls, error=%v", err)
		} else {
			// Add tls server
			c.app.With(okapi.WithTLS(tls))
			if c.server.port == 8080 {
				c.server.port = 8443
			}
		}
	}
	addr := fmt.Sprintf(":%d", c.server.port)
	c.app.With(okapi.WithAddr(addr))

	if err := c.validate(); err != nil {
		return err
	}
	return nil

}
func (c *Config) enableDocs() {
	securitySchemes := okapi.SecuritySchemes{}
	if c.server.enableDocs {
		if c.hasBasicAuth {
			securitySchemes = append(securitySchemes, okapi.SecurityScheme{
				Name:   "basicAuth",
				Type:   "http",
				Scheme: "basic",
			})
			c.Secutity = append(c.Secutity, map[string][]string{
				"basicAuth": {}})

		}
		if c.hasApiKeyAuth {

			securitySchemes = append(securitySchemes, okapi.SecurityScheme{
				Name: "X-API-Key",
				Type: "apiKey",
				In:   "header",
			})
			c.Secutity = append(c.Secutity, map[string][]string{"X-API-Key": {}})

		}
		c.app.WithOpenAPIDocs(okapi.OpenAPI{
			Title:   "Goma Gateway HTTP Provider",
			Version: utils.Version,
			License: okapi.License{
				Name: "MIT",
			},
			Contact: okapi.Contact{
				Name:  "Jonas Kaninda",
				Email: "me@jkaninda.dev",
				URL:   "https://github.com/jkaninda/goma-http-provider"},
			SecuritySchemes: securitySchemes,
		})
	}

}
