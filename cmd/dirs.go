/*
Copyright © 2023 shalomb <s.bhooshi@gmail.com>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
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
	Long: `Track and list frequently accessed directories based on current processes and historical usage.
This replaces the shell script cwds-list with better performance and Go-native implementation.
Combines current process data with historical cache to provide comprehensive directory listings.`,

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
	
	// Always fetch current process data
	currentEntries := fetchDirs()
	
	// Try to get historical data from cache first (unless refresh is requested)
	var historicalEntries []*DirEntry
	if !refresh {
		if c.Get("dirs", &historicalEntries) {
			// Cache hit - use cached historical data
		} else {
			// Cache miss - try to import from legacy cwds cache
			historicalEntries = importLegacyCwdsCache()
			if len(historicalEntries) == 0 {
				// No legacy cache - start with current data
				historicalEntries = currentEntries
			}
			c.Set("dirs", historicalEntries, cache.DirsCacheTTL)
		}
	} else {
		// Force refresh - load historical data first, then merge with current
		if len(historicalEntries) == 0 {
			historicalEntries = importLegacyCwdsCache()
		}
		historicalEntries = mergeDirectoryEntries(currentEntries, historicalEntries)
		c.Set("dirs", historicalEntries, cache.DirsCacheTTL)
	}
	
	// Use merged data (current + historical)
	entries = historicalEntries

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

// mergeDirectoryEntries merges current process data with historical cache data
func mergeDirectoryEntries(current, historical []*DirEntry) []*DirEntry {
	now := time.Now()
	merged := make(map[string]*DirEntry)
	
	// Add historical entries first
	for _, entry := range historical {
		merged[entry.Path] = entry
	}
	
	// Merge current entries, updating frequency and recency
	for _, currentEntry := range current {
		if existing, exists := merged[currentEntry.Path]; exists {
			// Update existing entry with current data
			existing.Frequency++
			existing.LastSeen = now
			existing.Score = calculateScore(existing.Frequency, existing.LastSeen, now)
		} else {
			// Add new entry
			merged[currentEntry.Path] = currentEntry
		}
	}
	
	// Convert map back to slice
	var result []*DirEntry
	for _, entry := range merged {
		result = append(result, entry)
	}
	
	return result
}

// importLegacyCwdsCache imports directory data from legacy cwds cache
func importLegacyCwdsCache() []*DirEntry {
	// Use XDG_CACHE_HOME or default to ~/.cache
	cacheDir := os.Getenv("XDG_CACHE_HOME")
	if cacheDir == "" {
		home := os.Getenv("HOME")
		if home == "" {
			log.Printf("HOME environment variable not set")
			return nil
		}
		cacheDir = home + "/.cache"
	}
	cacheFile := cacheDir + "/cwds"
	
	file, err := os.Open(cacheFile)
	if err != nil {
		log.Printf("Could not open legacy cwds cache: %v", err)
		return nil
	}
	defer file.Close()
	
	var entries []*DirEntry
	now := time.Now()
	
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		
		// Parse format: "<score> <path>"
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}
		
		scoreStr := parts[0]
		path := parts[1]
		
		// Parse score
		score, err := strconv.ParseFloat(scoreStr, 64)
		if err != nil {
			continue
		}
		
		// Convert score to frequency (approximate)
		frequency := int(score)
		if frequency < 1 {
			frequency = 1
		}
		
		// Create entry
		entry := &DirEntry{
			Path:      path,
			Score:     int64(score * 1000), // Convert to int64 for consistency
			Frequency: frequency,
			LastSeen:  now.Add(-time.Duration(frequency) * time.Hour), // Approximate last seen
		}
		
		entries = append(entries, entry)
	}
	
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading legacy cwds cache: %v", err)
		return nil
	}
	
	log.Printf("Imported %d directories from legacy cwds cache", len(entries))
	return entries
}