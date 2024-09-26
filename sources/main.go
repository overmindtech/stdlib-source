package sources

import (
	"github.com/google/uuid"
	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go/auth"
	"github.com/overmindtech/stdlib-source/sources/internet"
	"github.com/overmindtech/stdlib-source/sources/network"
	"github.com/overmindtech/stdlib-source/sources/test"
	log "github.com/sirupsen/logrus"

	_ "embed"
)

//go:generate sh -c "echo -n $(git describe --tags --exact-match 2>/dev/null || git rev-parse --short HEAD) > commit.txt"
//go:embed commit.txt
var ServiceVersion string

func InitializeEngine(natsOptions auth.NATSOptions, name string, sourceUUID uuid.UUID, heartbeatOptions *discovery.HeartbeatOptions, maxParallel int, reverseDNS bool) (*discovery.Engine, error) {
	e, err := discovery.NewEngine()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Error initializing Engine")
	}
	e.Name = "stdlib-source"
	e.NATSOptions = &natsOptions
	e.MaxParallelExecutions = maxParallel
	e.HeartbeatOptions = heartbeatOptions
	e.HeartbeatOptions.HealthCheck = func() error {
		// This can't fail, it's always healthy
		return nil
	}
	e.Name = name
	e.UUID = sourceUUID
	e.Version = ServiceVersion
	e.Type = "sdtlib"

	// Add the base sources
	sources := []discovery.Source{
		&network.CertificateSource{},
		&network.DNSSource{
			ReverseLookup: reverseDNS,
		},
		&network.HTTPSource{},
		&network.IPSource{},
		&test.TestDogSource{},
		&test.TestGroupSource{},
		&test.TestHobbySource{},
		&test.TestLocationSource{},
		&test.TestPersonSource{},
		&test.TestRegionSource{},
	}

	e.AddSources(sources...)

	// Add the "internet" (RDAP) sources
	e.AddSources(internet.NewSources()...)

	return e, nil
}
