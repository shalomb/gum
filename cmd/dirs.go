/*
Copyright © 2023 shalomb <s.bhooshi@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	ps "github.com/mitchellh/go-ps"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/shalomb/gum/internal/cache"
)

// dirsCmd represents the dirs command
var dirsCmd = &cobra.Command{
	Use:   "dirs",
	Short: "List frequently accessed directories",
	Long: `Track and list frequently accessed directories based on running processes.
This replaces the shell script cwds-list with better performance and Go-native implementation.`,

	Run: func(cmd *cobra.Command, args []string) {
		format, _ := cmd.Flags().GetString("format")
		verbose, _ := cmd.Flags().GetBool("verbose")
		refresh, _ := cmd.Flags().GetBool("refresh")
		clearCache, _ := cmd.Flags().GetBool("clear-cache")
		
		if clearCache {
			c := cache.New()
			if err := c.Clear("dirs"); err != nil {
				fmt.Fprintf(os.Stderr, "Error clearing cache: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Dirs cache cleared")
			return
		}
		
		doUpdateDirs(format, verbose, refresh)
	},
}

func init() {
	rootCmd.AddCommand(dirsCmd)

	// Add flags for different output formats
	dirsCmd.Flags().StringP("format", "f", "default", "Output format: default, fzf, json, simple")
	dirsCmd.Flags().BoolP("verbose", "v", false, "Show verbose output with scores")
	dirsCmd.Flags().BoolP("refresh", "r", false, "Force refresh cache")
	dirsCmd.Flags().BoolP("clear-cache", "", false, "Clear cache and exit")
}

type DirEntry struct {
	Path      string
	Score     int64
	Frequency int
	LastSeen  time.Time
}

func doUpdateDirs(format string, verbose bool, refresh bool) {
	c := cache.New()
	var entries []*DirEntry
	
	// Try to get from cache first (unless refresh is requested)
	if !refresh {
		if c.Get("dirs", &entries) {
			// Cache hit - use cached data
		} else {
			// Cache miss - fetch fresh data
			entries = fetchDirs()
			c.Set("dirs", entries, cache.DirsCacheTTL)
		}
	} else {
		// Force refresh - fetch fresh data
		entries = fetchDirs()
		c.Set("dirs", entries, cache.DirsCacheTTL)
	}

	// Sort by score (highest first)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Score > entries[j].Score
	})

	// Output based on format
	switch format {
	case "fzf":
		outputFzfFormat(entries, verbose)
	case "json":
		outputJsonFormat(entries)
	case "simple":
		outputSimpleFormat(entries)
	default:
		outputDefaultFormat(entries, verbose)
	}
}

func fetchDirs() []*DirEntry {
	pslist, err := ps.Processes()
	if err != nil {
		log.Printf("error listing processes: %v", err)
		return nil
	}

	dirs := make(map[string]*DirEntry)
	now := time.Now()

	for _, pid := range pslist {
		cwd := fmt.Sprintf("/proc/%d/cwd", pid.Pid())
		dir, err := os.Readlink(cwd)
		if err != nil || len(dir) == 0 {
			continue
		}

		// Filter to home directory only
		home := os.Getenv("HOME")
		if !strings.HasPrefix(dir, home) {
			continue
		}

		// Convert to ~ notation
		displayPath := "~" + dir[len(home):]

		if entry, exists := dirs[displayPath]; exists {
			entry.Frequency++
			entry.LastSeen = now
			// Update score with recency weighting
			entry.Score = calculateScore(entry.Frequency, entry.LastSeen, now)
		} else {
			dirs[displayPath] = &DirEntry{
				Path:      displayPath,
				Score:     1,
				Frequency: 1,
				LastSeen:  now,
			}
		}
	}

	// Convert to slice
	var entries []*DirEntry
	for _, entry := range dirs {
		entries = append(entries, entry)
	}

	return entries
}

func calculateScore(frequency int, lastSeen, now time.Time) int64 {
	// Simple scoring: frequency * recency factor
	hoursAgo := now.Sub(lastSeen).Hours()
	recencyFactor := 1.0 / (1.0 + hoursAgo/24.0) // Decay over days
	return int64(float64(frequency) * recencyFactor * 1000)
}

func outputDefaultFormat(entries []*DirEntry, verbose bool) {
	for _, entry := range entries {
		if verbose {
			fmt.Printf("%d\t%s\n", entry.Score, entry.Path)
		} else {
			fmt.Printf("%s\n", entry.Path)
		}
	}
}

func outputFzfFormat(entries []*DirEntry, verbose bool) {
	for _, entry := range entries {
		if verbose {
			// Format for fzf with visual indicators
			fmt.Printf("%-50s %-15s %d processes\n", 
				entry.Path, "[frequent]", entry.Frequency)
		} else {
			fmt.Printf("%s\n", entry.Path)
		}
	}
}

func outputSimpleFormat(entries []*DirEntry) {
	for _, entry := range entries {
		fmt.Printf("%s\n", entry.Path)
	}
}

func outputJsonFormat(entries []*DirEntry) {
	fmt.Printf("[\n")
	for i, entry := range entries {
		fmt.Printf("  {\n")
		fmt.Printf("    \"path\": \"%s\",\n", entry.Path)
		fmt.Printf("    \"score\": %d,\n", entry.Score)
		fmt.Printf("    \"frequency\": %d,\n", entry.Frequency)
		fmt.Printf("    \"last_seen\": \"%s\"\n", entry.LastSeen.Format(time.RFC3339))
		if i < len(entries)-1 {
			fmt.Printf("  },\n")
		} else {
			fmt.Printf("  }\n")
		}
	}
	fmt.Printf("]\n")
}