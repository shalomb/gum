# Concurrency Testing Guide

## Overview

This guide explains how to test gum's concurrency safety and data integrity under concurrent operations. These tests are essential for verifying that multiple gum processes can run simultaneously without data corruption.

## Quick Start

### Run Basic Concurrency Test

```bash
# Run the comprehensive concurrency test
./test_manual_concurrency.sh
```

This will:
- Test cache consistency under concurrent access
- Verify mixed operations (read/write/refresh)
- Check database integrity after load
- Test race condition prevention
- Run long-running operations

### Check Database Integrity

```bash
# Verify database integrity
gum integrity
```

This checks:
- Database file integrity
- Foreign key constraints
- Orphaned records
- Duplicate records
- Cache consistency

## Test Scenarios

### 1. Cache Consistency Test

**Purpose**: Verify that multiple processes get consistent results

```bash
# Test cache consistency
gum projects-v2 | wc -l && gum projects-v2 | wc -l && gum projects-v2 | wc -l
# All should return the same number
```

**Expected Result**: All commands return identical project counts

### 2. Mixed Operations Test

**Purpose**: Test concurrent read/write operations

```bash
# Run different operations concurrently
gum projects --refresh &
gum sync &
gum projects-v2 --verbose &
wait

# Check integrity
gum integrity
```

**Expected Result**: All operations complete successfully, integrity check passes

### 3. Race Condition Test

**Purpose**: Test simultaneous operations on the same data

```bash
# Run simultaneous operations
gum projects --refresh &
gum projects --refresh &
gum projects --refresh &
wait

# Verify consistency
gum projects | wc -l
```

**Expected Result**: No race conditions, consistent results

### 4. Long-Running Test

**Purpose**: Test system stability over time

```bash
# Run long-running operations
for i in {1..3}; do
    (for j in {1..60}; do gum projects-v2 >/dev/null; sleep 1; done) &
done
wait
```

**Expected Result**: All operations complete successfully

## Advanced Testing

### Custom Concurrency Test

Create your own test script:

```bash
#!/bin/bash
# custom_concurrency_test.sh

echo "Testing concurrent gum operations..."

# Test 1: Multiple readers
for i in {1..5}; do
    gum projects-v2 >/dev/null &
done
wait

# Test 2: Mixed operations
gum projects --refresh &
gum sync &
gum projects-v2 --verbose &
wait

# Test 3: Verify integrity
gum integrity

echo "Test completed successfully!"
```

### Stress Testing

For heavy load testing:

```bash
# Run multiple processes with high frequency
for i in {1..10}; do
    (for j in {1..100}; do gum projects-v2 >/dev/null; sleep 0.1; done) &
done
wait
```

## Monitoring and Debugging

### Check Database Status

```bash
# Check database integrity
gum integrity

# Check cache statistics
gum projects-v2 --verbose

# Check database locks
sqlite3 ~/.cache/gum/gum.db "PRAGMA database_list;"
```

### Debug Concurrency Issues

```bash
# Check for database locks
sqlite3 ~/.cache/gum/gum.db "PRAGMA locking_mode;"

# Check WAL mode
sqlite3 ~/.cache/gum/gum.db "PRAGMA journal_mode;"

# Check foreign keys
sqlite3 ~/.cache/gum/gum.db "PRAGMA foreign_keys;"
```

### Performance Monitoring

```bash
# Time operations
time gum projects-v2

# Monitor system resources
htop

# Check database size
ls -lh ~/.cache/gum/gum.db
```

## Troubleshooting

### Common Issues

1. **Database Locked**
   - **Cause**: Another process is writing to the database
   - **Solution**: Wait and retry, or check for hanging processes

2. **Cache Inconsistency**
   - **Cause**: Stale cache data
   - **Solution**: Run `gum projects --refresh`

3. **Integrity Failures**
   - **Cause**: Database corruption or constraint violations
   - **Solution**: Check `gum integrity` output and fix issues

4. **Performance Degradation**
   - **Cause**: Too many concurrent operations
   - **Solution**: Reduce concurrency or optimize queries

### Debug Commands

```bash
# Check process status
ps aux | grep gum

# Check database locks
lsof ~/.cache/gum/gum.db

# Check system load
uptime

# Check memory usage
free -h
```

## Best Practices

### For Testing

1. **Start Small**: Begin with simple tests before complex scenarios
2. **Monitor Resources**: Watch CPU, memory, and disk usage
3. **Verify Results**: Always check integrity after tests
4. **Clean Up**: Remove test data after testing

### For Production

1. **Regular Testing**: Run concurrency tests regularly
2. **Monitor Integrity**: Check `gum integrity` periodically
3. **Handle Errors**: Implement proper error handling
4. **Document Issues**: Keep track of any problems found

## Test Results Interpretation

### Success Criteria

- ✅ All operations complete successfully
- ✅ Database integrity check passes
- ✅ Cache consistency maintained
- ✅ No race conditions detected
- ✅ Performance remains stable

### Failure Indicators

- ❌ Operations fail or hang
- ❌ Database integrity check fails
- ❌ Inconsistent results between calls
- ❌ Race conditions or deadlocks
- ❌ Performance degradation

## Continuous Integration

### Automated Testing

Add to your CI pipeline:

```yaml
# .github/workflows/concurrency-test.yml
name: Concurrency Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - name: Run concurrency test
        run: ./test_manual_concurrency.sh
      - name: Check integrity
        run: gum integrity
```

### Pre-commit Hooks

```bash
#!/bin/bash
# .git/hooks/pre-commit
echo "Running concurrency tests..."
./test_manual_concurrency.sh
if [ $? -ne 0 ]; then
    echo "Concurrency tests failed!"
    exit 1
fi
echo "All tests passed!"
```

## Conclusion

Concurrency testing is essential for ensuring gum's reliability in production environments. The comprehensive test suite verifies that the system can handle real-world concurrent usage patterns safely and reliably. Regular testing helps identify and prevent issues before they affect users.