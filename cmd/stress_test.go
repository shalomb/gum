/*
Copyright © 2023 shalomb <s.bhooshi@gmail.com>
*/
package cmd

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/spf13/cobra"
	"github.com/shalomb/gum/internal/database"
)

var (
	stressWorkers    int
	stressDuration   time.Duration
	stressOperations int
)

// stressTestCmd represents the stress test command
var stressTestCmd = &cobra.Command{
	Use:   "stress-test",
	Short: "Run database stress test to verify concurrency and integrity",
	Long: `Run comprehensive stress tests to verify database integrity under concurrent operations.

This command simulates multiple gum processes running simultaneously and verifies:
- Data consistency under concurrent reads/writes
- Cache integrity under load
- Database integrity after operations
- No data corruption or race conditions

Examples:
  gum stress-test --workers 10 --duration 30s
  gum stress-test --workers 50 --operations 1000`,

	Run: func(cmd *cobra.Command, args []string) {
		runStressTest()
	},
}

func init() {
	rootCmd.AddCommand(stressTestCmd)

	stressTestCmd.Flags().IntVarP(&stressWorkers, "workers", "w", 10, "Number of concurrent workers")
	stressTestCmd.Flags().DurationVarP(&stressDuration, "duration", "d", 30*time.Second, "Test duration")
	stressTestCmd.Flags().IntVarP(&stressOperations, "operations", "o", 0, "Number of operations per worker (0 = unlimited)")
}

func runStressTest() {
	fmt.Printf("🧪 Starting Database Stress Test\n")
	fmt.Printf("================================\n")
	fmt.Printf("Workers: %d\n", stressWorkers)
	fmt.Printf("Duration: %v\n", stressDuration)
	fmt.Printf("Operations per worker: %d\n", stressOperations)
	fmt.Println()

	// Initialize database
	db, err := database.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Start integrity monitoring
	monitor := database.NewIntegrityMonitor(db)
	monitor.StartMonitoring(5 * time.Second)
	defer monitor.StopMonitoring()

	// Start operation tracking
	tracker := database.NewConcurrentOperationTracker()

	// Run stress test
	start := time.Now()
	results := runConcurrentOperations(db, tracker)
	duration := time.Since(start)

	// Print results
	printStressTestResults(results, duration, monitor.GetStats(), tracker.GetStats())
}

func runConcurrentOperations(db *database.Database, tracker *database.ConcurrentOperationTracker) StressTestResults {
	var wg sync.WaitGroup
	results := StressTestResults{
		Workers:     stressWorkers,
		StartTime:   time.Now(),
		Operations:  make([]OperationResult, 0, stressWorkers*100),
		Errors:      make([]error, 0),
		mu:          sync.Mutex{},
	}

	// Start workers
	for i := 0; i < stressWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			runWorker(db, tracker, workerID, &results)
		}(i)
	}

	// Wait for completion
	if stressDuration > 0 {
		// Time-based test
		time.Sleep(stressDuration)
	} else {
		// Operation-based test
		wg.Wait()
	}

	results.EndTime = time.Now()
	return results
}

func runWorker(db *database.Database, tracker *database.ConcurrentOperationTracker, workerID int, results *StressTestResults) {
	operationCount := 0
	start := time.Now()

	for {
		// Check if we should stop
		if stressDuration > 0 && time.Since(start) >= stressDuration {
			break
		}
		if stressOperations > 0 && operationCount >= stressOperations {
			break
		}

		// Choose random operation
		opType := chooseRandomOperation()
		
		// Track operation
		tracker.StartOperation(opType)
		opStart := time.Now()
		
		// Execute operation
		err := executeOperation(db, opType, workerID, operationCount)
		success := err == nil
		
		// Track completion
		tracker.EndOperation(opType, success)
		
		// Record result
		results.mu.Lock()
		results.Operations = append(results.Operations, OperationResult{
			WorkerID:    workerID,
			Operation:   opType,
			Duration:    time.Since(opStart),
			Success:     success,
			Error:       err,
			Timestamp:   time.Now(),
		})
		
		if !success {
			results.Errors = append(results.Errors, err)
		}
		results.mu.Unlock()
		
		operationCount++
		
		// Small delay to prevent overwhelming the system
		time.Sleep(1 * time.Millisecond)
	}
}

func chooseRandomOperation() string {
	operations := []string{
		"read_projects",
		"write_project",
		"read_project_dirs",
		"write_project_dir",
		"cache_get",
		"cache_set",
		"refresh_cache",
	}
	
	// Simple pseudo-random selection
	now := time.Now().UnixNano()
	return operations[now%int64(len(operations))]
}

func executeOperation(db *database.Database, opType string, workerID, opCount int) error {
	switch opType {
	case "read_projects":
		_, err := db.GetProjects(false, "")
		return err
		
	case "write_project":
		project := &database.Project{
			Path:      fmt.Sprintf("~/stress-test-%d-%d", workerID, opCount),
			Name:      fmt.Sprintf("stress-test-%d-%d", workerID, opCount),
			RemoteURL: fmt.Sprintf("https://github.com/user/stress-test-%d-%d.git", workerID, opCount),
			Branch:    "main",
		}
		return db.UpsertProject(project)
		
	case "read_project_dirs":
		_, err := db.GetProjectDirs()
		return err
		
	case "write_project_dir":
		dir := &database.ProjectDir{
			Path:        fmt.Sprintf("~/stress-dir-%d-%d", workerID, opCount),
			LastScanned: time.Now(),
			GitCount:    1,
		}
		return db.UpsertProjectDir(dir)
		
	case "cache_get":
		cache := database.NewDatabaseCache(db)
		_, err := cache.GetProjects()
		return err
		
	case "cache_set":
		cache := database.NewDatabaseCache(db)
		projects := []*database.Project{
			{
				Path:      fmt.Sprintf("~/cache-test-%d-%d", workerID, opCount),
				Name:      fmt.Sprintf("cache-test-%d-%d", workerID, opCount),
				RemoteURL: fmt.Sprintf("https://github.com/user/cache-test-%d-%d.git", workerID, opCount),
			},
		}
		return cache.SetProjects(projects)
		
	case "refresh_cache":
		cache := database.NewDatabaseCache(db)
		return cache.ClearCache("projects")
		
	default:
		return fmt.Errorf("unknown operation: %s", opType)
	}
}

type StressTestResults struct {
	Workers    int
	StartTime  time.Time
	EndTime    time.Time
	Operations []OperationResult
	Errors     []error
	mu         sync.Mutex
}

type OperationResult struct {
	WorkerID  int
	Operation string
	Duration  time.Duration
	Success   bool
	Error     error
	Timestamp time.Time
}

func printStressTestResults(results StressTestResults, duration time.Duration, integrityStats database.IntegrityStats, operationStats database.OperationStats) {
	fmt.Printf("📊 Stress Test Results\n")
	fmt.Printf("=====================\n")
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Workers: %d\n", results.Workers)
	fmt.Printf("Total Operations: %d\n", len(results.Operations))
	fmt.Printf("Successful Operations: %d\n", countSuccessfulOperations(results.Operations))
	fmt.Printf("Failed Operations: %d\n", len(results.Errors))
	fmt.Printf("Operations per second: %.2f\n", float64(len(results.Operations))/duration.Seconds())
	fmt.Printf("Success Rate: %.2f%%\n", float64(countSuccessfulOperations(results.Operations))/float64(len(results.Operations))*100)
	fmt.Println()

	// Operation breakdown
	fmt.Printf("📈 Operation Breakdown\n")
	fmt.Printf("=====================\n")
	operationCounts := make(map[string]int)
	operationDurations := make(map[string]time.Duration)
	
	for _, op := range results.Operations {
		operationCounts[op.Operation]++
		operationDurations[op.Operation] += op.Duration
	}
	
	for op, count := range operationCounts {
		avgDuration := operationDurations[op] / time.Duration(count)
		fmt.Printf("%-20s: %d operations, avg %v\n", op, count, avgDuration)
	}
	fmt.Println()

	// Integrity stats
	fmt.Printf("🔍 Integrity Monitoring\n")
	fmt.Printf("======================\n")
	fmt.Printf("Checks performed: %d\n", integrityStats.CheckCount)
	fmt.Printf("Errors detected: %d\n", integrityStats.ErrorCount)
	fmt.Printf("Error rate: %.2f%%\n", integrityStats.ErrorRate*100)
	fmt.Printf("Last check: %v\n", integrityStats.LastCheck.Format(time.RFC3339))
	fmt.Println()

	// Operation tracking
	fmt.Printf("⚡ Concurrent Operations\n")
	fmt.Printf("========================\n")
	fmt.Printf("Total active: %d\n", operationStats.TotalActive)
	fmt.Printf("Completed: %d\n", operationStats.CompletedOperations)
	fmt.Printf("Failed: %d\n", operationStats.FailedOperations)
	fmt.Printf("Active by type:\n")
	for opType, count := range operationStats.ActiveOperations {
		fmt.Printf("  %-20s: %d\n", opType, count)
	}
	fmt.Println()

	// Error summary
	if len(results.Errors) > 0 {
		fmt.Printf("❌ Error Summary\n")
		fmt.Printf("================\n")
		errorCounts := make(map[string]int)
		for _, err := range results.Errors {
			errorCounts[err.Error()]++
		}
		
		for errMsg, count := range errorCounts {
			fmt.Printf("%-50s: %d\n", errMsg, count)
		}
		fmt.Println()
	}

	// Final verdict
	successRate := float64(countSuccessfulOperations(results.Operations)) / float64(len(results.Operations)) * 100
	if successRate >= 99.0 && integrityStats.ErrorRate < 0.01 {
		fmt.Printf("✅ STRESS TEST PASSED\n")
		fmt.Printf("====================\n")
		fmt.Printf("Database integrity maintained under concurrent load\n")
		fmt.Printf("Success rate: %.2f%%\n", successRate)
		fmt.Printf("Integrity error rate: %.2f%%\n", integrityStats.ErrorRate*100)
	} else {
		fmt.Printf("❌ STRESS TEST FAILED\n")
		fmt.Printf("=====================\n")
		fmt.Printf("Issues detected under concurrent load\n")
		fmt.Printf("Success rate: %.2f%%\n", successRate)
		fmt.Printf("Integrity error rate: %.2f%%\n", integrityStats.ErrorRate*100)
		os.Exit(1)
	}
}

func countSuccessfulOperations(operations []OperationResult) int {
	count := 0
	for _, op := range operations {
		if op.Success {
			count++
		}
	}
	return count
}