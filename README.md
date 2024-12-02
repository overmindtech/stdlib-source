# stdlib

The source includes a standard library of item types that should always be present. Many of these item types exist in the `global` scope and serve to link together items in other scopes (e.g. ip addresses)

## Config

All configuration options can be provided via the command line or as environment variables:

| Environment Variable | CLI Flag | Automatic | Description |
|----------------------|----------|-----------|-------------|
| `CONFIG`| `--config` | ✅ | Config file location. Can be used instead of the CLI or environment variables if needed |
| `LOG`| `--log` | ✅ | Set the log level. Valid values: panic, fatal, error, warn, info, debug, trace |
| `NATS_NAME_PREFIX`| `--nats-name-prefix` | ✅ | A name label prefix. Sources should append a dot and their hostname .{hostname} to this, then set this is the NATS connection name which will be sent to the server on CONNECT to identify the client |
| `NATS_JWT` | `--nats-jwt` | ✅ | The JWT token that should be used to authenticate to NATS, provided in raw format e.g. `eyJ0eXAiOiJKV1Q{...}` |
| `NATS_NKEY_SEED` | `--nats-nkey-seed` | ✅ | The NKey seed which corresponds to the NATS JWT e.g. `SUAFK6QUC{...}` |
| `MAX_PARALLEL`| `--max-parallel` | ✅ | Max number of requests to run in parallel |

### `srcman` config

When running in srcman, all of the above parameters marked with a checkbox are provided automatically, any additional parameters must be provided under the `config` key. These key-value pairs will become files in the `/etc/srcman/config` directory within the container.

```yaml
apiVersion: srcman.example.com/v0
kind: Source
metadata:
  name: stdlib-source
spec:
  image: ghcr.io/overmindtech/stdlib-source:latest
  replicas: 2
  manager: manager-source
```

### Health Check

The source hosts a health check on `:8089/healthz` which will return an error if NATS is not connected. An example Kubernetes readiness probe is:

```yaml
readinessProbe:
  httpGet:
    path: /healthz
    port: 8080
```

## Development

### Running Locally

The source CLI can be interacted with locally by running:

```shell
go run main.go --help
```

To get automatic recompiles and reloading on code changes, use:

```shell
air
```

### Testing

Tests in this package can be run using:

```shell
go test ./...
```

### Packaging

Docker images can be created manually using `docker build`, but GitHub actions also exist that are able to create, tag and push images. Images will be build for the `main` branch, and also for any commits tagged with a version such as `v1.2.0`
