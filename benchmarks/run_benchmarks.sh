#!/bin/bash

# Script to run benchmarks for all three web framework implementations
# Requires wrk to be installed: https://github.com/wg/wrk

# Configuration
DURATION=10s  # Duration for each benchmark
THREADS=4     # Number of threads
CONNECTIONS=100  # Number of connections
PORT=3000     # Port for all servers
ENDPOINTS=("/" "/json" "/users/123")  # Endpoints to test
FRAMEWORKS=("ngebut" "gofiber" "nethttp")  # Frameworks to test

# Create a temporary directory for results
TEMP_DIR=$(mktemp -d)
echo "Storing benchmark results in $TEMP_DIR"

# Check if wrk is installed
if ! command -v wrk &> /dev/null; then
    echo "Error: wrk is not installed. Please install it from https://github.com/wg/wrk"
    exit 1
fi

# Check if lsof is installed (needed for port checking)
if ! command -v lsof &> /dev/null; then
    echo "Warning: lsof is not installed. Port checking may not work correctly."
fi

# Function to run a benchmark for a specific framework and endpoint
run_benchmark() {
    local framework=$1
    local endpoint=$2
    local url="http://localhost:$PORT$endpoint"
    local result_file="$TEMP_DIR/${framework}_${endpoint//\//_}.txt"

    echo "Running benchmark for $framework - $endpoint"
    wrk -t$THREADS -c$CONNECTIONS -d$DURATION $url > "$result_file"
    echo ""
}

# Function to check if port is in use and kill the process
ensure_port_is_free() {
    local port=$1
    echo "Checking if port $port is in use..."

    # Find process using the port
    local pid=$(lsof -i :$port -t 2>/dev/null)

    if [ -n "$pid" ]; then
        echo "Port $port is in use by process $pid. Killing it..."
        kill -9 $pid 2>/dev/null
        sleep 1
        echo "Process killed"
    else
        echo "Port $port is free"
    fi
    echo ""
}

# Function to start a server
start_server() {
    local framework=$1

    # Ensure port is free before starting
    ensure_port_is_free $PORT

    echo "Starting $framework server..."
    cd "$(dirname "$0")/$framework"
    go run main.go &
    SERVER_PID=$!

    # Wait for server to start
    sleep 2

    # Check if server started successfully
    if ! lsof -i :$PORT -t &>/dev/null; then
        echo "ERROR: $framework server failed to start on port $PORT"
        # If we have a PID, try to kill it just in case
        if [ -n "$SERVER_PID" ]; then
            kill $SERVER_PID 2>/dev/null
            SERVER_PID=""
        fi
        return 1
    fi

    echo "$framework server started with PID: $SERVER_PID"
    echo ""

    # Return to the original directory
    cd - > /dev/null
    return 0
}

# Function to stop a server
stop_server() {
    if [ -n "$SERVER_PID" ]; then
        echo "Stopping server with PID: $SERVER_PID"
        kill $SERVER_PID
        wait $SERVER_PID 2>/dev/null
        echo "Server stopped"
        echo ""
        SERVER_PID=""
    fi
}

# Make sure port is free before starting any benchmarks
ensure_port_is_free $PORT

# Main benchmark loop
for framework in "${FRAMEWORKS[@]}"; do
    echo "====================================================="
    echo "Benchmarking $framework"
    echo "====================================================="

    # Start the server
    if start_server $framework; then
        # Run benchmarks for each endpoint
        for endpoint in "${ENDPOINTS[@]}"; do
            run_benchmark $framework $endpoint
        done

        # Stop the server
        stop_server
    else
        echo "Skipping benchmarks for $framework due to server startup failure"
    fi

    echo "====================================================="
    echo ""
done

# Display summary of all benchmark results
echo "====================================================="
echo "BENCHMARK SUMMARY"
echo "====================================================="
echo "Configuration: $THREADS threads, $CONNECTIONS connections, $DURATION duration"
echo ""

# Function to extract requests per second from wrk output
extract_rps() {
    grep "Requests/sec:" "$1" | awk '{print $2}'
}

# Function to extract average latency from wrk output
extract_latency() {
    grep "Latency" "$1" | awk '{print $2}'
}

# Display results in a table format
printf "%-10s %-15s %-20s %-20s\n" "Framework" "Endpoint" "Requests/sec" "Avg Latency"
echo "----------------------------------------------------------------------"

for framework in "${FRAMEWORKS[@]}"; do
    for endpoint in "${ENDPOINTS[@]}"; do
        result_file="$TEMP_DIR/${framework}_${endpoint//\//_}.txt"
        if [ -f "$result_file" ]; then
            rps=$(extract_rps "$result_file")
            latency=$(extract_latency "$result_file")
            printf "%-10s %-15s %-20s %-20s\n" "$framework" "$endpoint" "$rps" "$latency"
        else
            printf "%-10s %-15s %-20s %-20s\n" "$framework" "$endpoint" "N/A" "N/A"
        fi
    done
done

echo ""
echo "Detailed results are available in: $TEMP_DIR"
echo "All benchmarks completed!"
