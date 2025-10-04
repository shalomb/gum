# GitHub Repository Sync Feature

## Background
Given gum is installed and configured
And GitHub CLI is authenticated with appropriate scopes
And the user has access to GitHub repositories

## Scenario: Full repository sync
Given the user has access to multiple GitHub repositories
When I run "gum sync --type full"
Then I should see sync progress messages
And all accessible repositories should be discovered
And repository metadata should be stored in the database
And the sync should complete successfully

## Scenario: Incremental repository sync
Given repository metadata has been synced previously
And some repositories have been updated since last sync
When I run "gum sync --type incremental"
Then only repositories older than 24 hours should be updated
And the sync should complete faster than full sync
And no duplicate data should be created

## Scenario: Dry-run sync
When I run "gum sync --dry-run --type full"
Then I should see what would be synced
And no actual changes should be made to the database
And the output should show repository names and descriptions

## Scenario: GitHub authentication required
Given GitHub CLI is not authenticated
When I run "gum sync --type full"
Then I should see an authentication error message
And the sync should fail gracefully
And I should be instructed to run "gh auth login"

## Scenario: Repository metadata extraction
Given a GitHub repository with rich metadata
When I run "gum sync --type full"
Then the repository topics should be captured
And the repository language should be detected
And the repository description should be stored
And activity metrics should be recorded

## Scenario: Cross-organization repository discovery
Given the user has access to multiple organizations
When I run "gum sync --type full"
Then repositories from all accessible organizations should be discovered
And organization-specific metadata should be preserved
And the total count should include all accessible repositories

## Scenario: Sync status tracking
When I run "gum sync --type full"
Then sync status should be recorded in the database
And progress should be tracked during execution
And completion status should be recorded
And error messages should be captured if sync fails

## Scenario: Database caching
Given repository metadata has been synced
When I run "gum sync --type incremental"
Then cached data should be used when appropriate
And only stale repositories should be updated
And API calls should be minimized

## Scenario: Rate limiting handling
Given GitHub API rate limits are encountered
When I run "gum sync --type full"
Then the sync should handle rate limits gracefully
And retry logic should be implemented
And the sync should continue after rate limit reset

## Scenario: Large repository collection
Given the user has access to 1000+ repositories
When I run "gum sync --type full"
Then the sync should process repositories in batches
And progress should be reported every 100 repositories
And the sync should complete without timeout
And all repositories should be processed

## Scenario: Repository property detection
Given repositories with various properties
When I run "gum sync --type full"
Then private repositories should be identified
And archived repositories should be marked
And fork repositories should be detected
And template repositories should be recognized

## Scenario: Activity metrics tracking
Given repositories with different activity levels
When I run "gum sync --type full"
Then star counts should be recorded
And fork counts should be captured
And open issue counts should be stored
And last pushed timestamps should be recorded

## Scenario: Language detection
Given repositories written in different languages
When I run "gum sync --type full"
Then the primary language should be detected
And language distribution should be captured
And repositories should be categorized by language

## Scenario: Topic-based categorization
Given repositories with GitHub topics
When I run "gum sync --type full"
Then repository topics should be extracted
And topics should be stored as JSON arrays
And repositories should be searchable by topic

## Scenario: Sync error recovery
Given a network error occurs during sync
When I run "gum sync --type full"
Then the sync should handle the error gracefully
And partial progress should be preserved
And the error should be logged appropriately
And the sync should be retryable

## Scenario: Crontab integration
When I run "gum --crontab"
Then GitHub sync should be included in the crontab output
And the sync should be scheduled for daily execution
And the sync type should be appropriate for automation

## Scenario: Database integrity
Given repository metadata has been synced
When I query the database directly
Then all repository records should be valid
And foreign key constraints should be satisfied
And data types should be correct
And indexes should be properly created