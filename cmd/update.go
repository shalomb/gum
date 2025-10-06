// Package cmd implements our commands
package cmd

/*
Copyright © 2023 shalomb <s.bhooshi@gmail.com>
*/

import (
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update project and directory caches (alias for sync)",
	Long: `Update project and directory caches with the latest information.

This command is an alias for 'gum sync' and follows package manager conventions
like 'apt update', 'dnf update', etc. It refreshes:

- Project discovery cache (locate + filesystem scan)
- Directory frecency scoring
- GitHub repository metadata
- Local repository synchronization

Examples:
  gum update                    # Full update (equivalent to 'gum sync')
  gum update --type incremental # Incremental update only
  gum update --type repos       # Update local repositories only
  gum update --dry-run          # Show what would be updated`,
	Run: func(cmd *cobra.Command, args []string) {
		// Delegate to sync command
		syncCmd.Run(cmd, args)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Copy all flags from sync command
	updateCmd.Flags().StringVarP(&syncType, "type", "t", "full", "Sync type: full, incremental, metadata, repos")
	updateCmd.Flags().BoolVarP(&dryRun, "dry-run", "n", false, "Show what would be updated without making changes")
}