# Database Migration Feature

## Background
Given gum is installed and has existing JSON caches
And the SQLite database exists with basic schema
And there are existing projects and GitHub repositories

## Scenario: Migrate from JSON caches to unified database
Given JSON cache files exist at "~/.cache/gum/projects.json" and "~/.cache/gum/project-dirs.json"
And the SQLite database has the v2 schema
When I run "gum migrate-database"
Then all JSON cache data should be migrated to SQLite tables
And the JSON cache files should be backed up to "~/.cache/gum/backup/"
And I should see a message "Migration completed: X projects, Y directories migrated"
And subsequent "gum projects" commands should use the database

## Scenario: Link GitHub repositories to local projects
Given local projects exist in the database
And GitHub repositories exist in the database
When I run "gum link-github-projects"
Then projects should be linked to GitHub repos by matching clone URLs
And I should see a message "Linked X projects to GitHub repositories"
And "gum projects --with-github" should show projects with GitHub metadata

## Scenario: Unified project discovery
Given the database contains projects and project directories
When I run "gum projects"
Then projects should be loaded from the database
And the results should be consistent with previous JSON cache
And the command should complete faster than file system scanning

## Scenario: Cache consistency after migration
Given the database migration is complete
When I run "gum projects --refresh"
Then the database should be updated with fresh project data
And subsequent "gum projects" commands should return the same data
And no JSON cache files should be created

## Scenario: GitHub metadata integration
Given projects are linked to GitHub repositories
When I run "gum projects --format json"
Then the output should include GitHub metadata (stars, language, topics)
And the data should be loaded from the database, not API calls

## Scenario: Handle migration errors gracefully
Given corrupted JSON cache files exist
When I run "gum migrate-database"
Then the migration should skip corrupted files
And I should see warnings about skipped files
And valid data should still be migrated
And the command should complete successfully

## Scenario: Rollback migration if needed
Given the database migration is complete
When I run "gum migrate-database --rollback"
Then the JSON cache files should be restored from backup
And the database should be reverted to pre-migration state
And I should see a message "Migration rolled back successfully"

## Scenario: Performance improvement after migration
Given the database migration is complete
And there are 1000+ projects in the database
When I run "gum projects"
Then the command should complete in under 1 second
And the response time should be significantly faster than JSON cache

## Scenario: Concurrent access safety
Given the database migration is complete
And multiple gum processes are running
When I run "gum projects" while another process runs "gum projects --refresh"
Then both commands should complete successfully
And the database should remain consistent
And no data corruption should occur

## Scenario: Backup and restore functionality
Given the database migration is complete
When I run "gum backup-database"
Then a backup should be created at "~/.cache/gum/backup/gum.db.backup"
And the backup should contain all project and GitHub data
And I should see a message "Database backed up successfully"

## Scenario: Restore from backup
Given a database backup exists
When I run "gum restore-database --backup ~/.cache/gum/backup/gum.db.backup"
Then the current database should be replaced with the backup
And all data should be restored correctly
And I should see a message "Database restored successfully"