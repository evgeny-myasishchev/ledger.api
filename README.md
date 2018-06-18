# ledger.api
Ledger API layer

# Config

Env vars:

* DB_URL - Postgres db url, defaults to: `host=localhost port=5432 user=postgres dbname=ledger-dev sslmode=disable`
* PORT - Port to listen on, defaults to 3000

# Dev

Repo skeleton taken from (here)[https://github.com/thockin/go-build-template]

### GOPATH and sources

Setup GOPATH dir (for example here ~/projects/go):

Assume direnv is unused. Create .envrc:
```
export GOPATH=${PWD}:${GOPATH}
```
and don't forget `direnv allow .`

Clone this repo to ${GOPATH}/src/ledger.api

### Postgres

Start postgres:

```
docker-compose up -d
```

On a very first run you would also have to create db:

```
docker-compose exec postgres psql -U postgres -c 'CREATE DATABASE "ledger-dev"'
```

Some useful stuff:

```
docker-compose exec postgres pg_dump -U postgres ledger-dev
```

### Dev/Testing

Use [reflex](https://github.com/cespare/reflex) to watch changes and restart server:

```
reflex $(cat .reflex) -- go run cmd/ledger-api/main.go
```

Use `goconvey` to automatically run tests in browser.

Alternatively with the **reflex**:

```
# Run all tests
reflex $(cat .reflex) -- go test ./... -v

# Or specific package
reflex $(cat .reflex) -- go test ./pkg/server/... -v

# Or specific examples
reflex $(cat .reflex) -- go test ./pkg/server/... -v --run TestAuthMiddleware
```

# TODO

Evaluate and integrate https://github.com/spf13/cobra
