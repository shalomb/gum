# Bug Report: gum projects cache inconsistency

## Issue Summary
The `gum projects` command shows inconsistent results between cached and fresh data. The `--refresh` flag works correctly but doesn't update the cache properly, leading to stale data being returned by subsequent `gum projects list` commands.

## Environment
- **gum version**: Built from source (latest)
- **OS**: Linux 6.1.0-9-arm64
- **Configuration**: Multiple project directories configured in `~/.config/gum/config.yaml`

## Steps to Reproduce

1. **Configure multiple project directories** in `~/.config/gum/config.yaml`:
   ```yaml
   projects:
     - ~/projects/
     - ~/projects-local/
     - ~/oneTakeda/
     - ~/shalomb/
     - ~/projects/docker-images/
   ```

2. **Run `gum projects list`** - shows only 3 projects (stale cache)

3. **Run `gum projects --refresh`** - shows hundreds of projects (correct)

4. **Run `gum projects list` again** - still shows only 3 projects (cache not updated)

5. **Run `gum projects --clear-cache`** - clears cache

6. **Run `gum projects list`** - now shows all projects correctly

## Expected Behavior
- `gum projects --refresh` should update the cache so that subsequent `gum projects list` commands return the same data
- Cache should be automatically refreshed when configuration changes
- No need to manually clear cache to see updated results

## Actual Behavior
- `gum projects --refresh` scans correctly but doesn't update the cache
- `gum projects list` continues to return stale cached data
- Must manually run `gum projects --clear-cache` to see updated results

## Impact
- **High**: Users see incomplete project lists in tmuxie and other tools that depend on `gum projects`
- **Confusing**: Refresh command appears to work but doesn't persist changes
- **Workflow disruption**: Requires manual cache clearing to see updated project lists

## Configuration Details
```yaml
# ~/.config/gum/config.yaml
projects:
  - ~/projects/
  - ~/projects-local/
  - ~/oneTakeda/
  - ~/shalomb/
  - ~/projects/docker-images/
```

## Debug Information
- **Cache location**: Not determined (would be helpful to document)
- **Cache format**: Not determined (would be helpful to document)
- **Refresh mechanism**: Appears to scan but doesn't write to cache

## Suggested Fix
1. Ensure `--refresh` flag updates the cache after scanning
2. Add cache invalidation when configuration changes
3. Consider automatic cache refresh on configuration changes
4. Document cache location and format for debugging

## Workaround
Run `gum projects --clear-cache` followed by `gum projects list` to see updated results.

---
**Reported by**: AI Assistant (cursor-agent)
**Date**: $(date)
**gum version**: $(gum --version)