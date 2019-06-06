# ledger.api
Ledger API layer

# Config

Env vars:

* DB_URL - Postgres db url, defaults to: `postgresql://postgres@localhost:5432/ledger_dev?sslmode=disable`
* PORT - Port to listen on, defaults to 3000
* APP_ENV - Application environment. Defaults to dev. Can be dev, test, stage and prod.
* AUTH0_AUD - auth0 audience, defaults to: https://staging.api.my-ledger.com
* AUTH0_ISS - auth0 issuer, defaults to: https://ledger-staging.eu.auth0.com/

# Dev

Docker and docker-compose assumed to be installed on a dev host.

Logs are written to STDOUT or test.log (when running tests). Log format is json compatible with [pino](github.com/pinojs/pino) so [pino-pretty](https://github.com/pinojs/pino-pretty) can be used to watch the output. Some snippets are below:

```
# Install pino-pretty
nvm use
npm i -g pino-pretty

# Tailf tests
tailf test.log | npx pino-pretty
```

### go

To run it please make sure you have golang installed. 
Simplest is to use [gvm](https://github.com/moovweb/gvm) and [direnv](https://github.com/direnv/direnv).
See [.golang-version](.golang-version) for a version of golang used.

Install required version of golang:

`gvm install $(cat .golang-version)`

### Postgres

Start postgres:

```
docker-compose up -d db
```

On a very first run you would also have to setup a db.
For now db schema is maintained by ledgerv1 app so schema has to be 
initialized using v1 stuff:

```
# Start shell within ledgerv1 env
docker-compose run --rm ledgerv1 bash

# Setup and seed the db
rake db:setup && rake ledger:dummy_seed

# Make sure projections got fully built. For this purpose
# start a backburner worker and wait it to cimplete it's job
# When it's done you should see no new logs
backburner -d && tailf log/development.log

# Setup test db
rake db:test:prepare
```

Optionally use pgadmin to see db structure and run queries:

`docker-compose up -d pgadmin`

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

You can also use a shorthand helper:
```
go-test-watch ./pkg/server/... -v
```

# Docker

Build and push image:

```
make docker_push_release
```

# TODO

* Do not use Convey, use assert and standard approach
