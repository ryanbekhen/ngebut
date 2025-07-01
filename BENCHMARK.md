# 📊 Ngebut Benchmarks

This document contains benchmark information for the Ngebut framework.

## Internal Benchmarks

The framework includes comprehensive internal benchmarks that measure various aspects of performance. You can run these
benchmarks using:

```bash
go test -bench=. -benchmem
```

The internal benchmarks cover:

1. **Routing Performance**
    - Static routes
    - Parameter routes
    - Nested parameter routes
    - Deep nested routes

2. **Response Types**
    - String responses
    - JSON responses (small and large)
    - HTML responses

3. **Middleware Performance**
    - No middleware
    - Single middleware
    - Multiple middleware

4. **Group Routing**
    - Basic group routes
    - Nested group routes
    - Groups with middleware

5. **Context Operations**
    - Parameter access
    - Query parameter access
    - Header access
    - JSON binding

6. **HTTP Methods**
    - GET, POST, PUT, DELETE, PATCH

7. **Static File Serving**
    - HTML files
    - CSS files
    - Text files

*Note: Benchmark results may vary based on hardware and environment. Ngebut is still under active development and
performance optimizations are ongoing.*
