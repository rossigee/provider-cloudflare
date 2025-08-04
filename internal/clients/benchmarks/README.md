# Client Performance Benchmarks

This directory contains performance benchmarks for Cloudflare provider clients to ensure optimal performance and track regressions.

## Overview

Performance benchmarks help us:
- **Track performance regressions** across releases
- **Optimize critical code paths** in client operations
- **Validate scaling characteristics** under load
- **Compare implementation approaches** objectively
- **Set performance baselines** for SLA compliance

## Benchmark Categories

### 🚀 Client Operations
- **CRUD Operations**: Create, Read, Update, Delete performance
- **Batch Operations**: Bulk operations and parallelization
- **Connection Management**: HTTP client connection pooling
- **Authentication**: API token validation and caching
- **Error Handling**: Error response parsing and retry logic

### 📊 Data Processing
- **JSON Marshaling/Unmarshaling**: Object serialization performance
- **Type Conversion**: Provider to Cloudflare type mapping
- **Validation**: Input validation and sanitization
- **Transformation**: Data transformation between formats

### 🌐 Network Operations
- **HTTP Request Performance**: Latency and throughput metrics
- **Retry Logic**: Backoff and retry strategy performance
- **Rate Limiting**: Client-side rate limiting efficiency
- **Connection Pooling**: HTTP connection reuse optimization

## Running Benchmarks

### Individual Benchmarks
```bash
# Run all benchmarks
go test -bench=. ./internal/clients/benchmarks/...

# Run specific benchmark category
go test -bench=BenchmarkZone ./internal/clients/benchmarks/

# Run with memory profiling
go test -bench=. -benchmem ./internal/clients/benchmarks/

# Extended benchmarking
go test -bench=. -benchtime=10s ./internal/clients/benchmarks/
```

### Comparative Analysis
```bash
# Baseline benchmarks
go test -bench=. ./internal/clients/benchmarks/ > baseline.txt

# After changes
go test -bench=. ./internal/clients/benchmarks/ > current.txt

# Compare results
benchcmp baseline.txt current.txt
```

### Continuous Integration
```bash
# Automated benchmark in CI
make benchmark
make benchmark-compare
```

## Benchmark Structure

Each client has corresponding benchmarks:
- `zone_benchmark_test.go` - Zone operations
- `record_benchmark_test.go` - DNS record operations  
- `loadbalancing_benchmark_test.go` - Load balancing operations
- `cache_benchmark_test.go` - Cache rule operations
- `security_benchmark_test.go` - Security rule operations

## Performance Targets

### Latency Targets (95th percentile)
- **Zone Operations**: < 500ms
- **DNS Records**: < 200ms  
- **Load Balancing**: < 300ms
- **Security Rules**: < 400ms
- **Cache Rules**: < 250ms

### Throughput Targets
- **Concurrent Operations**: 50+ ops/sec
- **Batch Operations**: 100+ items/sec
- **Memory Usage**: < 50MB for typical workloads
- **CPU Usage**: < 30% for sustained operations

### Scaling Characteristics
- **Linear scaling** up to 100 concurrent operations
- **Bounded memory usage** regardless of operation count
- **Graceful degradation** under rate limiting
- **Efficient connection reuse** for bulk operations

## Benchmark Metrics

### Core Metrics
- **Operations per second** (ops/sec)
- **Latency percentiles** (p50, p95, p99)
- **Memory allocations** (allocs/op)
- **Memory usage** (bytes/op)
- **CPU utilization** during operations

### Extended Metrics  
- **HTTP connection efficiency**
- **Rate limiting compliance**
- **Error handling overhead**
- **Retry logic performance**
- **Authentication caching efficiency**

## Profiling Integration

### CPU Profiling
```bash
go test -bench=BenchmarkZone -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

### Memory Profiling
```bash
go test -bench=BenchmarkZone -memprofile=mem.prof
go tool pprof mem.prof
```

### Trace Analysis
```bash
go test -bench=BenchmarkZone -trace=trace.out
go tool trace trace.out
```

## Best Practices

### Benchmark Design
- **Realistic test data** matching production patterns
- **Isolated test environment** without external dependencies
- **Consistent test conditions** for reliable comparisons
- **Multiple iterations** for statistical significance
- **Resource cleanup** to avoid interference

### Performance Optimization
- **Connection pooling** for HTTP clients
- **JSON streaming** for large payloads
- **Batch operations** where supported by API
- **Efficient error handling** without performance penalties
- **Memory pooling** for frequently allocated objects

### Monitoring Integration
- **Automated benchmark runs** in CI/CD pipeline
- **Performance regression detection** with alerts
- **Historical trend tracking** across releases
- **Performance dashboard** for visibility
- **SLA compliance monitoring** against targets