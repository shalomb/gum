/*
Copyright © 2023 shalomb <s.bhooshi@gmail.com>
*/
package cmd

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	ps "github.com/mitchellh/go-ps"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/shalomb/gum/internal/database"
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
		demo, _ := cmd.Flags().GetBool("demo")
		
		if demo {
			DemoFrecencyScores()
			return
		}
		
		if clearCache {
			db, err := database.New()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to initialize database: %v\n", err)
				os.Exit(1)
			}
			defer db.Close()
			
			cache := database.NewDatabaseCache(db)
			if err := cache.ClearCache("dirs"); err != nil {
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
	dirsCmd.Flags().Bool("demo", false, "Show frecency algorithm demonstration")
}

type DirEntry struct {
	Path      string
	Score     int64
	Frequency int
	LastSeen  time.Time
}

func doUpdateDirs(format string, verbose bool, refresh bool) {
	// Initialize database
	db, err := database.New()
	if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize database: %v\n", err)
			os.Exit(1)
	}
	defer db.Close()
	
	// Initialize cache
	cache := database.NewDatabaseCache(db)
	
	var entries []*DirEntry
	
	// Always fetch current process data
	currentEntries := fetchDirs()
	
	// Try to get historical data from database first (unless refresh is requested)
	var historicalEntries []*DirEntry
	if !refresh {
		dbDirs, err := cache.GetDirs()
		if err == nil {
			// Cache hit - convert from database format
			historicalEntries = convertFromDatabaseDirs(dbDirs)
		} else {
			// Cache miss - try one-time import from legacy cwds cache
			historicalEntries = importLegacyCwdsOnce(db)
			if len(historicalEntries) == 0 {
				// No legacy cache - start with current data
				historicalEntries = currentEntries
			}
		}
	} else {
		// Force refresh - load historical data first, then merge with current
		dbDirs, err := cache.GetDirs()
		if err == nil {
			historicalEntries = convertFromDatabaseDirs(dbDirs)
		}
		historicalEntries = mergeDirectoryEntries(currentEntries, historicalEntries)
	}
	
	// Merge current with historical
	if !refresh {
		historicalEntries = mergeDirectoryEntries(currentEntries, historicalEntries)
	}
	
	// Save to database
	dbDirs := convertToDatabaseDirs(historicalEntries)
	if err := cache.SetDirs(dbDirs); err != nil {
		log.Printf("Warning: Failed to update dirs cache: %v", err)
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
			entry.Score = calculateFrecencyScore(entry.Frequency, entry.LastSeen, now)
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

// calculateFrecencyScore calculates a frecency score (frequency + recency)
// Based on Firefox's frecency algorithm with improvements
func calculateFrecencyScore(frequency int, lastSeen, now time.Time) int64 {
	age := now.Sub(lastSeen)
	ageHours := age.Hours()
	
	// Frequency component with logarithmic scaling to prevent domination
	// log(frequency + 1) gives diminishing returns for high frequencies
	frequencyScore := math.Log(float64(frequency) + 1)
	
	// Recency component with exponential decay
	// Different decay rates for different time periods
	var recencyMultiplier float64
	switch {
	case ageHours < 1:
		// Recent (last hour): no decay
		recencyMultiplier = 1.0
	case ageHours < 24:
		// Today: mild decay
		recencyMultiplier = math.Exp(-0.1 * (ageHours - 1))
	case ageHours < 168: // 1 week
		// This week: moderate decay
		recencyMultiplier = math.Exp(-0.05 * (ageHours - 24)) * 0.9
	case ageHours < 720: // 1 month
		// This month: stronger decay
		recencyMultiplier = math.Exp(-0.02 * (ageHours - 168)) * 0.5
	default:
		// Older: significant decay but never zero
		recencyMultiplier = math.Exp(-0.01 * (ageHours - 720)) * 0.1
	}
	
	// Ensure minimum score to keep everything accessible
	if recencyMultiplier < 0.01 {
		recencyMultiplier = 0.01
	}
	
	// Combine frequency and recency
	score := frequencyScore * recencyMultiplier * 1000
	
	return int64(score)
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
			existing.Score = calculateFrecencyScore(existing.Frequency, existing.LastSeen, now)
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

// convertToDatabaseDirs converts DirEntry to database.DirUsage
func convertToDatabaseDirs(entries []*DirEntry) []*database.DirUsage {
	var dbDirs []*database.DirUsage
	for _, entry := range entries {
		dbDirs = append(dbDirs, &database.DirUsage{
			Path:      entry.Path,
			Frequency: entry.Frequency,
			LastSeen:  entry.LastSeen,
		})
	}
	return dbDirs
}

// convertFromDatabaseDirs converts database.DirUsage to DirEntry
func convertFromDatabaseDirs(dbDirs []*database.DirUsage) []*DirEntry {
	var entries []*DirEntry
	now := time.Now()
	for _, dbDir := range dbDirs {
		entries = append(entries, &DirEntry{
			Path:      dbDir.Path,
			Frequency: dbDir.Frequency,
			LastSeen:  dbDir.LastSeen,
			Score:     calculateFrecencyScore(dbDir.Frequency, dbDir.LastSeen, now),
		})
	}
	return entries
}

// importLegacyCwdsOnce imports directory data from legacy cwds cache (one-time only)
// This checks if data already exists in the database before importing
func importLegacyCwdsOnce(db *database.Database) []*DirEntry {
	// Check if we already have data in the database
	existingDirs, err := db.GetFrequentDirs(1)
	if err == nil && len(existingDirs) > 0 {
		// Database already has data, skip legacy import
		return nil
	}
	
	// No data in database, try legacy import
	return importLegacyCwdsCache()
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
	
	if len(entries) > 0 {
		log.Printf("One-time import: Imported %d directories from legacy cwds cache", len(entries))
	}
	return entries
}