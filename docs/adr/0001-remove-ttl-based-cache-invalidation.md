# ADR-0001: Remove TTL-Based Cache Invalidation

## Status
**ACCEPTED** - 2024-01-15

## Context

The gum tool currently uses TTL (Time-To-Live) based cache invalidation for both `projects` and `dirs` commands:

- **Projects**: 5-minute TTL, triggers file system scan on cache miss
- **Dirs**: 30-second TTL, triggers process scan on cache miss

However, the tool is designed to be used with cron jobs that periodically refresh the database:

```bash
# Current cron jobs
0 */6 * * * gum projects --refresh  # Every 6 hours
0 * * * * gum dirs --refresh        # Every hour
```

## Problem

TTL-based cache invalidation creates several issues:

1. **Unnecessary File System Scans**: When TTL expires, users experience slow responses due to file system discovery
2. **Inconsistent with Design**: Cron jobs are meant to keep data fresh, making TTL redundant
3. **Poor User Experience**: Users expect immediate responses, not discovery delays
4. **Data Nature Mismatch**: 
   - Projects are long-lived entities (Git repos don't disappear)
   - Dirs command needs ALL historical data, not just recent data

## Decision

**Remove TTL-based cache invalidation entirely** and rely on cron job-based updates.

### Changes

1. **Remove TTL Logic**: Delete `IsCacheValid()` and TTL checking
2. **Simplify Cache Methods**: Always return database data directly
3. **Keep Manual Refresh**: Maintain `--refresh` flags for manual updates
4. **Rely on Cron Jobs**: Database freshness managed by background processes

## Consequences

### Positive

- ✅ **Instant Response**: No file system scans during user commands
- ✅ **Consistent Performance**: Predictable response times
- ✅ **Simpler Architecture**: Remove complex TTL logic
- ✅ **Better User Experience**: Immediate data return
- ✅ **Aligns with Design**: Cron jobs handle data freshness

### Negative

- ❌ **Stale Data Risk**: If cron jobs fail, data becomes stale
- ❌ **No Automatic Recovery**: Manual intervention required if cron fails
- ❌ **Dependency on Cron**: System requires proper cron job setup

### Neutral

- 🔄 **Manual Refresh Still Available**: `--refresh` flags remain functional
- 🔄 **Database Remains Source of Truth**: No change to core architecture

## Implementation

### Code Changes

1. **Remove TTL Methods**:
   ```go
   // DELETE: IsCacheValid(), IsCacheHit()
   // DELETE: TTL checking in GetProjects(), GetDirs()
   ```

2. **Simplify Cache Methods**:
   ```go
   func (c *DatabaseCache) GetProjects() ([]*Project, error) {
       return c.db.GetProjects(false, "")
   }
   
   func (c *DatabaseCache) GetDirs() ([]*DirUsage, error) {
       return c.db.GetFrequentDirs(1000)
   }
   ```

3. **Remove Cache Metadata Table**:
   ```sql
   -- DELETE: cache_metadata table (no longer needed)
   ```

### Documentation Updates

1. **Update Architecture Docs**: Remove TTL references
2. **Update BDD Tests**: Remove TTL-based scenarios
3. **Update Product Vision**: Emphasize cron-based updates

## Monitoring

### Success Metrics

- **Response Time**: < 100ms for all commands
- **Cron Job Health**: Monitor cron job execution
- **Data Freshness**: Track last update timestamps

### Failure Scenarios

- **Cron Job Failure**: Alert if no updates in 24 hours
- **Stale Data**: Monitor data age vs expected refresh intervals

## Alternatives Considered

### 1. Keep TTL, Reduce Values
- **Rejected**: Still causes unnecessary scans
- **Reason**: Doesn't solve fundamental design mismatch

### 2. Hybrid Approach (TTL + Cron)
- **Rejected**: Adds complexity without benefit
- **Reason**: Cron jobs make TTL redundant

### 3. Event-Driven Updates
- **Rejected**: Over-engineering for current needs
- **Reason**: Cron jobs are sufficient and simple

## References

- [Product Vision](../README.md#product-vision)
- [Architecture Documentation](../architecture.md)
- [Caching Strategy](../caching-strategy.md)
- [BDD Tests](../bdd/project-discovery.feature)

---

**Decision by**: Development Team  
**Date**: 2024-01-15  
**Review Date**: 2024-04-15