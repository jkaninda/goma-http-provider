package services

import (
	"net/http"

	"github.com/jkaninda/goma-http-provider/config"
	"github.com/jkaninda/goma-http-provider/provider"
	"github.com/jkaninda/okapi"
)

type ProviderService struct {
	Provider *provider.HTTPProvider
}

func (p *ProviderService) HealthCheck(c okapi.C) error {
	return c.OK(okapi.M{
		"status":  "healthy",
		"service": "goma-gateway-http-provider",
	})
}
func (p *ProviderService) GetStats(c okapi.C) error {
	_, cfg, err := p.configBundle(c)
	if err != nil {
		return c.AbortNotFound("Config not found", err)
	}

	if err := p.Provider.Authenticate(c.Request(), cfg); err != nil {
		return c.AbortUnauthorized("Unauthorized", err)
	}
	return c.OK(p.Provider.GetStats())
}
func (p *ProviderService) ReloadConfig(c okapi.C) error {
	_, cfg, err := p.configBundle(c)
	if err != nil {
		return c.AbortNotFound("Config not found", err)
	}

	if err := p.Provider.Authenticate(c.Request(), cfg); err != nil {
		return c.AbortUnauthorized("Unauthorized", err)
	}

	if err := p.Provider.Reload(); err != nil {
		return c.AbortInternalServerError("Reload failed", err)
	}
	return c.OK(okapi.M{
		"status":    "reloaded",
		"timestamp": p.Provider.GetReloadTimestamp(),
	})
}

func (p *ProviderService) GetConfig(c okapi.C) error {

	bundle, cfg, err := p.configBundle(c)
	if err != nil {
		return c.AbortNotFound("Config not found", err)
	}
	if err := p.Provider.Authenticate(c.Request(), cfg); err != nil {
		return c.AbortUnauthorized("Unauthorized", err)
	}

	c.SetHeader("ETag", bundle.Checksum)
	if c.Header("If-None-Match") == bundle.Checksum {
		return c.AbortWithStatus(http.StatusNotModified, "No change")
	}

	return c.OK(bundle)
}
func (p *ProviderService) configBundle(c okapi.C) (*config.ConfigBundle, *config.Configuration, error) {
	metadata := p.Provider.ExtractMetadata(c.Request())

	bundle, cfg, err := p.Provider.GetConfig(c.Request().Context(), metadata)
	if err != nil {
		return nil, nil, err
	}
	return bundle, cfg, nil
}
