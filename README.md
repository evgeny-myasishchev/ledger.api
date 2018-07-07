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

Repo skeleton taken from (here)[https://github.com/thockin/go-build-template].
Docker and docker-compose assumed to be installed on a dev host.

### go

Requires go **1.10**

Assume gvm and direnv is unused.

Setup GOPATH dir (for example here ~/projects/go):

Create .envrc:
```
. $GVM_ROOT/scripts/gvm-default
gvm use 1.10
export GOPATH=${PWD}:${GOPATH}
export PATH=${PWD}/bin:${PATH}
```
and don't forget `direnv allow .`

Clone this repo to ${GOPATH}/src/ledger.api

### Postgres

Start postgres:

```
docker-compose up db -d
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

# TODO

Evaluate and integrate https://github.com/spf13/cobra