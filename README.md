# KVolt Test API ⚡

This is a comprehensive reference implementation and test suite for the **KVolt Framework**.

## Features Showcase
*   **Routing**: Parametrized routes (`/users/:id`), Groups (`/v1`), and Wildcards.
*   **Middleware**: Global and per-route middleware (Logging, Recovery).
*   **Performance**: High-speed JSON serialization via `bytedance/sonic`.
*   **Static Files**: Serves assets from `./public` (HTML/CSS/JS).
*   **Documentation**: Auto-generated Swagger UI integration.

## Getting Started

### Prerequisites
*   Go 1.21+
*   KVolt Framework (local or vendored)

### Running the Server
```bash
go run cmd/api/main.go
```
Server will start at `http://localhost:8080`.

### API Endpoints
| Method | Path | Description |
| :--- | :--- | :--- |
| `GET` | `/` | Hello World JSON |
| `GET` | `/static/` | Static Website (HTML/CSS) |
| `GET` | `/users` | List Users |
| `POST` | `/posts` | Create Post (JSON Echo) |
| `GET` | `/swagger/index.html` | Interactive API Docs |

## Performance Benchmarks 🚀
KVolt is designed for extreme throughput. Below is a load test result on a standard development machine:

**Command:**
```bash
wrk -t12 -c4000 -d1s http://127.0.0.1:8080/
```

**Results:**
```text
Running 1s test @ http://127.0.0.1:8080/
  12 threads and 4000 connections
  Thread Stats   Avg      Stdev     Max   +/- Stdev
    Latency     4.07ms    7.12ms  82.00ms   90.04%
    Req/Sec    21.77k     7.03k   39.29k    62.50%
  260299 requests in 1.01s, 34.75MB read
  Socket errors: connect 2987, read 0, write 0, timeout 0
Requests/sec: 258237.49
Transfer/sec:     34.48MB
```

## Testing
Run the verification suite to validate framework behavior:
```bash
go test ./cmd/verification -v
```

### Verification Suite Coverage
| Test Case | Feature Verified |
| :--- | :--- |
| `TestStaticRoutes` | Basic GET/POST routing |
| `TestParamRoutes` | URL parameters (`/:id`) |
| `TestWildcardRoutes` | Wildcard paths (`/*filepath`) |
| `TestGroupRoutes` | Route grouping and prefixes |
| `TestConcurrency` | Thread-safety under load |
| `Test404` | Custom Not Found handling |
| `TestPostEcho` | JSON serialization (Sonic) |
| `TestSwaggerUI` | API Documentation serving |
| `TestStaticFiles` | File server & MIME types |
| `TestWebSocket` | Real-time communication |
| `TestHTTP2` | HTTP/2 Protocol support |

## Repository
**GitHub**: [kaviraj-m/kvolt-test](https://github.com/kaviraj-m/kvolt-test)
