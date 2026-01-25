# Goma HTTP Provider

**Goma HTTP Provider** is an HTTP server that serves **Goma Gateway** configurations from multiple directories based on **authentication** and **request metadata**.

It enables dynamic configuration delivery per environment (production, staging, development, etc.) using headers and optional authentication.

## Features

- Serve gateway configurations from multiple directories
- Environment-based configuration selection using request metadata
- Supports multiple authentication methods:
  - API Key
  - Basic Authentication
  - Header-based metadata

- Optional default configuration (no auth, no metadata required)
- Docker-ready
- OpenAPI / Swagger documentation included

> ⚠️ Even if authentication succeeds, **metadata must match** the configuration definition for access to be granted (unless the configuration is marked as `default`).

## Architecture

```mermaid
graph TB
    subgraph "Control Plane"
        GHP[Goma HTTP Provider]
        subgraph "Configuration Directories"
            PROD_DIR[./data/configs/production]
            STAGE_DIR[./data/configs/staging]
            DEV_DIR[./data/configs/development]
        end
        GHP --> PROD_DIR
        GHP --> STAGE_DIR
        GHP --> DEV_DIR
    end

    subgraph "Data Plane"
        subgraph "Production Environment"
            G1[Goma Gateway G1<br/>Prod<br/>region: eu-central-bg1<br/>ID: goma-prod-01]
            G2[Goma Gateway G2<br/>Prod2<br/>region: eu-central-fsn1<br/>ID: goma-prod-02]
        end
        
        subgraph "Staging Environment"
            G3[Goma Gateway G3<br/>Stage<br/>region: eu-central-fsn1<br/>ID: goma-stage-01]
        end
        
        subgraph "Development Environment"
            G4[Goma Gateway G4<br/>Dev<br/>region: eu-central-fsn1<br/>ID: goma-dev-01]
        end
    end

    G1 -.->|Periodic Sync<br/>BasicAuth| GHP
    G2 -.->|Periodic Sync<br/>BasicAuth| GHP
    G3 -.->|Periodic Sync<br/>BasicAuth| GHP
    G4 -.->|Periodic Sync<br/>ApiKey   | GHP

    GHP -.->|Routes & Middlewares Config| G1
    GHP -.->|Routes & Middlewares Config| G2
    GHP -.->|Routes & Middlewares Config| G3
    GHP -.->|Routes & Middlewares Config| G4

    PROD_DIR -.->|environment: production<br/>region: eu-central-bg1| G1
    PROD_DIR -.->|environment: production<br/>region: eu-central-fsn1| G2
    STAGE_DIR -.->|environment: staging| G3
    DEV_DIR -.->|environment: dev<br/>No Auth Required| G4

    style GHP fill:#e1f5ff
    style G1 fill:#ffebee
    style G2 fill:#ffebee
    style G3 fill:#fff3e0
    style G4 fill:#e8f5e9
    style PROD_DIR fill:#ffcdd2
    style STAGE_DIR fill:#ffe0b2
    style DEV_DIR fill:#c8e6c9
```
---

## Supported Authentication

- **API Key**
- **Basic Auth**
- **Request Metadata Headers**

---

## Links

- **Gateway**: [Goma Gateway on GitHub](https://github.com/jkaninda/goma-gateway)
- **Source Code**: [goma-http-provider](https://github.com/jkaninda/goma-http-provider)
- **Docker Image**: [jkaninda/goma-http-provider](https://hub.docker.com/r/jkaninda/goma-http-provider)

---

## Metadata Handling

Metadata is used to select the appropriate configuration.

If a configuration defines the metadata key `environment`, the incoming HTTP request **must include** the corresponding header:

```
X-Goma-Meta-Environment: production
```

### Metadata Header Rules

- All metadata headers **must be prefixed** with:
  `X-Goma-Meta-`
- Metadata keys are **case-insensitive**
- Metadata **must match exactly** unless the configuration is marked as `default`

---

## API Endpoints

The Goma HTTP Provider exposes the following HTTP endpoints:

| Method | Endpoint                | Description                                                                     |
| ------ | ----------------------- | ------------------------------------------------------------------------------- |
| `GET`  | `/api/v1/config`        | Retrieve the gateway configuration based on request metadata and authentication |
| `POST` | `/api/v1/config/reload` | Reload the configuration for the matching environment (based on metadata)       |
| `GET`  | `/api/v1/config/stats`  | Get statistics for the selected configuration                                   |
| `GET`  | `/healthz`              | Health check endpoint                                                           |

### Metadata-Based Resolution

All `/api/v1/config*` endpoints:

- Require **valid authentication** (unless the configuration is marked as `default`)
- Match configurations using request metadata headers (`X-Goma-Meta-*`)
- Return the configuration associated with the matching environment

---

## Environment Variables

The following environment variables can be used to configure the Goma HTTP Provider:

| Variable        | Description                                           | Default    |
| --------------- | ----------------------------------------------------- | ---------- |
| `PORT`          | Port the HTTP server listens on                       | `8080`     |
| `ENABLE_DOCS`   | Enable or disable the Swagger / OpenAPI documentation | `true`     |
| `TLS_CERT_PATH` | Path to the TLS certificate file (PEM format)         | _disabled_ |
| `TLS_KEY_PATH`  | Path to the TLS private key file (PEM format)         | _disabled_ |

### Server Port

By default, the server runs on port **8080**.
You can override it using the `PORT` environment variable:

```sh
export PORT=9000
```

> CLI flags (e.g. `--port`) typically take precedence over environment variables, if both are set.

### TLS Configuration

To enable HTTPS, both `TLS_CERT_PATH` and `TLS_KEY_PATH` **must be set**:

```sh
export TLS_CERT_PATH=/etc/ssl/certs/server.crt
export TLS_KEY_PATH=/etc/ssl/private/server.key
```

If one of the values is missing, the server falls back to **HTTP**.

### API Documentation

The OpenAPI specification is automatically generated from the application configuration.

Swagger UI is enabled by default. To disable it:

```sh
export ENABLE_DOCS=false
```

## Local Development

```sh
# Clone the repository
git clone https://github.com/jkaninda/goma-http-provider
cd goma-http-provider

# Install dependencies
go mod tidy

# Run the application
go run . --config data/config.yaml
```

### Configuration

- Default port: **8080**
- Override port:

  ```sh
  go run . --config data/config.yaml --port 9000
  ```

### Available Endpoints

- Server: `http://localhost:8080`
- API Docs (Swagger UI):
  `http://localhost:8080/docs`

---

## 🐳 Docker Deployment

```sh
docker run --rm --name goma-http-provider \
  -p 8080:8080 \
  -v ./data:/data \
  jkaninda/goma-http-provider \
  --config /data/config.yaml
```

## Configuration Example

```yaml
configurations:
  - directory: ./data/configs/production
    default: false
    metadata:
      environment: production
      region: eu-central-bg1
      gateway-id: goma-prod-01
    auth:
      basicAuth:
        username: admin
        password: "change me"

  - directory: ./data/configs/staging
    metadata:
      environment: staging
      region: eu-central-fsn1
      gateway-id: goma-stage-01
    auth:
      basicAuth:
        username: admin
        password: staging-pass

  - directory: ./data/configs/development
    default: true
    metadata:
      environment: dev
      region: eu-central-fsn1
      gateway-id: goma-dev-01
    auth: {} # No authentication required
    # auth:
    #   apiKey: dev-secret-key-123
```

### Notes

- When `default: true`:
  - Authentication is **not required**
  - Metadata is **ignored**

- Only **one configuration** should be marked as default

## Goma Gateway HTTP Provider Configuration

```yaml
gateway:
  providers:
    http:
      enabled: true
      endpoint: "https://config.example.com/api/v1/config"
      interval: 60s
      timeout: 10s
      retryAttempts: 5
      retryDelay: 3s
      cacheDir: "" # Defaults to /tmp/goma/cache/config.json
      insecureSkipVerify: false
      headers:
        X-Goma-Meta-Gateway-Id: goma-prod-01
        X-Goma-Meta-Region: eu-central-bg1
        X-Goma-Meta-Environment: production
        Authorization: "${GOMA_PROD_AUTHORIZATION}"
        # X-API-Key: xxxxxx-xxxxx-xxxx
```

---

## License

[MIT](LICENSE) — Free to use, modify, and distribute.

---

## © Copyright

Copyright (c) 2026
**Jonas Kaninda**
