# Gitea Identity

Expose gitea users' authentication data through:

- [x] LDAP
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
