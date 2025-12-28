# SMTP Router

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A smart SMTP proxy that routes outgoing emails to different downstream SMTP servers based on custom logic defined in **Starlark** (Python-like) scripts.

## Quick Start

1.  **Clone and Run**
    ```bash
    git clone https://github.com/goyal-aman/mailmux.git
    cd mailmux
    go run .
    ```

2.  **Configure**
    Edit `config.yaml` to define your users and downstream servers.

3.  **Scripting**
    Edit `round_robin.star` to define your routing logic.

## Docker

Run with Docker:

```bash
docker build -t smtp-router .
docker run -p 1020:1020 -v $(pwd)/config.yaml:/app/config.yaml -v $(pwd)/round_robin.star:/app/round_robin.star smtpmux
```

## Development

Run tests:
```bash
go test ./...
```

## Examples

### Downstream 1 (Port 1026)
```bash
docker run --rm -d -p 1027:1025 -p 8027:8025 --name mailserve1 mailhog/mailhog
```

### Downstream 2 (Port 1027)
```bash
docker run --rm -d -p 1026:1025 -p 8026:8025 --name mailserve2 mailhog/mailhog
```

## Send Email Locally
```
swaks --to recipient@example.com \
      --from test@user.com \
      --server localhost:1022 \
      -a PLAIN \
      --auth-user test@user.com \
      --auth-password password123 \
      --body "maaaa"
```

## Dynamic Selector (Starlark)
TODO: update for go plugins
You can define your own routing logic in a Starlark script (Python-like syntax).

Example `round_robin.star`:

```python
def selector(downstreams):
    for ds in downstreams:
        err = send(ds=ds)
        if err == None:
            return None
    return "all failed"
```