# slashbot_sample

A simple boilerplate for Slack's slash command handling bot

## usage

### run directly

- socket mode
  - `go run ./cmd/socket`
- webhook mode
  - `go run ./cmd/webhook`

### run on docker container

- socket mode
  - `docker build -t socket -f ./Dockerfile.socket .`
  - `docker run --rm --env-file ./.env socket`
- webhook mode
  - `docker build -t webhook -f ./Dockerfile.webhook .`
  - `docker run --rm --env-file ./.env -p 38080:8080  webhook`

## license

MIT

## author

walkure
