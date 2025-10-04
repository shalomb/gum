// Package cmd implements our commands
package cmd

/*
Copyright © 2023 shalomb <s.bhooshi@gmail.com>
*/

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	// "github.com/spf13/viper"
)

var (
	// Debug Enable debugging
	Debug bool
	// Crontab Generate crontab configuration
	Crontab bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gum",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		if Crontab {
			generateCrontab()
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.gum.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().BoolP("clear-all-cache", "", false, "Clear all cache and exit")
	rootCmd.Flags().BoolVarP(&Crontab, "crontab", "", false, "Generate ideal crontab configuration")

  // rootCmd.PersistentFlags().BoolVarP(&Debug, "debug", "d", false, "Display debugging output in the console. (default: false)")
	// viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

// generateCrontab generates an ideal crontab configuration for gum
func generateCrontab() {
	gumPath, err := exec.LookPath("gum")
	if err != nil {
		// Fallback to current executable path
		gumPath, err = os.Executable()
		if err != nil {
			gumPath = "/usr/local/bin/gum" // Default fallback
		}
	}

	// Check if updatedb is available
	hasUpdatedb := false
	if _, err := exec.LookPath("updatedb"); err == nil {
		hasUpdatedb = true
	}

	// Check current crontab
	currentCrontab := getCurrentCrontab()

	fmt.Println("# Ideal crontab configuration for gum")
	fmt.Println("# Generated on:", getCurrentTime())
	fmt.Println("#")
	fmt.Println("# To install: crontab -e")
	fmt.Println("# Copy the lines below (uncomment as needed)")
	fmt.Println()

	if hasUpdatedb {
		if !strings.Contains(currentCrontab, "updatedb") {
			fmt.Println("# Update locate database daily at 2 AM")
			fmt.Println("0 2 * * * /usr/bin/updatedb")
		} else {
			fmt.Println("# updatedb already configured in crontab")
		}
		fmt.Println()
	}

	if !strings.Contains(currentCrontab, "gum projects") {
		fmt.Println("# Refresh project cache daily at 3 AM (after updatedb)")
		fmt.Printf("0 3 * * * %s projects --refresh\n", gumPath)
	} else {
		fmt.Println("# gum projects already configured in crontab")
	}
	fmt.Println()

	if !strings.Contains(currentCrontab, "gum dirs") {
		fmt.Println("# Refresh directory cache every 2 hours")
		fmt.Printf("0 */2 * * * %s dirs --refresh\n", gumPath)
	} else {
		fmt.Println("# gum dirs already configured in crontab")
	}
	fmt.Println()

	if !strings.Contains(currentCrontab, "gum update") {
		fmt.Println("# Update databases every 6 hours")
		fmt.Printf("0 */6 * * * %s update\n", gumPath)
	} else {
		fmt.Println("# gum update already configured in crontab")
	}
	fmt.Println()

	if !strings.Contains(currentCrontab, "gum sync") {
		fmt.Println("# Sync GitHub repository metadata daily at 5 AM")
		fmt.Printf("0 5 * * * %s sync --type full\n", gumPath)
	} else {
		fmt.Println("# gum sync already configured in crontab")
	}
	fmt.Println()

	fmt.Println("# Optional: Data export for external tools")
	fmt.Printf("# 0 4 * * * %s projects --format json > ~/.cache/gum/projects-$(date +\\%%Y\\%%m\\%%d).json\n", gumPath)
	fmt.Printf("# 0 4 * * * %s dirs --format json > ~/.cache/gum/dirs-$(date +\\%%Y\\%%m\\%%d).json\n", gumPath)
	fmt.Println()

	fmt.Println("# Optional: Weekly cache cleanup")
	fmt.Printf("# 0 5 * * 0 %s projects --clear-cache\n", gumPath)
	fmt.Printf("# 0 5 * * 0 %s dirs --clear-cache\n", gumPath)
	fmt.Println()

	fmt.Println("# Optional: Health monitoring")
	fmt.Printf("# 0 6 * * * %s projects --format simple | wc -l > ~/.cache/gum/project-count-$(date +\\%%Y\\%%m\\%%d).txt\n", gumPath)
}

// getCurrentCrontab returns the current user's crontab
func getCurrentCrontab() string {
	cmd := exec.Command("crontab", "-l")
	output, err := cmd.Output()
	if err != nil {
		return "" // No crontab or error
	}
	return string(output)
}

// getCurrentTime returns the current time in a readable format
func getCurrentTime() string {
	cmd := exec.Command("date")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}


