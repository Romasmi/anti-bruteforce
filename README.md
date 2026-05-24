# Anti-bruteforce

A microservice that acts as anti-bruteforce middleware in an API gateway. It rate-limits login attempts by login, password, and IP using a leaky bucket algorithm, and supports IP whitelisting/blacklisting.

See [docs/rate-limiting-flow.md](docs/rate-limiting-flow.md) for a sequence diagram of how a request is processed.

## Quick start

```bash
make up      # start Postgres + Redis via Docker Compose
make run     # build and run the service locally
```

The service reads `configs/config.yaml` by default. Pass a different path with `--config`.

## Make commands

| Command | Description |
|---|---|
| `make build` | Compile binary to `./bin/anti-bruteforce` |
| `make run` | Build and run with the default config |
| `make up` | Start dependencies (Postgres, Redis) in Docker |
| `make down` | Stop Docker dependencies |
| `make test` | Run unit tests |
| `make integration-tests` | Run integration tests in Docker |
| `make lint` | Run golangci-lint |
| `make version` | Print build version info |
| `make clean` | Remove build artifacts and test cache |

## CLI

The binary doubles as a management tool. Run any command against a live service:

```bash
# Clear rate limit bucket
./bin/anti-bruteforce clear login <login>
./bin/anti-bruteforce clear ip <ip>

# Blacklist / whitelist (CIDR subnets)
./bin/anti-bruteforce blacklist add 192.168.1.0/24
./bin/anti-bruteforce blacklist remove 192.168.1.0/24

./bin/anti-bruteforce whitelist add 10.0.0.0/8
./bin/anti-bruteforce whitelist remove 10.0.0.0/8
```

By default the CLI connects to `localhost:50051`. Override with `--addr`:

```bash
./bin/anti-bruteforce clear login alice --addr staging.internal:50051
```
