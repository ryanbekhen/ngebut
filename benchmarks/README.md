# Web Framework Benchmarks

This directory contains benchmark tests for six different web frameworks:

1. **Ngebut** - A high-performance web framework
2. **Fiber** - A fast web framework built on top of Fasthttp
3. **Net/HTTP** - Go's standard HTTP package
4. **Chi** - A lightweight, idiomatic and composable router for Go
5. **Echo** - A high performance, extensible, minimalist web framework for Go
6. **Gin** - A HTTP web framework written in Go

Each framework implements the same basic functionality:
- A root route that returns a simple string
- A `/json` route that returns a JSON response
- A `/users/:id` route that demonstrates path parameter handling

## Directory Structure

```
benchmarks/
├── chi/         # Chi implementation
├── echo/        # Echo implementation
├── gin/         # Gin implementation
├── gofiber/     # Go Fiber implementation
├── nethttp/     # Standard net/http implementation
└── ngebut/      # Ngebut implementation
```

## Running the Benchmarks

You can run the benchmarks either manually or using the provided automated script.

### Automated Benchmarking

The easiest way to run benchmarks is to use the provided script:

```bash
cd benchmarks
./run_benchmarks.sh
```

This script will:
1. Check if port 3000 is in use and automatically kill any process using it
2. Start each server (ngebut, gofiber, nethttp, chi, echo, gin) one by one
3. Run wrk benchmarks against each endpoint (/, /json, /users/123)
4. Stop the server after benchmarking
5. Display a summary table at the end comparing all results side by side

The summary table shows requests per second and average latency for each framework and endpoint, making it easy to compare performance at a glance.

By default, the script uses the following parameters:
- Duration: 10 seconds per benchmark
- Threads: 4
- Connections: 100
- Port: 3000

You can modify these parameters by editing the script.

**Note:** The script uses `lsof` to check if the port is in use. If `lsof` is not installed on your system, you'll see a warning, but the script will still run. However, the port checking functionality may not work correctly.

### Manual Benchmarking

If you prefer to run the benchmarks manually, each directory contains its own Go module.

#### Starting the Servers

To run each server, navigate to its directory and run:

```bash
cd benchmarks/[framework]
go run main.go
```

Each server will start on port 3000 by default.

#### Running Benchmarks with wrk

[wrk](https://github.com/wg/wrk) is a modern HTTP benchmarking tool capable of generating significant load. To install wrk, follow the instructions on its GitHub page.

Once you have wrk installed, you can run benchmarks against each server:

```bash
# Basic string response benchmark
wrk -t12 -c400 -d30s http://localhost:3000/

# JSON response benchmark
wrk -t12 -c400 -d30s http://localhost:3000/json

# Path parameter benchmark
wrk -t12 -c400 -d30s http://localhost:3000/users/123
```

Parameters:
- `-t12`: Use 12 threads
- `-c400`: Keep 400 HTTP connections open
- `-d30s`: Run the test for 30 seconds

## Comparing Results

After running the benchmarks, you can compare the results to see which framework performs best for your specific use case. Look for metrics like:

- Requests per second
- Latency (average, min, max)
- Transfer rate

Remember that performance can vary based on hardware, operating system, and specific workload, so it's important to benchmark with scenarios that match your actual use case.
