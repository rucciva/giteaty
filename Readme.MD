# Gitea Identity

Expose gitea users' identity through:

- [x] LDAP
- [x] [Caddy V1 Plugin](https://caddyserver.com/v1/)
- [ ] OpenID Connect (with [ORY Hydra](https://github.com/ory/hydra))

## LDAP Server

Look at [docker-compose.yml](build/docker/docker-compose.yml#L78) for refference on how to run the LDAP Server

## Caddy V1 Plugin

To use it with caddy, you need to build caddy yourself and include the plugin, such as:

```go
package main

import (
    "github.com/caddyserver/caddy/caddy/caddymain"
    "github.com/caddyserver/caddy/caddyhttp/httpserver"
    _ "github.com/rucciva/giteaty/cmd/giteaty-caddy"
    // add additional plugins as needed
)

func main() {
    httpserver.RegisterDevDirective("giteaty", "authz")
    caddymain.Run()
}

```

Look at the test files to see [examples](pkg/caddyhandler/setup_test.go#L31-L53).

## Development

### Build and Test

```bash
make generate
make build
make test
```

### Build Docker Image

```bash
cd build/docker
docker-compose \
    -f docker-compose.yml \
    -f docker-compose.production.yml \
    build giteaty
cd ../../
```
