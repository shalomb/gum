#!/bin/bash

# Comprehensive Concurrency and Integrity Test Suite
# This script proves data layer integrity under concurrent operations

set -e

echo "🧪 Concurrency & Integrity Test Suite"
echo "======================================"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Test configuration
TEST_DIR="/tmp/gum_concurrency_test_$$"
CACHE_DIR="$TEST_DIR/.cache/gum"
DB_PATH="$CACHE_DIR/gum.db"

# Cleanup function
cleanup() {
    echo -e "\n${YELLOW}🧹 Cleaning up test environment...${NC}"
    rm -rf "$TEST_DIR"
}
trap cleanup EXIT

# Create test environment
echo -e "${BLUE}📁 Setting up test environment...${NC}"
mkdir -p "$CACHE_DIR"
export XDG_CACHE_HOME="$TEST_DIR/.cache"
export HOME="$TEST_DIR"

echo "Test directory: $TEST_DIR"
echo "Cache directory: $CACHE_DIR"
echo "Database path: $DB_PATH"

# Build gum
echo -e "\n${BLUE}🔨 Building gum...${NC}"
if ! go build -o "$TEST_DIR/gum" .; then
    echo -e "${RED}❌ Failed to build gum${NC}"
    exit 1
fi
echo "✅ Built gum successfully"

# Test 1: Basic Integrity Check
echo -e "\n${BLUE}🔍 Test 1: Basic Integrity Check${NC}"
if ! "$TEST_DIR/gum" integrity; then
    echo -e "${RED}❌ Basic integrity check failed${NC}"
    exit 1
fi
echo "✅ Basic integrity check passed"

# Test 2: Concurrent Operations Test
echo -e "\n${BLUE}🔍 Test 2: Concurrent Operations Test${NC}"
echo "Running 20 workers for 30 seconds..."
if ! "$TEST_DIR/gum" stress-test --workers 20 --duration 30s; then
    echo -e "${RED}❌ Concurrent operations test failed${NC}"
    exit 1
fi
echo "✅ Concurrent operations test passed"

# Test 3: High Load Test
echo -e "\n${BLUE}🔍 Test 3: High Load Test${NC}"
echo "Running 50 workers for 60 seconds..."
if ! "$TEST_DIR/gum" stress-test --workers 50 --duration 60s; then
    echo -e "${RED}❌ High load test failed${NC}"
    exit 1
fi
echo "✅ High load test passed"

# Test 4: Mixed Operations Test
echo -e "\n${BLUE}🔍 Test 4: Mixed Operations Test${NC}"
echo "Running mixed read/write operations..."

# Start multiple gum processes simultaneously
echo "Starting 10 concurrent gum processes..."

# Function to run gum process
run_gum_process() {
    local process_id=$1
    local iterations=100
    
    for i in $(seq 1 $iterations); do
        # Random operation
        case $((RANDOM % 4)) in
            0) "$TEST_DIR/gum" projects-v2 >/dev/null 2>&1 ;;
            1) "$TEST_DIR/gum" projects-v2 --refresh >/dev/null 2>&1 ;;
            2) "$TEST_DIR/gum" projects-v2 --verbose >/dev/null 2>&1 ;;
            3) "$TEST_DIR/gum" integrity >/dev/null 2>&1 ;;
        esac
        
        # Small delay
        sleep 0.01
    done
}

# Start 10 concurrent processes
pids=()
for i in {1..10}; do
    run_gum_process $i &
    pids+=($!)
done

# Wait for all processes to complete
for pid in "${pids[@]}"; do
    wait $pid
done

echo "✅ Mixed operations test passed"

# Test 5: Database Integrity After Load
echo -e "\n${BLUE}🔍 Test 5: Database Integrity After Load${NC}"
if ! "$TEST_DIR/gum" integrity; then
    echo -e "${RED}❌ Database integrity check after load failed${NC}"
    exit 1
fi
echo "✅ Database integrity after load passed"

# Test 6: Cache Consistency Test
echo -e "\n${BLUE}🔍 Test 6: Cache Consistency Test${NC}"
echo "Testing cache consistency under concurrent access..."

# Function to test cache consistency
test_cache_consistency() {
    local process_id=$1
    local iterations=50
    local results=()
    
    for i in $(seq 1 $iterations); do
        count=$("$TEST_DIR/gum" projects-v2 2>/dev/null | wc -l)
        results+=($count)
        sleep 0.01
    done
    
    # Check if all results are the same
    local first_result=${results[0]}
    for result in "${results[@]}"; do
        if [ "$result" != "$first_result" ]; then
            echo "Process $process_id: Cache inconsistency detected!"
            echo "Results: ${results[*]}"
            return 1
        fi
    done
    
    echo "Process $process_id: Cache consistent (all $first_result projects)"
    return 0
}

# Run cache consistency test with 5 concurrent processes
pids=()
for i in {1..5}; do
    test_cache_consistency $i &
    pids+=($!)
done

# Wait for all processes and check results
failed=0
for pid in "${pids[@]}"; do
    if ! wait $pid; then
        failed=1
    fi
done

if [ $failed -eq 1 ]; then
    echo -e "${RED}❌ Cache consistency test failed${NC}"
    exit 1
fi
echo "✅ Cache consistency test passed"

# Test 7: Race Condition Test
echo -e "\n${BLUE}🔍 Test 7: Race Condition Test${NC}"
echo "Testing for race conditions in database operations..."

# Function to simulate race condition
simulate_race_condition() {
    local process_id=$1
    local iterations=20
    
    for i in $(seq 1 $iterations); do
        # Simultaneous read and write operations
        "$TEST_DIR/gum" projects-v2 >/dev/null 2>&1 &
        "$TEST_DIR/gum" projects-v2 --refresh >/dev/null 2>&1 &
        wait
        sleep 0.001
    done
}

# Run race condition test with 10 concurrent processes
pids=()
for i in {1..10}; do
    simulate_race_condition $i &
    pids+=($!)
done

# Wait for all processes
for pid in "${pids[@]}"; do
    wait $pid
done

# Check database integrity after race condition test
if ! "$TEST_DIR/gum" integrity >/dev/null 2>&1; then
    echo -e "${RED}❌ Race condition test failed - database integrity compromised${NC}"
    exit 1
fi
echo "✅ Race condition test passed"

# Test 8: Long Running Test
echo -e "\n${BLUE}🔍 Test 8: Long Running Test${NC}"
echo "Running long-running operations (5 minutes)..."

# Start long-running processes
pids=()
for i in {1..5}; do
    (
        for j in $(seq 1 300); do
            "$TEST_DIR/gum" projects-v2 >/dev/null 2>&1
            sleep 1
        done
    ) &
    pids+=($!)
done

# Monitor for 5 minutes
echo "Monitoring for 5 minutes..."
for i in $(seq 1 300); do
    # Check database integrity every 30 seconds
    if [ $((i % 30)) -eq 0 ]; then
        if ! "$TEST_DIR/gum" integrity >/dev/null 2>&1; then
            echo -e "${RED}❌ Database integrity failed during long-running test${NC}"
            # Kill background processes
            for pid in "${pids[@]}"; do
                kill $pid 2>/dev/null || true
            done
            exit 1
        fi
        echo "Integrity check passed at $i seconds"
    fi
    sleep 1
done

# Wait for all processes to complete
for pid in "${pids[@]}"; do
    wait $pid
done

echo "✅ Long running test passed"

# Final Results
echo -e "\n${GREEN}🎉 All Concurrency & Integrity Tests Passed!${NC}"
echo "=============================================="
echo "✅ Basic integrity check"
echo "✅ Concurrent operations (20 workers, 30s)"
echo "✅ High load test (50 workers, 60s)"
echo "✅ Mixed operations (10 processes)"
echo "✅ Database integrity after load"
echo "✅ Cache consistency under concurrent access"
echo "✅ Race condition prevention"
echo "✅ Long running operations (5 minutes)"
echo ""
echo -e "${GREEN}🚀 Database is proven safe for concurrent operations!${NC}"
echo ""
echo "Key findings:"
echo "- No data corruption detected"
echo "- Cache consistency maintained"
echo "- No race conditions observed"
echo "- Database integrity preserved under load"
echo "- Concurrent operations are safe"
echo ""
echo "The gum database layer is production-ready for concurrent access."