# Visitor Analytics & Real-Time URL Tracker

A high-performance, self-contained Go application for tracking link redirects and capturing real-time visitor telemetry, including IP address, geographic location, ISP / ASN, operating system, browser engine, device type, referrer, and timestamps.

Built using the clean architectural patterns from *Let's Go Further* with zero external dependencies and an embedded single-page dashboard.

---

## Features

- **Link Generation & Custom Slugs**: Create trackable URLs with auto-generated or custom path identifiers (`/r/:slug`).
- **Real-Time Telemetry Extraction**:
  - **Network & IP**: Client IP (handles proxies like Cloudflare/Nginx via `X-Forwarded-For` and `X-Real-IP`), ISP Name, Organization, and Autonomous System Number (ASN).
  - **Geolocation**: Country, City, Region / State, and Timezone.
  - **Client Diagnostics**: Operating System (Windows, macOS, iOS, Android, Linux, ChromeOS), Browser (Chrome, Safari, Firefox, Edge, Brave, Opera, curl), and Device Classification (Desktop, Mobile, Tablet, Bot).
  - **Traffic Attribution**: HTTP Referrer URL and Language headers.
- **Aggregated Analytics Dashboard**:
  - Breakdown distributions for OS, Browser, Device types, and Top Countries.
  - Live 5-second polling stream of individual visit logs.
  - One-click link copying and redirect testing.
- **Zero-Dependency Deployment**:
  - HTML, CSS, and JS assets are compiled directly into the binary using Go's `embed.FS`.
  - Statically linked binary (`CGO_ENABLED=0`) requiring no external runtime.
  - Concurrency-safe in-memory data layer (`sync.RWMutex`).

---

## Project Structure

```
visitor_analytics/
├── bin/
│   └── visitor_analytics_linux_amd64  # Pre-compiled static Ubuntu/Linux binary
├── cmd/
│   └── api/
│       ├── errors.go                  # Standardized JSON error responses
│       ├── handlers.go                # Endpoint handlers (creation, redirects, analytics)
│       ├── helpers.go                 # JSON reading/writing utilities
│       ├── main.go                    # Entry point, dependency injection, and CLI flags
│       ├── middleware.go              # CORS, panic recovery, and request logging
│       └── routes.go                  # HTTP routing and embedded static asset serving
├── internal/
│   ├── data/
│   │   └── models.go                  # Concurrency-safe in-memory data store
│   ├── geo/
│   │   └── geo.go                     # IP Geolocation & ASN provider abstraction
│   ├── parser/
│   │   └── parser.go                  # User-Agent classifier (OS, Browser, Device)
│   └── validator/
│       └── validator.go               # Input validation helpers
├── ui/
│   ├── efs.go                         # Go embed.FS filesystem wrapper
│   └── static/
│       ├── css/style.css              # Technical console styling
│       ├── js/app.js                  # Frontend telemetry polling & aggregations
│       └── index.html                 # Single-page dashboard interface
├── go.mod
└── README.md
```

---

## Getting Started

### Prerequisites

- Go 1.22+ (if building from source)

### Running Locally

```bash
# Clone the repository and run
go run ./cmd/api -port 4000
```

Open your browser at [http://localhost:4000](http://localhost:4000).

---

## CLI Flags

| Flag | Default | Description |
| :--- | :--- | :--- |
| `-port` | `4000` | Port for the HTTP server |
| `-base-url` | `http://localhost:<port>` | Public domain or base URL (e.g. `https://track.yourdomain.com`). Used for generating full short links. |
| `-env` | `development` | Environment name (`development`, `staging`, `production`) |

### Example with Custom Domain:
```bash
go run ./cmd/api -port 4000 -base-url "https://analytics.example.com" -env production
```

---

## API Endpoints

### 1. Healthcheck
- **`GET /v1/healthcheck`**
  - Returns service status and environment metadata.

### 2. Create Trackable Link
- **`POST /v1/links`**
  - **Headers**: `Content-Type: application/json`
  - **Body**:
    ```json
    {
      "target_url": "https://github.com",
      "custom_slug": "github-repo"
    }
    ```
  - **Response** (`201 Created`):
    ```json
    {
      "link": {
        "id": "835edc5709cfcf52",
        "slug": "github-repo",
        "target_url": "https://github.com",
        "short_url": "http://localhost:4000/r/github-repo",
        "created_at": "2026-08-29T07:04:44Z",
        "total_visits": 0,
        "visits": []
      }
    }
    ```

### 3. List All Links
- **`GET /v1/links`**
  - Returns a list of all tracked links and their total visit counts.

### 4. Get Link Analytics
- **`GET /v1/links/{slug}`**
  - Returns full metadata, total visits, unique visitor IP count, and individual visit telemetry logs.

### 5. Tracking Redirect Endpoint
- **`GET /r/{slug}`**
  - Inspects incoming request headers (`IP`, `User-Agent`, `Referer`, `Accept-Language`), performs geolocation lookup, records the visit, and returns an `HTTP 302 Found` redirect to the destination URL.

---

## Building & Deploying to Ubuntu / Linux

### 1. Compile Standalone Binary
To compile a lightweight, static Linux binary from source:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o bin/visitor_analytics_linux_amd64 ./cmd/api
```

### 2. Run as a Systemd Service

1. Copy the binary to your server:
   ```bash
   sudo cp bin/visitor_analytics_linux_amd64 /usr/local/bin/visitor_analytics
   sudo chmod +x /usr/local/bin/visitor_analytics
   ```

2. Create `/etc/systemd/system/visitor-analytics.service`:
   ```ini
   [Unit]
   Description=Visitor Analytics Tracking Service
   After=network.target

   [Service]
   Type=simple
   User=www-data
   Group=www-data
   ExecStart=/usr/local/bin/visitor_analytics -port 4000 -base-url "https://yourdomain.com" -env production
   Restart=always
   RestartSec=5s

   [Install]
   WantedBy=multi-user.target
   ```

3. Enable and start:
   ```bash
   sudo systemctl daemon-reload
   sudo systemctl enable --now visitor-analytics
   ```

### 3. Nginx Reverse Proxy Configuration (Optional)

```nginx
server {
    server_name yourdomain.com;

    location / {
        proxy_pass http://127.0.0.1:4000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

## Verification & Quality

To run code analysis and verify standard Go conventions:

```bash
go vet ./...
```
