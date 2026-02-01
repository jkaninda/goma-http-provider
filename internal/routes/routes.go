package routes

import (
	"fmt"
	"net/http"

	"github.com/jkaninda/goma-http-provider/internal/config"
	"github.com/jkaninda/goma-http-provider/internal/provider"
	"github.com/jkaninda/goma-http-provider/internal/services"
	"github.com/jkaninda/goma-http-provider/utils"
	"github.com/jkaninda/okapi"
)

var providerService = &services.ProviderService{}

type Route struct {
	app      *okapi.Okapi
	group    *okapi.Group
	metadata map[string]string
	secutity []map[string][]string
}

// NewRoute creates a new Route instance with the provided Okapi app
func New(app *okapi.Okapi, provider *provider.HTTPProvider, secutity []map[string][]string) *Route {
	providerService.Provider = provider

	return &Route{
		app:      app,
		group:    &okapi.Group{Prefix: "api/v1"},
		metadata: provider.GetMetadata(),
		secutity: secutity,
	}
}
func (r *Route) RegisterRoutes() {
	r.app.Get("/", func(ctx *okapi.Context) error {
		return ctx.OK(okapi.M{
			"service": "http-provider",
		})
	})
	r.app.Register(r.providerRoutes()...)

}

// providerRoutes returns the route definitions for the ProviderService
func (r *Route) providerRoutes() []okapi.RouteDefinition {
	cfgGroup := r.group.Group("/config").WithTags([]string{"provider-config"})

	options := []okapi.RouteOption{}
	if len(r.metadata) > 0 {
		for k := range r.metadata {
			meta := fmt.Sprintf("X-Goma-Meta-%s", utils.Capitalize(k))
			options = append(options, okapi.DocHeader(meta, "string", "", true))

		}
	}
	return []okapi.RouteDefinition{
		{
			Method:      http.MethodGet,
			Path:        "/healthz",
			Handler:     providerService.HealthCheck,
			Middlewares: []okapi.Middleware{},
			Summary:     "Service health check",
			Description: "Goma HTTP provider service health check",
		},
		{
			Method:      http.MethodGet,
			Path:        "/stats",
			Handler:     providerService.GetStats,
			Group:       cfgGroup,
			Middlewares: []okapi.Middleware{},
			Response:    &provider.ProviderStats{},
			Summary:     "Get provider statistics",
			Description: "Goma provider statistics",
			Security:    r.secutity,
			Options:     options,
		},
		{
			Method:      http.MethodGet,
			Path:        "/reload",
			Handler:     providerService.ReloadConfig,
			Group:       cfgGroup,
			Middlewares: []okapi.Middleware{},
			Summary:     "Reload configuration",
			Description: "Goma HTTP provider service reload config",
			Security:    r.secutity,
			Options:     options,
		},
		{
			Method:      http.MethodGet,
			Path:        "/",
			Handler:     providerService.GetConfig,
			Group:       cfgGroup,
			Middlewares: []okapi.Middleware{},
			Summary:     "Get provider config",
			Description: "Retrieve Goma gateway config",
			Response:    &config.ConfigBundle{},
			Security:    r.secutity,
			Options:     options,
		},
	}
}
