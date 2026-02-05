# Epic 03: Performance Benchmarks

## Overview

Create a comprehensive benchmark suite comparing modkit against NestJS and Go DI alternatives. The benchmarks demonstrate that modkit delivers NestJS-style architecture with Go's performance characteristics.

**Goals:**
1. Quantify modkit's performance vs NestJS (cross-language)
2. Compare modkit vs Go alternatives (wire, fx, do)
3. Provide reproducible, automated benchmarks
4. Generate clear visualizations for documentation

**Repository:** `github.com/go-modkit/benchmarks` (new repo)

---

## Benchmark Categories

### 1. HTTP Throughput

Measure requests per second under sustained load.

| Metric | Tool | Command |
|--------|------|---------|
| RPS (requests/sec) | wrk | `wrk -t12 -c400 -d30s` |
| Latency distribution | wrk | p50, p75, p90, p99 |

### 2. Resource Usage

Measure memory and CPU consumption.

| Metric | Tool | Method |
|--------|------|--------|
| Memory (RSS) | docker stats | Peak during load test |
| Memory (idle) | docker stats | After startup, before load |
| CPU % | docker stats | Average during load test |

### 3. Startup Time

Measure cold start performance.

| Metric | Method |
|--------|--------|
| Bootstrap time | Time from process start to first request ready |
| Build time (Go) | `time go build` |
| Compile time (NestJS) | `time npm run build` |

### 4. Scalability

Measure performance across different app sizes.

| Scenario | Providers | Modules | Purpose |
|----------|-----------|---------|---------|
| Small | 5 | 2 | Minimal overhead |
| Medium | 25 | 5 | Typical app |
| Large | 100 | 20 | Stress test |

---

## Frameworks Under Test

### Primary Comparison

| Framework | Language | Type | Why Include |
|-----------|----------|------|-------------|
| **modkit** | Go | Module system | Subject |
| **NestJS** | Node.js | Module system | Inspiration, cross-language |

### Go Alternatives

| Framework | Type | Why Include |
|-----------|------|-------------|
| **No framework** | Manual wiring | Baseline |
| **google/wire** | Compile-time codegen | Industry standard |
| **uber-go/fx** | Reflection-based | Similar to NestJS |
| **samber/do** | Generics-based | Modern alternative |

---

## Test Application

All frameworks implement the same application with identical behavior.

### Architecture

```
AppModule
├── ConfigModule
│   └── ConfigProvider (reads env vars)
├── DatabaseModule
│   └── ConnectionProvider (mock or real DB)
└── UsersModule
    ├── UsersController
    ├── UsersService
    └── UsersRepository
```

### Endpoints

| Method | Path | Description | DB? |
|--------|------|-------------|-----|
| GET | `/health` | Health check | No |
| GET | `/users` | List users (paginated) | Yes |
| GET | `/users/:id` | Get user by ID | Yes |
| POST | `/users` | Create user | Yes |
| PUT | `/users/:id` | Update user | Yes |
| DELETE | `/users/:id` | Delete user | Yes |

### Response Format

```json
{
  "id": 1,
  "name": "John Doe",
  "email": "john@example.com",
  "createdAt": "2024-01-15T10:30:00Z"
}
```

### Database

Use SQLite in-memory for consistency:
- No network latency variance
- Same SQL across all implementations
- Fast reset between runs

---

## Repository Structure

```
github.com/go-modkit/benchmarks/
├── README.md                    # Results and methodology
├── METHODOLOGY.md               # Detailed benchmark methodology
├── docker-compose.yml           # Run all benchmarks
├── Makefile                     # Convenience commands
│
├── apps/
│   ├── modkit/
│   │   ├── main.go
│   │   ├── go.mod
│   │   ├── modules/
│   │   │   ├── config/
│   │   │   ├── database/
│   │   │   └── users/
│   │   └── Dockerfile
│   │
│   ├── nestjs/
│   │   ├── src/
│   │   │   ├── app.module.ts
│   │   │   ├── config/
│   │   │   ├── database/
│   │   │   └── users/
│   │   ├── package.json
│   │   ├── tsconfig.json
│   │   └── Dockerfile
│   │
│   ├── wire/
│   │   ├── main.go
│   │   ├── wire.go
│   │   ├── wire_gen.go
│   │   └── Dockerfile
│   │
│   ├── fx/
│   │   ├── main.go
│   │   └── Dockerfile
│   │
│   ├── do/
│   │   ├── main.go
│   │   └── Dockerfile
│   │
│   └── baseline/                # No framework, manual wiring
│       ├── main.go
│       └── Dockerfile
│
├── scripts/
│   ├── run-all.sh               # Run complete benchmark suite
│   ├── run-single.sh            # Run single framework
│   ├── warmup.sh                # Warmup requests before benchmark
│   ├── collect-metrics.sh       # Collect docker stats
│   └── generate-report.py       # Generate charts and tables
│
├── results/
│   ├── latest/                  # Symlink to most recent
│   │   ├── raw/                 # Raw benchmark output
│   │   ├── summary.json         # Parsed results
│   │   ├── charts/              # Generated charts
│   │   └── report.md            # Generated report
│   └── archive/
│       └── 2024-01-15/
│
└── .github/
    └── workflows/
        ├── benchmark.yml        # Manual trigger
        └── scheduled.yml        # Weekly run
```

---

## Benchmark Execution

### Prerequisites

- Docker and Docker Compose
- wrk (HTTP benchmarking tool)
- Python 3 (for chart generation)

### Running Benchmarks

```bash
# Clone the repo
git clone https://github.com/go-modkit/benchmarks
cd benchmarks

# Run all benchmarks
make benchmark

# Run specific framework
make benchmark-modkit
make benchmark-nestjs

# Generate report
make report
```

### Benchmark Script Flow

```
1. Build all Docker images
2. For each framework:
   a. Start container
   b. Wait for health check
   c. Record startup time
   d. Warmup (1000 requests)
   e. Run benchmark (30s sustained load)
   f. Collect memory/CPU stats
   g. Stop container
3. Parse results
4. Generate charts
5. Generate report
```

### Benchmark Parameters

```bash
# wrk configuration
THREADS=12
CONNECTIONS=400
DURATION=30s

# Runs per framework (for statistical significance)
RUNS=3
```

---

## Results Presentation

### README.md Summary

```markdown
# modkit Benchmarks

Comparing modkit against NestJS and Go DI frameworks.

## Latest Results (2024-01-15)

### HTTP Throughput (GET /users/:id)

| Framework | RPS | Latency p50 | Latency p99 | Memory |
|-----------|-----|-------------|-------------|--------|
| baseline | 52,000 | 1.8ms | 6.2ms | 8MB |
| modkit | 48,000 | 2.0ms | 7.1ms | 12MB |
| wire | 51,000 | 1.9ms | 6.5ms | 9MB |
| fx | 45,000 | 2.2ms | 8.0ms | 18MB |
| do | 47,000 | 2.1ms | 7.5ms | 14MB |
| NestJS | 9,500 | 11ms | 42ms | 95MB |

### Requests per Second (Higher is Better)

[Chart: Bar chart showing RPS for each framework]

### Latency Distribution (Lower is Better)

[Chart: Box plot showing latency distribution]

### Memory Usage (Lower is Better)

[Chart: Bar chart showing memory consumption]

## Important Notes

This benchmark compares Go frameworks to a Node.js framework (NestJS).
The performance difference primarily reflects Go vs Node.js runtime 
characteristics, not just framework design.

### Why Compare to NestJS?

modkit is inspired by NestJS's module architecture. This benchmark 
demonstrates that you can have the same architectural benefits with 
Go's performance profile.

### When to Choose NestJS

- Team expertise in TypeScript/Node.js
- Need for NestJS's rich decorator ecosystem
- Rapid prototyping where raw performance isn't critical

### When to Choose modkit

- Performance-critical services
- Resource-constrained environments
- Preference for explicit over implicit behavior
- Existing Go expertise
```

### Chart Types

1. **Bar Chart: Requests per Second**
   - All frameworks side by side
   - Color-coded: Go (blue), Node.js (orange)

2. **Box Plot: Latency Distribution**
   - Shows p50, p75, p90, p99 for each framework
   - Highlights tail latency differences

3. **Bar Chart: Memory Usage**
   - Idle memory vs peak memory
   - Stacked or grouped bars

4. **Line Chart: Scalability**
   - X-axis: Number of providers
   - Y-axis: RPS or startup time
   - Shows how each framework scales

---

## Stories Breakdown

### Story 3.1: Repository Setup
**Points:** 2

- Create `go-modkit/benchmarks` repo
- Set up directory structure
- Create Makefile with common commands
- Write initial README with methodology

### Story 3.2: modkit Test App
**Points:** 2

- Implement modkit app with all endpoints
- SQLite in-memory database
- Dockerfile with optimized build
- Verify all endpoints work

### Story 3.3: NestJS Test App
**Points:** 3

- Implement NestJS app with identical endpoints
- Match response formats exactly
- SQLite with TypeORM or Prisma
- Dockerfile with production build
- Verify behavior matches modkit app

### Story 3.4: Go Alternatives (wire, fx, do, baseline)
**Points:** 4

- Implement same app in each framework
- Ensure identical behavior
- Dockerfiles for each
- Verify all endpoints

### Story 3.5: Benchmark Scripts
**Points:** 3

- Write `run-all.sh` orchestration script
- Implement warmup routine
- Collect docker stats (memory, CPU)
- Record startup times
- Output raw results to JSON

### Story 3.6: Report Generation
**Points:** 3

- Python script to parse results
- Generate charts (matplotlib or plotly)
- Generate markdown tables
- Create summary.json

### Story 3.7: CI/CD Integration
**Points:** 2

- GitHub Actions workflow for manual trigger
- Optional: scheduled weekly runs
- Upload results as artifacts
- Auto-commit to results/ directory

### Story 3.8: Documentation
**Points:** 2

- Write METHODOLOGY.md
- Document how to run locally
- Document how to interpret results
- Link from main modkit README

---

## Total Estimate

| Story | Points |
|-------|--------|
| 3.1 Repository Setup | 2 |
| 3.2 modkit Test App | 2 |
| 3.3 NestJS Test App | 3 |
| 3.4 Go Alternatives | 4 |
| 3.5 Benchmark Scripts | 3 |
| 3.6 Report Generation | 3 |
| 3.7 CI/CD Integration | 2 |
| 3.8 Documentation | 2 |
| **Total** | **21** |

---

## Technical Decisions

### 1. SQLite for Database

**Decision:** Use SQLite in-memory for all tests.

**Rationale:**
- Eliminates network latency variance
- Same SQL dialect works for Go and Node.js
- Fast reset between benchmark runs
- No external database container needed

### 2. Docker for Isolation

**Decision:** Run each framework in its own Docker container.

**Rationale:**
- Consistent environment
- Easy memory/CPU measurement via docker stats
- Reproducible across machines
- Clear resource boundaries

### 3. wrk for Load Testing

**Decision:** Use wrk (not ab, siege, or k6).

**Rationale:**
- High performance, low overhead
- Lua scripting for complex scenarios
- Widely used and understood
- Available on all platforms

### 4. Multiple Runs

**Decision:** Run each benchmark 3 times, report median.

**Rationale:**
- Reduces variance from system noise
- More statistically meaningful
- Catches outliers

### 5. Warmup Period

**Decision:** Send 1000 requests before measuring.

**Rationale:**
- JIT compilation for Node.js
- Connection pool warming
- CPU frequency scaling
- Consistent starting state

---

## Fair Comparison Guidelines

### What We Control

1. **Identical endpoints** - Same routes, same response format
2. **Same database** - SQLite in-memory for all
3. **Same hardware** - Run on identical Docker resource limits
4. **Same load** - Same wrk parameters for all
5. **Warmup** - Same warmup for all frameworks

### What We Acknowledge

1. **Language difference** - Go is compiled, Node.js is interpreted
2. **Runtime difference** - Go has different GC than V8
3. **Framework philosophy** - Some prioritize DX over performance

### Transparency Requirements

- Publish all source code
- Document exact versions (Go, Node, framework versions)
- Document hardware/VM specs
- Provide instructions to reproduce

---

## Success Metrics

1. **Reproducibility** - Anyone can clone and run benchmarks
2. **Fairness** - No artificial handicaps on any framework
3. **Clarity** - Results are easy to understand
4. **Usefulness** - Helps users make informed decisions

---

## Risks

| Risk | Mitigation |
|------|------------|
| Unfair comparison claims | Document methodology transparently |
| Results vary by machine | Use Docker with resource limits |
| Framework updates invalidate results | Include version numbers, re-run periodically |
| NestJS community backlash | Frame as "different tools for different needs" |

---

## Future Enhancements

1. **More frameworks** - Gin, Echo, Fiber (HTTP only, no DI)
2. **More scenarios** - WebSocket, file upload, streaming
3. **Cloud benchmarks** - AWS Lambda cold starts
4. **Interactive dashboard** - Web UI for exploring results
