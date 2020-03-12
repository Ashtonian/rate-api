# rate-api

Rate-Api allows a user to enter a date time range and get back the rate at which they would be charged to park for that time span built for spot hero.

## Feature Summary

* Get/Set parking rates via `/rates`
* Get parking price via `/rate`
* Get metrics via `/metrics`
* Docker build (see commands below)
* Swagger file located `./docs/swagger.yaml`

## Env Variables

| Name                |    Default     |                                                      Description |
| :------------------ | :------------: | ---------------------------------------------------------------: |
| RATE_API_RATES_PATH | "./rates.json" | The default file location to load the initial rates for the api. |
| RATE_API_PORT       |      3000      |                                    The default port for the api. |

## Common Commands

Run Tests

```bash
go test
```

Run App

```bash
go run ./...
```

or via install

```bash
go install
rate-api
```

Build Docker Image

```bash
docker build --tag "rate-api:v1" .
```

Generate Swagger Docs

```bash
# install swaggo
go get -u github.com/swaggo/swag/cmd/swag

# run to generate docs in ./docs
# remove unused docs.go to avoid go run ./... conflict
swag init && rm ./docs/docs.go
 ```
