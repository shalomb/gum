# Configuration Management Feature

## Background
Given gum is installed
And the user has a home directory with various project directories

## Scenario: Generate config stub automatically
Given no configuration file exists
And the user has Git repositories in "~/projects" and "~/code"
When I run "gum projects --refresh"
Then a config stub should be generated at "~/.config/gum/config.yaml"
And the config should contain discovered directories
And each directory should show the Git repository count
And the config should include helpful comments

## Scenario: Config stub content validation
Given a config stub has been generated
When I read the file "~/.config/gum/config.yaml"
Then it should contain:
  """
  # Gum Configuration
  # This file was auto-generated based on discovered project directories
  # Edit this file to customize which directories gum scans for Git repositories

  projects:
    - ~/projects  # 5 Git repositories
    - ~/code      # 12 Git repositories

  # Additional directories you can add:
  # - ~/code
  # - ~/dev
  # - ~/workspace
  # - ~/repos
  # - ~/repositories
  # - /path/to/any/directory

  # Note: Directories with 0 Git repositories will be ignored
  # Remove directories from this list to exclude them from scanning
  """

## Scenario: Only generate config stub when needed
Given no configuration file exists
And the user has only one directory with Git repositories
When I run "gum projects"
Then no config stub should be generated
And I should see a message about auto-discovery
And the single directory should be used directly

## Scenario: Respect existing configuration
Given a configuration file already exists at "~/.config/gum/config.yaml"
And it contains custom directories
When I run "gum projects"
Then the existing configuration should be used
And no new config stub should be generated
And I should not see auto-discovery messages

## Scenario: Handle XDG_CONFIG_HOME environment variable
Given XDG_CONFIG_HOME is set to "/custom/config"
And no configuration file exists
When I run "gum projects --refresh"
Then the config stub should be generated at "/custom/config/gum/config.yaml"
And gum should use the custom configuration directory

## Scenario: Create configuration directory if needed
Given no ".config/gum" directory exists
When I run "gum projects --refresh"
Then the directory should be created with proper permissions (0755)
And the config file should be created with proper permissions (0644)

## Scenario: Handle configuration file write errors
Given no write permissions to the configuration directory
When I run "gum projects --refresh"
Then gum should continue without generating a config stub
And I should see projects from auto-discovery
And no error should be displayed

## Scenario: Validate YAML syntax
Given a configuration file with invalid YAML syntax exists
When I run "gum projects"
Then gum should fall back to auto-discovery
And I should see an error message about invalid configuration
And the command should not crash

## Scenario: Support comments in configuration
Given a configuration file with comments exists:
  """
  # My custom project directories
  projects:
    - ~/work-projects    # Work repositories
    - ~/personal-code    # Personal projects
    # - ~/archive        # Archived projects (disabled)
  """
When I run "gum projects"
Then I should see projects from "~/work-projects" and "~/personal-code"
And "~/archive" should be ignored due to the comment

## Scenario: Handle empty configuration
Given a configuration file with no projects listed exists:
  """
  projects: []
  """
When I run "gum projects"
Then I should see no projects
And I should see a message about empty configuration

## Scenario: Configuration precedence
Given both "~/.config/gum/config.yaml" and "~/.config/projects-dirs.list" exist
And they contain different directories
When I run "gum projects"
Then the YAML configuration should take precedence
And projects from the legacy list should be ignored

## Scenario: Migrate from legacy configuration
Given a legacy "~/.config/projects-dirs.list" file exists:
  """
  ~/legacy-projects
  ~/old-repos
  """
And no YAML configuration exists
When I run "gum projects --refresh"
Then a config stub should be generated
And it should include directories from the legacy file
And the legacy file should still be respected

## Scenario: Handle relative paths in configuration
Given a configuration contains relative paths:
  """
  projects:
    - ./relative-path
    - ../parent-path
  """
When I run "gum projects"
Then the paths should be resolved relative to the current working directory
And projects should be found if the paths exist

## Scenario: Configuration file permissions
Given a configuration file has been generated
When I check the file permissions
Then the file should have 0644 permissions
And the directory should have 0755 permissions

## Scenario: Handle configuration file corruption
Given a configuration file exists but is corrupted (binary data)
When I run "gum projects"
Then gum should fall back to auto-discovery
And I should see an error message about the corrupted file
And the command should not crash

## Scenario: Configuration validation
Given a configuration file contains invalid directory paths:
  """
  projects:
    - ""                    # Empty path
    - "   "                 # Whitespace only
    - "/nonexistent/path"   # Non-existent path
  """
When I run "gum projects"
Then gum should skip invalid paths
And I should see projects from valid paths
And I should see warnings about invalid paths

## Scenario: Configuration backup
Given a configuration file exists
And I modify it to be invalid
When I run "gum projects"
Then gum should fall back to auto-discovery
And the original configuration file should remain unchanged
And no backup should be created automatically

## Scenario: Environment variable override
Given XDG_CONFIG_HOME is set to "/custom/config"
And a configuration exists at "~/.config/gum/config.yaml"
When I run "gum projects"
Then gum should use the configuration from "/custom/config/gum/config.yaml"
And ignore the default location configuration

## Scenario: Configuration file monitoring
Given a configuration file exists
And I modify it while gum is running
When I run "gum projects --refresh"
Then gum should read the updated configuration
And use the new directory list

## Scenario: Handle configuration file locks
Given a configuration file is locked by another process
When I run "gum projects --refresh"
Then gum should wait for the lock to be released
And then proceed with configuration reading
And the command should not fail

## Scenario: Configuration file encoding
Given a configuration file contains non-ASCII characters in comments:
  """
  projects:
    - ~/projets  # French: projects
    - ~/travail  # French: work
  """
When I run "gum projects"
Then gum should handle the UTF-8 encoding correctly
And the comments should be preserved
And the paths should work correctly

## Scenario: Crontab generation
When I run "gum --crontab"
Then I should see an ideal crontab configuration
And the configuration should include updatedb entries
And the configuration should include gum automation entries
And the configuration should show the correct gum executable path
And optional features should be commented out

## Scenario: Crontab generation with existing entries
Given I have existing crontab entries for gum
When I run "gum --crontab"
Then I should see a message indicating existing entries
And duplicate entries should not be suggested
And only missing entries should be shown

## Scenario: Crontab generation without updatedb
Given updatedb is not available on the system
When I run "gum --crontab"
Then the configuration should not include updatedb entries
And gum automation entries should still be shown
And a note about updatedb should be included