# Database Migration Implementation Plan

## Phase 1: Core Infrastructure (Week 1)

### 1.1 Database Schema Updates
- [ ] Update `internal/database/schema.sql` to v2
- [ ] Add `github_repo_id` foreign key to projects table
- [ ] Add `cache_metadata` table for TTL management
- [ ] Add indexes for performance optimization
- [ ] Create migration scripts for existing databases

### 1.2 Migration Framework
- [ ] Implement `internal/database/migration.go`
- [ ] Add JSON parsing and validation
- [ ] Implement backup and restore functionality
- [ ] Add rollback mechanism
- [ ] Create error handling and recovery

### 1.3 Database Cache Layer
- [ ] Implement `internal/database/cache.go`
- [ ] Add TTL-based cache invalidation
- [ ] Implement atomic operations
- [ ] Add cache statistics and monitoring
- [ ] Create concurrent access safety

## Phase 2: Command Implementation (Week 2)

### 2.1 Migration Command
- [ ] Implement `cmd/migrate.go`
- [ ] Add migration status checking
- [ ] Implement backup/restore commands
- [ ] Add rollback functionality
- [ ] Create progress reporting

### 2.2 New Projects Command
- [ ] Implement `cmd/projects_v2.go`
- [ ] Add GitHub metadata integration
- [ ] Implement consistent discovery algorithm
- [ ] Add verbose output and statistics
- [ ] Create performance optimizations

### 2.3 Update Existing Commands
- [ ] Update `cmd/update.go` to use database
- [ ] Modify `cmd/dirs.go` for consistency
- [ ] Update `cmd/sync.go` for GitHub integration
- [ ] Ensure all commands use unified cache

## Phase 3: Testing & Validation (Week 3)

### 3.1 Unit Tests
- [ ] Test migration from JSON to database
- [ ] Test GitHub repository linking
- [ ] Test cache consistency and TTL
- [ ] Test concurrent access safety
- [ ] Test error handling and recovery

### 3.2 Integration Tests
- [ ] Test end-to-end migration process
- [ ] Test rollback functionality
- [ ] Test performance improvements
- [ ] Test data integrity
- [ ] Test cron job compatibility

### 3.3 BDD Tests
- [ ] Implement `docs/bdd/database-migration.feature`
- [ ] Test migration scenarios
- [ ] Test GitHub integration
- [ ] Test cache consistency
- [ ] Test performance improvements

## Phase 4: Documentation & Deployment (Week 4)

### 4.1 Documentation
- [ ] Update `docs/features/database-migration.md`
- [ ] Create migration guide
- [ ] Update command reference
- [ ] Add troubleshooting guide
- [ ] Create performance benchmarks

### 4.2 Deployment Strategy
- [ ] Create migration script
- [ ] Add safety checks and validation
- [ ] Implement gradual rollout
- [ ] Add monitoring and alerting
- [ ] Create rollback procedures

## Implementation Details

### Database Schema Changes

```sql
-- Add GitHub repo link to projects
ALTER TABLE projects ADD COLUMN github_repo_id INTEGER;
ALTER TABLE projects ADD FOREIGN KEY (github_repo_id) REFERENCES github_repos(id);

-- Add cache metadata table
CREATE TABLE cache_metadata (
    cache_key TEXT PRIMARY KEY,
    last_updated DATETIME,
    ttl_seconds INTEGER
);

-- Add GitHub metadata columns
ALTER TABLE github_repos ADD COLUMN language TEXT;
ALTER TABLE github_repos ADD COLUMN topics TEXT;
ALTER TABLE github_repos ADD COLUMN star_count INTEGER;
ALTER TABLE github_repos ADD COLUMN fork_count INTEGER;
ALTER TABLE github_repos ADD COLUMN open_issues_count INTEGER;
ALTER TABLE github_repos ADD COLUMN pushed_at DATETIME;
```

### Migration Process

1. **Pre-Migration Checks**
   - Verify JSON file integrity
   - Check database schema version
   - Validate available disk space
   - Create backup of current state

2. **Migration Execution**
   - Parse JSON cache files
   - Insert data into database
   - Link GitHub repositories
   - Update cache metadata
   - Backup original files

3. **Post-Migration Validation**
   - Verify data integrity
   - Test command functionality
   - Check performance metrics
   - Validate cache consistency

### Command Interface

```bash
# Migration commands
gum migrate                    # Run migration
gum migrate --rollback         # Rollback migration
gum migrate --backup /path     # Create backup
gum migrate --restore /path    # Restore from backup

# New projects command
gum projects-v2                # List projects from database
gum projects-v2 --refresh      # Force refresh
gum projects-v2 --with-github  # Include GitHub metadata
gum projects-v2 --verbose      # Show cache stats
gum projects-v2 --clear-cache  # Clear cache
```

### Performance Targets

- **Query Time**: < 100ms for 1000+ projects
- **Cache Hit Rate**: > 95% for repeated queries
- **Memory Usage**: < 50MB for database operations
- **Migration Time**: < 30 seconds for typical datasets
- **Concurrent Access**: Support 10+ simultaneous operations

### Error Handling

1. **Migration Errors**
   - Invalid JSON files → Skip and log warning
   - Database errors → Rollback and report
   - Disk space issues → Abort with clear message
   - Permission errors → Provide fix instructions

2. **Runtime Errors**
   - Database corruption → Auto-repair or recreate
   - Cache inconsistency → Force refresh
   - GitHub API errors → Graceful degradation
   - Concurrent access → Retry with backoff

### Monitoring & Observability

1. **Cache Statistics**
   - Hit/miss rates
   - Query performance
   - Memory usage
   - Data freshness

2. **Migration Metrics**
   - Migration success rate
   - Data integrity checks
   - Performance improvements
   - Error rates

3. **Health Checks**
   - Database integrity
   - Cache consistency
   - GitHub API connectivity
   - Disk space usage

### Rollback Strategy

1. **Automatic Rollback**
   - Migration validation fails
   - Critical errors during migration
   - Performance degradation detected

2. **Manual Rollback**
   - User-initiated rollback
   - Restore from backup
   - Revert to JSON caches

3. **Partial Rollback**
   - Keep successful migrations
   - Retry failed components
   - Maintain data integrity

### Testing Strategy

1. **Unit Tests**
   - Individual component testing
   - Mock external dependencies
   - Edge case validation
   - Error condition testing

2. **Integration Tests**
   - End-to-end workflows
   - Real database operations
   - Concurrent access testing
   - Performance benchmarking

3. **BDD Tests**
   - User scenario validation
   - Business logic testing
   - Acceptance criteria verification
   - Regression testing

### Deployment Checklist

- [ ] All tests passing
- [ ] Documentation updated
- [ ] Migration scripts tested
- [ ] Rollback procedures verified
- [ ] Performance benchmarks met
- [ ] Security review completed
- [ ] Backup procedures tested
- [ ] Monitoring configured
- [ ] User communication sent
- [ ] Support team trained

### Success Criteria

1. **Functional Requirements**
   - Cache inconsistency bug fixed
   - All existing functionality preserved
   - GitHub integration working
   - Performance improvements achieved

2. **Non-Functional Requirements**
   - Migration completes successfully
   - No data loss during migration
   - Rollback works if needed
   - Performance targets met

3. **User Experience**
   - Seamless migration process
   - Improved command performance
   - Better error messages
   - Enhanced functionality

### Risk Mitigation

1. **Data Loss Prevention**
   - Multiple backup strategies
   - Validation at each step
   - Atomic operations where possible
   - Rollback capabilities

2. **Performance Issues**
   - Gradual rollout
   - Performance monitoring
   - Load testing
   - Optimization strategies

3. **Compatibility Issues**
   - Backward compatibility
   - Version checking
   - Graceful degradation
   - Clear error messages

### Post-Migration Tasks

1. **Immediate (Day 1)**
   - Monitor migration success
   - Verify data integrity
   - Check performance metrics
   - Address any issues

2. **Short-term (Week 1)**
   - Gather user feedback
   - Optimize performance
   - Fix any bugs
   - Update documentation

3. **Long-term (Month 1)**
   - Analyze usage patterns
   - Plan future enhancements
   - Optimize database schema
   - Consider additional features

This implementation plan provides a comprehensive roadmap for migrating gum from JSON-based caching to a unified SQLite database system, ensuring the cache inconsistency bug is fixed while providing significant performance and feature improvements.