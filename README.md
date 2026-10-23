# Smart-Env

A high-performance, cross-platform CLI that analyzes a project repository, detects missing environment variables, checks system dependencies and background services, and helps spin up the environment.

## Build

```sh
go build -o smart-env ./cmd/smart-env
# or
make build
```

Cross-compile:

```sh
make cross
```

## Usage

```sh
smart-env check                 # scan + check deps/services
smart-env scan                  # alias of check
smart-env init                  # create .env from .env.example
smart-env fix                   # interactively resolve missing/empty env vars
smart-env run                   # spin up environment (services + guidance)
smart-env run --run             # also attempt detected start command
smart-env check -p ./myproject  # specify project root
smart-env check -v              # verbose
```

## Features

- Recursive scan ignoring `.git`, `node_modules`, etc.
- Detect stack: Node, Go, Rust, Python, Java, Ruby, Docker.
- Compare `.env` with `.env.example`/`.env.template`/`.env.sample`.
- Check host tools: node, npm, go, cargo, python, docker, ...
- Parse `docker-compose.yml` and check service status/ports.
- Interactive resolution with secret masking and auto-generation.
- Start services via `docker compose up -d`.
