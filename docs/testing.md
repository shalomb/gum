# Gum Testing Guide

## Overview

Gum has a comprehensive test suite covering unit tests, integration tests, and database operations. This guide explains how to run tests, understand test results, and contribute to the test suite. The test suite ensures reliability, performance, and correctness of the SQLite-based project management system.

## Test Structure

### Unit Tests

#### Database Tests (`internal/database/database_test.go`)
- **Coverage**: 83.2% of statements
- **CRUD Operations**: Test all database operations (Create, Read, Update, Delete)
- **Concurrency**: Test concurrent database access with WAL mode
- **Schema**: Verify database schema initialization and constraints
- **Performance**: Test query performance and indexing

**Key Test Areas:**
- Project management (insert, update, query)
- Project directory tracking
- GitHub repository storage
- Directory usage statistics
- Similarity scoring
- Database statistics and maintenance

#### Cache Tests (`internal/cache/cache_test.go`)
- **Coverage**: 85.3% of statements
- **Legacy Cache**: Test file-based caching (being phased out)
- **Concurrency**: Test concurrent cache operations
- **File Operations**: Test cache file creation and corruption handling
- **XDG Compliance**: Test XDG directory usage

**Key Test Areas:**
- Cache set/get operations
- Legacy cache functionality
- Concurrent access patterns
- File system operations
- Error handling and recovery

#### Command Tests (`cmd/projects_test.go`, `cmd/frecency_test.go`, `cmd/sync_test.go`)
- **Coverage**: 34.2% of statements
- **CLI Functionality**: Test command-line interface
- **Output Formats**: Test all output formats (default, JSON, FZF, simple)
- **Directory Discovery**: Test project directory scanning
- **Git Operations**: Test Git repository detection
- **Frecency Algorithm**: Test frecency scoring properties and edge cases
- **Directory Management**: Test directory tracking and historical import
- **Locate Integration**: Test locate database usage and fallback behavior
- **GitHub Sync**: Test repository metadata synchronization
- **Similarity Functions**: Test project similarity algorithms

#### Concurrency Tests (`internal/database/concurrency_test.go`)
- **Coverage**: 100% of concurrency scenarios
- **Concurrent Upserts**: Test multiple goroutines upserting the same project
- **Mixed Operations**: Test concurrent read/write operations
- **Transaction Integrity**: Test transaction rollback and consistency
- **Cache Consistency**: Test cache consistency under concurrent load
- **Database Integrity**: Test database integrity after concurrent operations
- **Race Condition Prevention**: Test for race conditions and deadlocks
- **Long-Running Operations**: Test system stability over extended periods

**Key Test Areas:**
- Project directory discovery
- Git repository scanning
- Output formatting
- Similarity scoring algorithms
- Command-line argument handling
- Frecency algorithm validation
- Directory tracking and scoring
- Legacy cache import functionality
- Locate database integration and fallback
- GitHub API integration and caching
- Repository metadata synchronization

### Integration Tests (`integration_test.go`)
- **End-to-End Testing**: Full command execution
- **Database Integration**: Test database creation and persistence
- **XDG Compliance**: Test XDG environment variable handling
- **Error Handling**: Test invalid commands and flags

**Key Test Areas:**
- Command execution
- Database file creation
- Cache operations
- Environment variable handling
- Error scenarios

## Test Coverage

### Overall Coverage: 53.9%

| Package | Coverage | Status |
|---------|----------|--------|
| `internal/database` | 83.2% | ✅ Excellent |
| `internal/cache` | 85.3% | ✅ Excellent |
| `cmd` | 34.2% | ⚠️ Needs improvement |

### Coverage Details

#### High Coverage Areas (>80%)
- Database CRUD operations
- Cache TTL and concurrency
- Project similarity algorithms
- Output formatting functions
- Git repository detection

#### Medium Coverage Areas (50-80%)
- Database initialization
- Cache directory management
- Project directory discovery
- Command execution flow

#### Low Coverage Areas (<50%)
- GitHub API integration
- Clone command functionality
- Update command operations
- Error handling paths

## Running Tests

### Unit Tests Only
```bash
# Run all unit tests
go test ./internal/... ./cmd

# Run with coverage
go test ./internal/... ./cmd -coverprofile=coverage.out

# Generate coverage report
go tool cover -html=coverage.out -o coverage.html
go tool cover -func=coverage.out
```

### Individual Package Tests
```bash
# Database tests
go test ./internal/database -v

# Cache tests
go test ./internal/cache -v

# Command tests
go test ./cmd -v
```

### Integration Tests
```bash
# All tests (including integration)
go test ./... -v

# Integration tests only
go test -run TestIntegration -v
```

### Concurrency Tests
```bash
# Run comprehensive concurrency tests
./test_manual_concurrency.sh

# Run specific concurrency tests
go test ./internal/database -run TestConcurrent -v

# Test database integrity
gum integrity

# Test cache consistency
gum projects-v2 | wc -l && gum projects-v2 | wc -l
```

## Test Data

### Test Files
- `testdata/projects-dirs.list`: Sample configuration file for testing

### Test Environment
- **Temporary Directories**: All tests use `t.TempDir()` for isolation
- **Environment Variables**: Tests set `XDG_CACHE_HOME` and `XDG_CONFIG_HOME`
- **Database**: Tests create isolated SQLite databases
- **Cleanup**: Automatic cleanup of test artifacts

## Test Patterns

### Frecency Algorithm Testing
```go
func TestFrecencyScore(t *testing.T) {
    now := time.Now()
    
    tests := []struct {
        name      string
        frequency int
        age       time.Duration
        wantMin   int64
        wantMax   int64
    }{
        {
            name:      "Recent high frequency",
            frequency: 100,
            age:       30 * time.Minute,
            wantMin:   4000,
            wantMax:   5000,
        },
        // ... more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            lastSeen := now.Add(-tt.age)
            score := calculateFrecencyScore(tt.frequency, lastSeen, now)
            
            if score < tt.wantMin || score > tt.wantMax {
                t.Errorf("calculateFrecencyScore() = %d, want between %d and %d", 
                    score, tt.wantMin, tt.wantMax)
            }
        })
    }
}
```

### Locate Integration Testing
```go
func TestLocateIntegration(t *testing.T) {
    finder := locate.NewLocateFinder()
    
    if !finder.GetStatus().Available {
        t.Skip("locate not available")
    }
    
    // Test database freshness
    status := finder.GetStatus()
    if !status.IsFresh {
        t.Logf("Warning: locate database is %v old", status.Age)
    }
    
    // Test git repository discovery
    repos, err := finder.FindGitRepos("/home/user")
    if err != nil {
        t.Fatalf("FindGitRepos failed: %v", err)
    }
    
    // Validate results
    for _, repo := range repos {
        if !strings.HasPrefix(repo, "/home/user") {
            t.Errorf("Repo %s not in expected path", repo)
        }
    }
}
```

### GitHub Sync Testing
```go
func TestGitHubSync(t *testing.T) {
    client, err := github.NewGitHubClient()
    if err != nil {
        t.Skip("GitHub authentication required")
    }
    
    // Test repository discovery
    repos, err := client.DiscoverAllRepositories(context.Background())
    if err != nil {
        t.Fatalf("DiscoverAllRepositories failed: %v", err)
    }
    
    if len(repos) == 0 {
        t.Error("No repositories discovered")
    }
    
    // Test metadata extraction
    for _, repo := range repos[:5] { // Test first 5 repos
        if repo.FullName == "" {
            t.Error("Repository missing full_name")
        }
        if repo.Name == "" {
            t.Error("Repository missing name")
        }
    }
}

func TestSyncCommand(t *testing.T) {
    // Test dry-run sync
    cmd := exec.Command("./gum", "sync", "--dry-run", "--type", "incremental")
    output, err := cmd.Output()
    if err != nil {
        t.Fatalf("Dry-run sync failed: %v", err)
    }
    
    if !strings.Contains(string(output), "Starting incremental sync") {
        t.Error("Expected sync start message")
    }
}
```

### Database Testing
```go
func TestDatabaseOperations(t *testing.T) {
    // Create isolated test database
    tempDir := t.TempDir()
    os.Setenv("XDG_CACHE_HOME", tempDir)
    
    db, err := New()
    if err != nil {
        t.Fatalf("Failed to create database: %v", err)
    }
    defer db.Close()
    
    // Test operations
    // ...
}
```

### Cache Testing
```go
func TestCacheOperations(t *testing.T) {
    // Create isolated test cache
    tempDir := t.TempDir()
    os.Setenv("XDG_CACHE_HOME", tempDir)
    
    cache := New()
    
    // Test cache operations
    // ...
}
```

### CLI Testing
```go
func TestProjectsCommand(t *testing.T) {
    // Set up test environment
    tempDir := t.TempDir()
    os.Setenv("XDG_CACHE_HOME", tempDir)
    
    // Test command functions
    // ...
}
```

## Performance Testing

### Concurrency Tests
- **Database**: Test concurrent reads/writes with WAL mode
- **Cache**: Test concurrent cache operations
- **Commands**: Test command execution under load

### Benchmarking
```bash
# Run benchmarks
go test -bench=. ./...

# Memory profiling
go test -memprofile=mem.prof ./...
```

## Continuous Integration

### Test Requirements
- All unit tests must pass
- Coverage should be >80% for core packages
- No race conditions in concurrent tests
- XDG compliance verified

### Test Environment
- **Go Version**: 1.21+
- **SQLite**: Available system dependency
- **OS**: Linux (primary), macOS, Windows

## Future Improvements

### Coverage Improvements
1. **GitHub Integration**: Add tests for GitHub API calls
2. **Clone Command**: Test repository cloning functionality
3. **Update Command**: Test database update operations
4. **Error Handling**: Test error scenarios and recovery

### Test Enhancements
1. **Mocking**: Add mocks for external dependencies
2. **Benchmarks**: Add performance benchmarks
3. **Property Testing**: Add property-based tests
4. **Fuzzing**: Add fuzz testing for input validation

### Integration Testing
1. **End-to-End**: Complete workflow testing
2. **Performance**: Load testing with large datasets
3. **Compatibility**: Cross-platform testing
4. **Migration**: Test data migration scenarios

## Troubleshooting

### Common Issues

#### Test Failures
- **Environment**: Ensure XDG environment variables are set
- **Dependencies**: Verify SQLite is available
- **Permissions**: Check file system permissions
- **Cleanup**: Ensure test cleanup is working

#### Coverage Issues
- **Missing Tests**: Add tests for uncovered functions
- **Edge Cases**: Test error conditions and edge cases
- **Integration**: Add integration tests for complex workflows

#### Performance Issues
- **Concurrency**: Check for race conditions
- **Memory**: Monitor memory usage in tests
- **Database**: Verify database performance under load

## Best Practices

### Test Organization
- **Package Structure**: Mirror source package structure
- **Test Naming**: Use descriptive test names
- **Test Data**: Use realistic test data
- **Cleanup**: Always clean up test artifacts

### Test Quality
- **Isolation**: Tests should be independent
- **Deterministic**: Tests should produce consistent results
- **Fast**: Tests should run quickly
- **Reliable**: Tests should not be flaky

### Maintenance
- **Regular Updates**: Keep tests up to date with code changes
- **Coverage Monitoring**: Track coverage trends
- **Performance**: Monitor test execution time
- **Documentation**: Keep test documentation current