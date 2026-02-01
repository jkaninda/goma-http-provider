package models

type (
	Route struct {
		// Name provides a descriptive name for the route.
		Name string `yaml:"name" json:"name"`
		// Path specifies the route's path.
		Path string `yaml:"path" json:"path"`
		// Rewrite rewrites the incoming request path to a desired path.
		//
		// For example, `/cart` to `/` rewrites `/cart` to `/`.
		Rewrite string `yaml:"rewrite,omitempty" json:"rewrite,omitempty"`
		// Priority, Determines route matching order
		Priority int `yaml:"priority,omitempty" json:"priority,omitempty"`
		// Enabled specifies whether the route is enabled.
		Enabled bool `yaml:"enabled,omitempty" default:"true" json:"enabled,omitempty"`
		// Hosts lists domains or hosts for request routing.
		Hosts []string `yaml:"hosts,omitempty" json:"hosts,omitempty"`
		// Methods specifies the HTTP methods allowed for this route (e.g., GET, POST).
		Methods []string `yaml:"methods,omitempty" json:"methods,omitempty"`
		// Target defines the primary backend URL for this route.
		Target string `yaml:"target,omitempty" json:"target,omitempty"`
		// Backends specifies a list of backend URLs for load balancing.
		Backends    []Backend   `yaml:"backends,omitempty" json:"backends,omitempty"`
		Maintenance Maintenance `yaml:"maintenance,omitempty" json:"maintenance,omitempty"`
		// TLS contains the TLS configuration for the route, Raw or base64 content
		TLS TlsCertificates `yaml:"tls,omitempty" json:"tls,omitempty"`
		// HealthCheck contains configuration for monitoring the health of backends.
		HealthCheck    RouteHealthCheck `yaml:"healthCheck,omitempty" json:"healthCheck,omitempty"`
		Security       Security         `yaml:"security,omitempty" json:"security,omitempty"`
		DisableMetrics bool             `yaml:"disableMetrics,omitempty" json:"disableMetrics,omitempty"`
		Middlewares    []string         `yaml:"middlewares,omitempty" json:"middlewares,omitempty"`
	}
	Middleware struct {
		// Name specifies the unique name of the middleware.
		Name string `yaml:"name" json:"name"`
		// Type indicates the type of middleware.
		Type string `yaml:"type" json:"type"`
		// Paths lists the routes or paths that this middleware will protect.
		Paths []string `yaml:"paths,omitempty" json:"paths,omitempty"`
		// Rule represents the specific configuration or rules for the middleware.
		// The structure of Rule depends on the middleware Type. For example:
		// - "rateLimit" might use a struct defining rate limits.
		// - "accessPolicy" could use a struct specifying accessPolicy control rules.
		Rule interface{} `yaml:"rule,omitempty" json:"rule,omitempty"`
	}
	TlsCertificates struct {
		Certificates []TLS `yaml:"certificates,omitempty" json:"certificates,omitempty"`
	}
	TLS struct {
		Cert string `yaml:"cert" json:"cert"`
		Key  string `yaml:"key" json:"key"`
	}
)
type RouteHealthCheck struct {
	Path            string `yaml:"path,omitempty" json:"path,omitempty"`
	Interval        string `yaml:"interval,omitempty" json:"interval,omitempty"`
	Timeout         string `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	HealthyStatuses []int  `yaml:"healthyStatuses,omitempty" json:"healthyStatuses,omitempty"`
}
type Maintenance struct {
	Enabled    bool   `yaml:"enabled,omitempty" json:"enabled,omitempty"`
	StatusCode int    `yaml:"statusCode,omitempty" json:"statusCode,omitempty" default:"503"` // default HTTP 503
	Message    string `yaml:"message,omitempty" json:"message,omitempty" default:"Service temporarily unavailable"`
}
type Security struct {
	ForwardHostHeaders      bool        `yaml:"forwardHostHeaders" json:"forwardHostHeaders" default:"true"`
	EnableExploitProtection bool        `yaml:"enableExploitProtection" json:"enableExploitProtection"`
	TLS                     SecurityTLS `yaml:"tls" json:"tls"`
}
type SecurityTLS struct {
	InsecureSkipVerify bool   `yaml:"insecureSkipVerify,omitempty" json:"insecureSkipVerify,omitempty"`
	RootCAs            string `yaml:"rootCAs,omitempty" json:"rootCAs,omitempty"`
	ClientCert         string `yaml:"clientCert,omitempty" json:"clientCert,omitempty"`
	ClientKey          string `yaml:"clientKey,omitempty" json:"clientKey,omitempty"`
}
type Backend struct {
	// Endpoint defines the endpoint of the backend
	Endpoint string `yaml:"endpoint,omitempty" json:"endpoint,omitempty"`
	// Weight defines Weight for weighted algorithm, it optional
	Weight    int  `yaml:"weight,omitempty" json:"weight,omitempty"`
	Exclusive bool `yaml:"exclusive,omitempty" json:"exclusive,omitempty"`
}
