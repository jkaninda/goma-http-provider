package services

import (
	"testing"

	"github.com/jkaninda/okapi"
	"github.com/jkaninda/okapi/okapitest"
)

var providerService = &ProviderService{}

func TestHealthCheck(t *testing.T) {
	app := okapi.NewTestServer(t)
	app.Get("/helth", providerService.HealthCheck)

	okapitest.GET(t, "http://localhost:8080/helth").ExpectStatusOK()
}
