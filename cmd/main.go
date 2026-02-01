package main

import (
	"github.com/jkaninda/goma-http-provider/internal/config"
	"github.com/jkaninda/goma-http-provider/internal/provider"
	"github.com/jkaninda/goma-http-provider/internal/routes"
	"github.com/jkaninda/logger"
	"github.com/jkaninda/okapi"
	"github.com/jkaninda/okapi/okapicli"
)

func main() {
	app := okapi.New()
	// Create CLI instance
	cli := okapicli.New(app, "Goma").
		String("config", "c", "config.yaml", "Path to configuration file").
		Int("port", "p", 8080, "HTTP server port")
	logger.Info("Starting Goma Gateway HTTP Provider")
	conf, err := config.New(app, cli)
	if err != nil {
		logger.Fatal("Failed to initialize config", "error", err)
	}
	httpProvider, err := provider.NewHTTPProvider(conf.ProviderConf)
	if err != nil {
		logger.Fatal("Failed to initialize HTTPProvider", "error", err)
	}
	route := routes.New(app, httpProvider, conf.Secutity)
	route.RegisterRoutes()

	// Run server
	if err := cli.Run(); err != nil {
		panic(err)
	}
}
