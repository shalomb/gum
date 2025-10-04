package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	// Build information - set during build
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
	GoVersion = runtime.Version()
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long: `Show detailed version information including:
- Application version
- Git commit hash
- Build date
- Go version
- Runtime information

Use --verbose for additional runtime details.`,
	Run: func(cmd *cobra.Command, args []string) {
		showVersion()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	
	// Add verbose flag to version command
	versionCmd.Flags().BoolVarP(&verboseMode, "verbose", "v", false, "Show verbose runtime information")
}

func showVersion() {
	fmt.Printf("gum version %s\n", Version)
	fmt.Printf("  Git commit: %s\n", GitCommit)
	fmt.Printf("  Build date: %s\n", BuildDate)
	fmt.Printf("  Go version: %s\n", GoVersion)
	fmt.Printf("  OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	
	// Show runtime info if verbose
	if verboseMode {
		fmt.Printf("  Runtime: %s\n", runtime.Version())
		fmt.Printf("  NumCPU: %d\n", runtime.NumCPU())
		fmt.Printf("  GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	}
}