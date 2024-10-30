package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/nats-io/jwt/v2"
	"github.com/nats-io/nkeys"
	"github.com/overmindtech/discovery"
	"github.com/overmindtech/sdp-go"
	"github.com/overmindtech/sdp-go/auth"
	"github.com/overmindtech/sdp-go/sdpconnect"
	"github.com/overmindtech/stdlib-source/adapters"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"golang.org/x/oauth2"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "stdlib-source",
	Short: "Standard library of remotely accessible items",
	Long: `Gets details of items that are globally scoped
(usually) and able to be queried without authentication.
`,
	Run: func(cmd *cobra.Command, args []string) {
		defer func() {
			err := recover()

			if err != nil {
				sentry.CurrentHub().Recover(err)
				defer sentry.Flush(time.Second * 5)
				panic(err)
			}
		}()

		// get engine config
		ec, err := discovery.EngineConfigFromViper("stdlib", ServiceVersion)
		if err != nil {
			log.WithError(err).Fatal("Could not get engine config from viper")
		}
		// Get srcman supplied config
		natsServers := viper.GetStringSlice("nats-servers")
		natsJWT := viper.GetString("nats-jwt")
		natsNKeySeed := viper.GetString("nats-nkey-seed")
		natsConnectionName := viper.GetString("nats-connection-name")

		reverseDNS := viper.GetBool("reverse-dns")

		var natsNKeySeedLog string
		var tokenClient auth.TokenClient

		if natsNKeySeed != "" {
			natsNKeySeedLog = "[REDACTED]"
		}

		log.WithFields(log.Fields{
			"nats-servers":         natsServers,
			"nats-connection-name": natsConnectionName,
			"max-parallel":         ec.MaxParallelExecutions,
			"nats-jwt":             natsJWT,
			"nats-nkey-seed":       natsNKeySeedLog,
			"reverse-dns":          reverseDNS,
			"app":                  ec.App,
			"source-name":          ec.SourceName,
			"source-uuid":          ec.SourceUUID,
		}).Info("Got config")

		// Determine the required Overmind URLs
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		oi, err := sdp.NewOvermindInstance(ctx, ec.App)
		if err != nil {
			log.WithError(err).Fatal("Could not determine Overmind instance URLs")
		}

		// Validate the auth params and create a token client if we are using
		// auth
		var heartbeatOptions *discovery.HeartbeatOptions
		if natsJWT != "" && natsNKeySeed != "" {
			var err error
			tokenClient, err = createTokenClient(natsJWT, natsNKeySeed)
			if err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Fatal("Error creating token client with NATS JWT and NKey seed")
			}
		} else if ec.ApiKey != "" {
			var err error
			tokenClient, err = createAPITokenClient(ec.ApiKey, oi.ApiUrl.String())
			if err != nil {
				log.WithFields(log.Fields{
					"error": err.Error(),
				}).Fatal("Error creating token client with API key")
			}

			tokenSource := auth.NewAPIKeyTokenSource(ec.ApiKey, oi.ApiUrl.String())
			transport := oauth2.Transport{
				Source: tokenSource,
				Base:   http.DefaultTransport,
			}
			authenticatedClient := http.Client{
				Transport: otelhttp.NewTransport(&transport),
			}

			heartbeatOptions = &discovery.HeartbeatOptions{
				ManagementClient: sdpconnect.NewManagementServiceClient(
					&authenticatedClient,
					oi.ApiUrl.String(),
				),
				Frequency: time.Second * 30,
				HealthCheck: func() error {
					// There isn't anything to check here, we just return nil
					return nil
				},
			}
		}

		natsOptions := auth.NATSOptions{
			NumRetries:        -1,
			RetryDelay:        5 * time.Second,
			Servers:           natsServers,
			ConnectionName:    natsConnectionName,
			ConnectionTimeout: (10 * time.Second), // TODO: Make configurable
			MaxReconnects:     -1,
			ReconnectWait:     1 * time.Second,
			ReconnectJitter:   1 * time.Second,
			TokenClient:       tokenClient,
		}

		e, err := adapters.InitializeEngine(
			ec,
			natsOptions,
			heartbeatOptions,
			reverseDNS,
		)
		if err != nil {
			log.WithError(err).Error("Could not initialize aws source")
			return
		}

		// Start HTTP server for status
		healthCheckPort := viper.GetString("service-port")
		healthCheckPath := "/healthz"

		http.HandleFunc(healthCheckPath, func(rw http.ResponseWriter, r *http.Request) {
			if e.IsNATSConnected() {
				fmt.Fprint(rw, "ok")
			} else {
				http.Error(rw, "NATS not connected", http.StatusInternalServerError)
			}
		})

		log.WithFields(log.Fields{
			"port": healthCheckPort,
			"path": healthCheckPath,
		}).Debug("Starting healthcheck server")

		go func() {
			defer sentry.Recover()

			server := &http.Server{
				Addr:    fmt.Sprintf(":%v", healthCheckPort),
				Handler: nil,
				// due to https://github.com/securego/gosec/pull/842
				ReadTimeout:  5 * time.Second, // Set the read timeout to 5 seconds
				WriteTimeout: 5 * time.Second, // Set the write timeout to 5 seconds
			}

			err := server.ListenAndServe()

			log.WithError(err).WithFields(log.Fields{
				"port": healthCheckPort,
				"path": healthCheckPath,
			}).Error("Could not start HTTP server for /healthz health checks")
		}()

		err = e.Start()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Could not start engine")

			os.Exit(1)
		}

		sigs := make(chan os.Signal, 1)

		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		<-sigs

		log.Info("Stopping engine")

		err = e.Stop()

		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Could not stop engine")

			os.Exit(1)
		}

		log.Info("Stopped")

		os.Exit(0)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	var logLevel string

	// add the documentation subcommand
	hostname, err := os.Hostname()
	if err != nil {
		log.WithError(err).Fatal("Could not determine hostname")
	}

	// General config options
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "/etc/srcman/config/source.yaml", "config file path")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log", "info", "Set the log level. Valid values: panic, fatal, error, warn, info, debug, trace")
	cobra.CheckErr(viper.BindEnv("log", "STDLIB_LOG", "LOG")) // fallback to global config
	rootCmd.PersistentFlags().Bool("reverse-dns", false, "If true, will perform reverse DNS lookups on IP addresses")

	// engine config options
	discovery.AddEngineFlags(rootCmd)
	// Config required by all sources in order to connect to NATS. You shouldn't
	// need to change these
	rootCmd.PersistentFlags().StringArray("nats-servers", []string{"nats://localhost:4222", "nats://nats:4222"}, "A list of NATS servers to connect to.")
	cobra.CheckErr(viper.BindEnv("nats-servers", "STDLIB_NATS_SERVERS", "NATS_SERVERS")) // fallback to srcman config
	rootCmd.PersistentFlags().String("nats-connection-name", hostname, "The name that the source should use to connect to NATS")
	cobra.CheckErr(viper.BindEnv("nats-connection-name", "STDLIB_NATS_CONNECTION_NAME", "NATS_CONNECTION_NAME")) // fallback to srcman config
	rootCmd.PersistentFlags().String("nats-jwt", "", "The JWT token that should be used to authenticate to NATS, provided in raw format e.g. eyJ0eXAiOiJKV1Q...")
	cobra.CheckErr(viper.BindEnv("nats-jwt", "STDLIB_NATS_JWT", "NATS_JWT")) // fallback to srcman config
	rootCmd.PersistentFlags().String("nats-nkey-seed", "", "The NKey seed which corresponds to the NATS JWT e.g. SUAFK6QUC...")
	cobra.CheckErr(viper.BindEnv("nats-nkey-seed", "STDLIB_NATS_NKEY_SEED", "NATS_NKEY_SEED")) // fallback to srcman config
	rootCmd.PersistentFlags().String("service-port", "8089", "the port to listen on")
	cobra.CheckErr(viper.BindEnv("service-port", "STDLIB_SERVICE_PORT", "SERVICE_PORT")) // fallback to srcman config
	// tracing
	rootCmd.PersistentFlags().String("honeycomb-api-key", "", "If specified, configures opentelemetry libraries to submit traces to honeycomb")
	cobra.CheckErr(viper.BindEnv("honeycomb-api-key", "STDLIB_HONEYCOMB_API_KEY", "HONEYCOMB_API_KEY")) // fallback to global config
	rootCmd.PersistentFlags().String("sentry-dsn", "", "If specified, configures sentry libraries to capture errors")
	cobra.CheckErr(viper.BindEnv("sentry-dsn", "STDLIB_SENTRY_DSN", "SENTRY_DSN")) // fallback to global config
	rootCmd.PersistentFlags().String("run-mode", "release", "Set the run mode for this service, 'release', 'debug' or 'test'. Defaults to 'release'.")

	// Bind these to viper
	err = viper.BindPFlags(rootCmd.PersistentFlags())
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Could not bind flags to viper")
	}

	// Run this before we do anything to set up the loglevel
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if lvl, err := log.ParseLevel(logLevel); err == nil {
			log.SetLevel(lvl)
		} else {
			log.SetLevel(log.InfoLevel)
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Could not parse log level")
		}

		log.AddHook(TerminationLogHook{})

		// Bind flags that haven't been set to the values from viper of we have them
		cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
			// Bind the flag to viper only if it has a non-empty default
			if f.DefValue != "" || f.Changed {
				err = viper.BindPFlag(f.Name, f)
				if err != nil {
					log.WithFields(log.Fields{
						"error": err,
					}).Fatal("Could not bind flag to viper")
				}
			}
		})

		honeycomb_api_key := viper.GetString("honeycomb-api-key")
		tracingOpts := make([]otlptracehttp.Option, 0)
		if honeycomb_api_key != "" {
			tracingOpts = []otlptracehttp.Option{
				otlptracehttp.WithEndpoint("api.honeycomb.io"),
				otlptracehttp.WithHeaders(map[string]string{"x-honeycomb-team": honeycomb_api_key}),
			}
		}
		if err := initTracing(tracingOpts...); err != nil {
			log.Fatal(err)
		}
	}
	// shut down tracing at the end of the process
	rootCmd.PersistentPostRun = func(cmd *cobra.Command, args []string) {
		shutdownTracing()
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigFile(cfgFile)

	replacer := strings.NewReplacer("-", "_")

	viper.SetEnvKeyReplacer(replacer)
	viper.SetEnvPrefix("STDLIB")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Infof("Using config file: %v", viper.ConfigFileUsed())
	}
}

// createTokenClient Creates a basic token client that will authenticate to NATS
// using the given values
func createTokenClient(natsJWT string, natsNKeySeed string) (auth.TokenClient, error) {
	var kp nkeys.KeyPair
	var err error

	if natsJWT == "" {
		return nil, errors.New("nats-jwt was blank. This is required when using authentication")
	}

	if natsNKeySeed == "" {
		return nil, errors.New("nats-nkey-seed was blank. This is required when using authentication")
	}

	if _, err = jwt.DecodeUserClaims(natsJWT); err != nil {
		return nil, fmt.Errorf("could not parse nats-jwt: %w", err)
	}

	if kp, err = nkeys.FromSeed([]byte(natsNKeySeed)); err != nil {
		return nil, fmt.Errorf("could not parse nats-nkey-seed: %w", err)
	}

	return auth.NewBasicTokenClient(natsJWT, kp), nil
}

func createAPITokenClient(apiKey string, overmindURL string) (auth.TokenClient, error) {
	if apiKey == "" {
		return nil, errors.New("api-key was blank. This is required when using API key authentication")
	}

	return auth.NewAPIKeyClient(overmindURL, apiKey)
}

// TerminationLogHook A hook that logs fatal errors to the termination log
type TerminationLogHook struct{}

func (t TerminationLogHook) Levels() []log.Level {
	return []log.Level{log.FatalLevel}
}

func (t TerminationLogHook) Fire(e *log.Entry) error {
	tLog, err := os.OpenFile("/dev/termination-log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	if err != nil {
		return err
	}

	var message string

	message = e.Message

	for k, v := range e.Data {
		message = fmt.Sprintf("%v %v=%v", message, k, v)
	}

	_, err = tLog.WriteString(message)

	return err
}
