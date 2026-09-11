# Development

Guide for developing the Cloudflare provider.

## Prerequisites

- Go 1.27+
- Kubernetes cluster
- kubectl configured

## Building

```bash
make build
```

## Testing

```bash
make test
```

## Running Locally

```bash
make run
```

## Code Generation

After modifying API types:

```bash
make generate
```
