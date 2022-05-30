module github.com/overmindtech/stdlib-source

go 1.17

// Direct dependencies
require (
	github.com/nats-io/jwt/v2 v2.2.1-0.20220330180145-442af02fd36a
	github.com/overmindtech/discovery v0.12.6
	github.com/overmindtech/multiconn v0.3.2
	github.com/overmindtech/sdp-go v0.9.2
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.4.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.12.0
)

// Transitive dependencies
require (
	github.com/dgraph-io/dgo/v210 v210.0.0-20220113041351-ba0e5dfc4c3e // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/nats-io/nats.go v1.16.0 // indirect
	github.com/nats-io/nkeys v0.3.0
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/overmindtech/sdpcache v0.3.2 // indirect
	github.com/overmindtech/tokenx-client v0.1.2 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pelletier/go-toml/v2 v2.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/spf13/afero v1.8.2 // indirect
	github.com/spf13/cast v1.5.0 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/subosito/gotenv v1.3.0 // indirect
	golang.org/x/crypto v0.0.0-20220525230936-793ad666bf5e // indirect
	golang.org/x/net v0.0.0-20220526153639-5463443f8c37 // indirect
	golang.org/x/oauth2 v0.0.0-20220524215830-622c5d57e401 // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220527130721-00d5c0f3be58 // indirect
	google.golang.org/grpc v1.46.2 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/ini.v1 v1.66.5 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
