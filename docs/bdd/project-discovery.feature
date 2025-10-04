# Project Discovery Feature

## Background
Given gum is installed and configured
And the user has Git repositories in their home directory

## Scenario: Auto-discover projects without configuration
Given no configuration file exists at "~/.config/gum/config.yaml"
And the user has Git repositories in "~/projects" and "~/oneTakeda"
When I run "gum projects"
Then I should see all Git projects from both directories
And I should see a message "gum: Auto-discovered 2 project directories"
And I should see a message "gum: Generated config stub at ~/.config/gum/config.yaml"

## Scenario: Use existing YAML configuration
Given a configuration file exists at "~/.config/gum/config.yaml"
And the file contains:
  """
  projects:
    - ~/my-projects
    - ~/work/repos
  """
And these directories contain Git repositories
When I run "gum projects"
Then I should see only projects from the configured directories
And I should not see the auto-discovery message

## Scenario: Hybrid approach with config stub generation
Given no configuration file exists
And the user has Git repositories in multiple directories
When I run "gum projects --refresh"
Then I should see all discovered projects
And a config stub should be generated at "~/.config/gum/config.yaml"
And the config stub should contain discovered directories with Git repository counts
And the config stub should include helpful comments and examples

## Scenario: Ignore directories without Git repositories
Given the user has directories "~/empty-dir" and "~/projects"
And "~/empty-dir" contains no Git repositories
And "~/projects" contains Git repositories
When I run "gum projects"
Then I should see projects only from "~/projects"
And "~/empty-dir" should not appear in the output

## Scenario: Handle permission errors gracefully
Given the user has a directory "~/restricted-dir" with no read permissions
And "~/projects" with normal permissions contains Git repositories
When I run "gum projects"
Then I should see projects from "~/projects"
And I should not see an error about "~/restricted-dir"
And the command should complete successfully

## Scenario: Support legacy projects-dirs.list
Given a legacy file exists at "~/.config/projects-dirs.list"
And the file contains:
  """
  ~/legacy-projects
  ~/old-repos
  """
And these directories contain Git repositories
When I run "gum projects"
Then I should see projects from the legacy directories
And the legacy file should be respected

## Scenario: Configuration priority order
Given both "~/.config/gum/config.yaml" and "~/.config/projects-dirs.list" exist
And they contain different directories
When I run "gum projects"
Then I should see projects from the YAML configuration
And I should not see projects from the legacy list file

## Scenario: Refresh forces re-discovery
Given projects have been cached
And new Git repositories have been added to directories
When I run "gum projects --refresh"
Then I should see the new repositories
And the cache should be updated

## Scenario: Search functionality
Given multiple projects exist with names "api-service", "web-frontend", "mobile-app"
When I run "gum projects --search api"
Then I should see only "api-service"
And the search should be case-insensitive

## Scenario: Similarity search
Given projects exist with names "web-app", "web-service", "mobile-app"
When I run "gum projects --similar web"
Then I should see "web-app" and "web-service" ranked by similarity
And "mobile-app" should not appear

## Scenario: Limit results
Given many projects exist
When I run "gum projects --limit 5"
Then I should see at most 5 projects
And the results should be ordered by relevance

## Scenario: JSON output format
Given projects exist
When I run "gum projects --format json"
Then I should see valid JSON output
And each project should have "name", "path", and "directory" fields

## Scenario: Table output format
Given projects exist
When I run "gum projects --format table"
Then I should see a table with columns "NAME", "PATH", "DIRECTORY"
And the output should be properly formatted

## Scenario: Handle tilde expansion
Given a configuration contains "~/projects"
And the user's home directory is "/home/user"
When I run "gum projects"
Then I should see projects from "/home/user/projects"
And the tilde should be properly expanded

## Scenario: Handle absolute paths
Given a configuration contains "/absolute/path/to/projects"
When I run "gum projects"
Then I should see projects from the absolute path
And the path should be used as-is

## Scenario: Cache performance
Given projects have been discovered and cached
When I run "gum projects" again
Then the results should be returned quickly
And the cache should be used

## Scenario: Cache invalidation
Given projects have been cached
And the cache TTL has expired
When I run "gum projects"
Then the projects should be re-discovered
And the cache should be updated

## Scenario: Directory frecency scoring
Given directories have been accessed with different frequencies and ages
When I run "gum dirs --verbose"
Then directories should be ranked by frecency score
And recent directories should have higher scores
And frequently used directories should have higher scores
And old directories should have lower scores

## Scenario: Legacy directory import
Given a legacy cwds cache exists with historical directory data
When I run "gum dirs" for the first time
Then the legacy data should be imported
And I should see all historical directories
And the import should be logged

## Scenario: Frecency algorithm demonstration
When I run "gum dirs --demo"
Then I should see a demonstration of the frecency algorithm
And different scenarios should show appropriate scores
And the output should be clearly formatted

## Scenario: Locate integration for project discovery
Given locate database is available and fresh
When I run "gum projects --refresh --verbose"
Then I should see locate database usage information
And the discovery should be faster than file system scanning
And I should see the number of repositories found via locate

## Scenario: Locate fallback to file system
Given locate database is unavailable or stale
When I run "gum projects --refresh --verbose"
Then I should see a warning about locate status
And the system should fall back to file system scanning
And all projects should still be discovered correctly

## Scenario: Locate database freshness warning
Given locate database is older than 24 hours
When I run "gum projects --refresh"
Then I should see a warning about database age
And the system should still use locate for bulk discovery
And file system scanning should supplement recent changes

## Scenario: Database persistence
Given projects have been discovered
When I restart gum and run "gum projects"
Then I should see the same projects
And the database should persist the information

## Scenario: Error handling for invalid configuration
Given an invalid YAML configuration file exists
When I run "gum projects"
Then I should see an error message about the invalid configuration
And gum should fall back to auto-discovery
And the command should not crash

## Scenario: Handle missing directories
Given a configuration references a non-existent directory
When I run "gum projects"
Then I should see projects from existing directories
And I should not see an error about the missing directory
And the command should complete successfully