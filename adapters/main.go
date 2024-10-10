package adapters

import (
	"github.com/google/uuid"
	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go/auth"
	"github.com/overmindtech/stdlib-source/adapters/internet"
	"github.com/overmindtech/stdlib-source/adapters/network"
	"github.com/overmindtech/stdlib-source/adapters/test"
	log "github.com/sirupsen/logrus"

	_ "embed"
)

func InitializeEngine(natsOptions auth.NATSOptions, name string, version string, sourceUUID uuid.UUID, heartbeatOptions *discovery.HeartbeatOptions, maxParallel int, reverseDNS bool) (*discovery.Engine, error) {
	e, err := discovery.NewEngine()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Fatal("Error initializing Engine")
	}
	e.Name = "stdlib-source"
	e.NATSOptions = &natsOptions
	e.MaxParallelExecutions = maxParallel
	e.Name = name
	e.UUID = sourceUUID
	e.Version = version
	e.Type = "stdlib"

	if heartbeatOptions != nil {
		heartbeatOptions.HealthCheck = func() error {
			// This can't fail, it's always healthy
			return nil
		}
		e.HeartbeatOptions = heartbeatOptions
	}

	// Add the base adapters
	adapters := []discovery.Adapter{
		&network.CertificateAdapter{},
		&network.DNSAdapter{
			ReverseLookup: reverseDNS,
		},
		&network.HTTPAdapter{},
		&network.IPAdapter{},
		&test.TestDogAdapter{},
		&test.TestGroupAdapter{},
		&test.TestHobbyAdapter{},
		&test.TestLocationAdapter{},
		&test.TestPersonAdapter{},
		&test.TestRegionAdapter{},
	}

	e.AddAdapters(adapters...)

	// Add the "internet" (RDAP) sources
	e.AddAdapters(internet.NewAdapters()...)

	return e, nil
}
