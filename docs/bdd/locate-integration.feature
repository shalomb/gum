# Locate Integration Feature

## Background
Given gum is installed and configured
And the system has a locate database available

## Scenario: Automatic locate usage
Given locate database is available and fresh
When I run "gum projects --refresh"
Then gum should automatically use locate for discovery
And the discovery should be faster than file system scanning
And all projects should be found correctly

## Scenario: Locate database freshness check
Given locate database is available but older than 24 hours
When I run "gum projects --refresh"
Then gum should show a warning about database age
And gum should still use locate for bulk discovery
And gum should supplement with file system scanning for recent changes

## Scenario: Locate unavailable fallback
Given locate command is not available on the system
When I run "gum projects --refresh"
Then gum should fall back to file system scanning
And all projects should still be discovered correctly
And no errors should occur

## Scenario: Locate database corruption handling
Given locate database exists but is corrupted
When I run "gum projects --refresh"
Then gum should detect the corruption
And gum should fall back to file system scanning
And a warning should be shown to the user

## Scenario: Verbose locate output
Given locate database is available
When I run "gum projects --refresh --verbose"
Then I should see locate database usage information
And I should see the number of repositories found via locate
And I should see the number of project directories discovered

## Scenario: Cross-platform locate support
Given the system uses plocate implementation
When I run "gum projects --refresh"
Then gum should detect and use plocate
And the discovery should work correctly

Given the system uses GNU locate implementation
When I run "gum projects --refresh"
Then gum should detect and use GNU locate
And the discovery should work correctly

Given the system uses BSD locate implementation
When I run "gum projects --refresh"
Then gum should detect and use BSD locate
And the discovery should work correctly

## Scenario: Locate performance comparison
Given a large project directory with 1000+ repositories
When I run "gum projects --refresh" with locate available
Then the discovery should complete in under 1 second
And all repositories should be found

When I run "gum projects --refresh" without locate
Then the discovery should take significantly longer
But all repositories should still be found

## Scenario: Locate database path detection
Given locate database is in a non-standard location
When I run "gum projects --refresh"
Then gum should detect the database path
And use the correct database for discovery

## Scenario: Locate regex pattern matching
Given locate database contains various file types
When I run "gum projects --refresh"
Then gum should use correct regex pattern for .git directories
And only git repositories should be found
And false positives should be filtered out

## Scenario: Locate output parsing
Given locate command returns various output formats
When I run "gum projects --refresh"
Then gum should parse the output correctly
And extract repository paths accurately
And handle edge cases in path formatting

## Scenario: Locate integration with YAML config
Given both locate database and YAML config are available
When I run "gum projects --refresh"
Then gum should prioritize YAML config
And skip locate-based auto-discovery
And use the configured directories

## Scenario: Locate integration with auto-discovery
Given no YAML config exists
And locate database is available
When I run "gum projects --refresh"
Then gum should use locate for auto-discovery
And find common project directories
And generate appropriate config stub

## Scenario: Locate database monitoring
Given locate database is being used
When I run "gum projects --verbose"
Then I should see database last updated timestamp
And I should see database path information
And I should see freshness status

## Scenario: Locate error handling
Given locate command fails with permission error
When I run "gum projects --refresh"
Then gum should handle the error gracefully
And fall back to file system scanning
And show appropriate error message

Given locate command times out
When I run "gum projects --refresh"
Then gum should handle the timeout
And fall back to file system scanning
And continue operation normally

## Scenario: Locate integration with caching
Given locate has been used for discovery
When I run "gum projects" again
Then gum should use cached results
And not call locate again
And maintain performance benefits

When I run "gum projects --refresh"
Then gum should call locate again
And update cached results
And show fresh discovery information