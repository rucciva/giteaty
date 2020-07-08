# Gitea Identity

Expose gitea users' identity through:

- [x] LDAP
- [x] [Caddy Plugin](https://caddyserver.com/v1/)
- [ ] OpenID Connect (with [ORY Hydra](https://github.com/ory/hydra))

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
