# Directory Management Feature

## Background
Given gum is installed and configured
And the user has accessed various directories over time

## Scenario: List frequently accessed directories
Given directories have been accessed with different frequencies
When I run "gum dirs"
Then I should see directories ranked by frecency score
And the most relevant directories should appear first
And all accessed directories should be included

## Scenario: Show frecency scores
Given directories have been accessed with different patterns
When I run "gum dirs --verbose"
Then I should see frecency scores for each directory
And scores should reflect frequency and recency
And directories should be sorted by score (highest first)

## Scenario: Demonstrate frecency algorithm
When I run "gum dirs --demo"
Then I should see a demonstration table
And different scenarios should show appropriate scores
And recent directories should have higher scores than old ones
And frequent directories should have higher scores than rare ones

## Scenario: Import legacy cwds data
Given a legacy cwds cache file exists at "~/.cache/cwds"
And it contains historical directory data
When I run "gum dirs" for the first time
Then the legacy data should be imported
And I should see all historical directories
And the import should be logged with count
And subsequent runs should use cached data

## Scenario: Refresh with current processes
Given historical directory data exists in cache
And current processes are running in different directories
When I run "gum dirs --refresh"
Then current process directories should be merged with historical data
And frequency counts should be updated
And scores should be recalculated
And the cache should be updated

## Scenario: Frecency scoring properties
Given directories with different access patterns
When I run "gum dirs --verbose"
Then higher frequency should give higher scores (for same age)
And more recent access should give higher scores (for same frequency)
And logarithmic scaling should prevent high-frequency domination
And minimum score should ensure all directories remain accessible

## Scenario: Multi-tier decay behavior
Given directories accessed at different times
When I run "gum dirs --verbose"
Then directories accessed in the last hour should have no decay
And directories accessed today should have mild decay
And directories accessed this week should have moderate decay
And directories accessed this month should have stronger decay
And very old directories should have significant decay but never zero

## Scenario: Cache persistence
Given directories have been tracked and cached
When I restart gum and run "gum dirs"
Then I should see the same directories
And scores should be preserved
And historical data should persist

## Scenario: Clear directory cache
Given directory cache contains historical data
When I run "gum dirs --clear-cache"
Then the cache should be cleared
And I should see a confirmation message
And subsequent runs should start fresh

## Scenario: Handle missing legacy cache
Given no legacy cwds cache exists
When I run "gum dirs"
Then gum should start with current process data
And no import should be attempted
And the command should complete successfully

## Scenario: XDG compliance for cache
Given XDG_CACHE_HOME is set to "/custom/cache"
And a legacy cwds cache exists at "/custom/cache/cwds"
When I run "gum dirs"
Then gum should import from the custom cache location
And respect the XDG environment variable

## Scenario: Output format variations
Given directories have been tracked
When I run "gum dirs --format simple"
Then I should see one directory per line
And no additional information should be shown

When I run "gum dirs --format json"
Then I should see valid JSON output
And each directory should have path, score, frequency, and last_seen fields

When I run "gum dirs --format fzf"
Then I should see fzf-compatible output
And directories should be formatted for fuzzy finding

## Scenario: Concurrent access handling
Given multiple gum processes are running simultaneously
When I run "gum dirs" from multiple terminals
Then all processes should work correctly
And cache operations should be safe
And no corruption should occur

## Scenario: Large directory collections
Given thousands of directories have been accessed
When I run "gum dirs"
Then the command should complete within reasonable time
And memory usage should remain stable
And all directories should be accessible

## Scenario: Directory path normalization
Given directories have been accessed with different path formats
When I run "gum dirs"
Then paths should be normalized consistently
And ~ notation should be used for home directory paths
And duplicates should be handled correctly

## Scenario: Frequency tracking accuracy
Given a directory is accessed multiple times
When I run "gum dirs --refresh" multiple times
Then the frequency count should increase
And the last_seen timestamp should be updated
And the score should reflect the increased frequency

## Scenario: Score calculation consistency
Given the same directory access pattern
When I run "gum dirs --verbose" multiple times
Then scores should be consistent
And only last_seen timestamps should change
And frequency counts should remain stable