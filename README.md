# i-chat-server

A Go chat service with a PostgreSQL-backed chat catalog and WebSocket messaging.

## Requirements

- Go version from `go.mod`
- PostgreSQL 16 (or Docker Compose)
- Docker Compose v2 for the containerized setup

## Run locally with Docker Compose

Copy the example environment file and start the services:

```sh
cp env.example .env
docker compose up --build
```

Compose starts PostgreSQL, Redis, runs database migrations, and starts the HTTP server. The server is exposed on `http://localhost:8181` by default. The example database password is for local development only; replace it anywhere this setup is used beyond a local machine.

Useful commands:

```sh
docker compose logs -f chat-server
docker compose ps
docker compose down
```

`docker compose down` keeps the named database volumes. To remove those volumes and their data, use `docker compose down -v`.

## Run the Go application directly

Start PostgreSQL and apply the migrations first. Goose can be installed with:

```sh
go install github.com/pressly/goose/v3/cmd/goose@v3.14.0
```

Set the PostgreSQL variables from `env.example` for your local database, then run:

```sh
goose -dir ./migrations postgres "host=localhost port=5432 user=postgres password=postgres dbname=db sslmode=disable" up
POSTGRES_HOST=localhost HTTP_HOST_CHAT=127.0.0.1 go run ./cmd/main.go
```

The application listens on `HTTP_HOST_CHAT:HTTP_PORT_CHAT` (defaults to `0.0.0.0:8181`). It exits with an error if it cannot connect to PostgreSQL.

## Configuration

| Variable | Default | Purpose |
| --- | --- | --- |
| `HTTP_HOST_CHAT` | `0.0.0.0` | HTTP bind address |
| `HTTP_PORT_CHAT` | `8181` | HTTP port inside the service |
| `POSTGRES_HOST` | `0.0.0.0` | PostgreSQL host |
| `POSTGRES_PORT` | `5432` | PostgreSQL port |
| `POSTGRES_DB` | `db` | Database name |
| `POSTGRES_USER` | `postgres` | Database user |
| `POSTGRES_PASSWORD` | `postgres` | Database password; set explicitly outside local development |
| `POSTGRES_SSL` | `disable` | PostgreSQL SSL mode |

Compose also reads `HTTP_PORT_CHAT_EXPOSE` (default `8181`) and `PG_PORT` (default `5432`) to choose the host ports. `env.example` includes the Redis and migration settings used by Compose.

## HTTP endpoints

| Method | Path | Description |
| --- | --- | --- |
| `GET` | `/api/v1/chats` | List chats |
| `POST` | `/api/v1/chats` | Create a chat; JSON body: `{"type": 1, "name": "general"}` |
| `GET` | `/healthz` | Liveness check |
| `GET` | `/readyz` | Readiness check; returns `503` if PostgreSQL is unavailable |
| `GET` (WebSocket upgrade) | `/ws` | Connect to the chat WebSocket handler |

Chat creation returns `201 Created` and the created chat as JSON, for example:

```json
{"id": 4611686018427387905, "name": "general"}
```

WebSocket messages are JSON objects with `action`, `sender`, `chat_id`, and `content` fields. Supported actions are `join_chat`, `leave_chat`, `send_text`, and `send_binary`. For example:

```json
{"action":"join_chat","sender":"alice","chat_id":4611686018427387905,"content":""}
```

This service currently does not implement user registration, login, or JWT authentication.

## Development commands

```sh
make build
make test
make lint
```

## CI

GitHub Actions runs linting and tests on pushes and pull requests targeting `main`. On pushes to `main`, it also builds and publishes the server and migration images to GitHub Container Registry (GHCR).

## License

MIT
