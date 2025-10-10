/*
Copyright © 2023 shalomb <s.bhooshi@gmail.com>
*/
package database

import (
	"fmt"
	"sync"
	"time"
)

// IntegrityMonitor monitors database integrity and concurrency
type IntegrityMonitor struct {
	db           *Database
	lastCheck    time.Time
	checkCount   int64
	errorCount   int64
	mu           sync.RWMutex
	monitoring   bool
	stopChan     chan bool
}

// NewIntegrityMonitor creates a new integrity monitor
func NewIntegrityMonitor(db *Database) *IntegrityMonitor {
	return &IntegrityMonitor{
		db:       db,
		stopChan: make(chan bool),
	}
}

// StartMonitoring starts continuous integrity monitoring
func (im *IntegrityMonitor) StartMonitoring(interval time.Duration) {
	im.mu.Lock()
	defer im.mu.Unlock()
	
	if im.monitoring {
		return // Already monitoring
	}
	
	im.monitoring = true
	
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		
		for {
			select {
			case <-ticker.C:
				im.performIntegrityCheck()
			case <-im.stopChan:
				return
			}
		}
	}()
}

// StopMonitoring stops the integrity monitoring
func (im *IntegrityMonitor) StopMonitoring() {
	im.mu.Lock()
	defer im.mu.Unlock()
	
	if !im.monitoring {
		return
	}
	
	im.monitoring = false
	close(im.stopChan)
}

// performIntegrityCheck performs a comprehensive integrity check
func (im *IntegrityMonitor) performIntegrityCheck() {
	im.mu.Lock()
	defer im.mu.Unlock()
	
	im.checkCount++
	im.lastCheck = time.Now()
	
	// Check database integrity
	if err := im.checkDatabaseIntegrity(); err != nil {
		im.errorCount++
		fmt.Printf("Database integrity check failed: %v\n", err)
		return
	}
	
	// Check for orphaned records
	if err := im.checkOrphanedRecords(); err != nil {
		im.errorCount++
		fmt.Printf("Orphaned records check failed: %v\n", err)
		return
	}
	
	// Check for duplicate records
	if err := im.checkDuplicateRecords(); err != nil {
		im.errorCount++
		fmt.Printf("Duplicate records check failed: %v\n", err)
		return
	}
	
	// Check foreign key constraints
	if err := im.checkForeignKeyConstraints(); err != nil {
		im.errorCount++
		fmt.Printf("Foreign key constraints check failed: %v\n", err)
		return
	}
}

// checkDatabaseIntegrity performs SQLite integrity check
func (im *IntegrityMonitor) checkDatabaseIntegrity() error {
	var result string
	err := im.db.db.QueryRow("PRAGMA integrity_check").Scan(&result)
	if err != nil {
		return fmt.Errorf("failed to run integrity check: %w", err)
	}
	
	if result != "ok" {
		return fmt.Errorf("database integrity check failed: %s", result)
	}
	
	return nil
}

// checkOrphanedRecords checks for orphaned records
func (im *IntegrityMonitor) checkOrphanedRecords() error {
	// Check for projects with invalid github_repo_id
	query := `
		SELECT COUNT(*) 
		FROM projects p 
		LEFT JOIN github_repos gr ON p.github_repo_id = gr.id 
		WHERE p.github_repo_id IS NOT NULL AND gr.id IS NULL
	`
	
	var count int
	err := im.db.db.QueryRow(query).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check orphaned projects: %w", err)
	}
	
	if count > 0 {
		return fmt.Errorf("found %d orphaned projects with invalid github_repo_id", count)
	}
	
	return nil
}

// checkDuplicateRecords checks for duplicate records
func (im *IntegrityMonitor) checkDuplicateRecords() error {
	// Check for duplicate projects by path
	query := `
		SELECT path, COUNT(*) 
		FROM projects 
		GROUP BY path 
		HAVING COUNT(*) > 1
	`
	
	rows, err := im.db.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to check duplicate projects: %w", err)
	}
	defer rows.Close()
	
	var duplicates []string
	for rows.Next() {
		var path string
		var count int
		if err := rows.Scan(&path, &count); err != nil {
			return fmt.Errorf("failed to scan duplicate record: %w", err)
		}
		duplicates = append(duplicates, fmt.Sprintf("%s (%d copies)", path, count))
	}
	
	if len(duplicates) > 0 {
		return fmt.Errorf("found duplicate projects: %v", duplicates)
	}
	
	return nil
}

// checkForeignKeyConstraints checks foreign key constraints
func (im *IntegrityMonitor) checkForeignKeyConstraints() error {
	// Enable foreign key checking
	_, err := im.db.db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}
	
	// Check foreign key constraints
	_, err = im.db.db.Exec("PRAGMA foreign_key_check")
	if err != nil {
		return fmt.Errorf("foreign key constraint violation: %w", err)
	}
	
	return nil
}

// GetStats returns monitoring statistics
func (im *IntegrityMonitor) GetStats() IntegrityStats {
	im.mu.RLock()
	defer im.mu.RUnlock()
	
	return IntegrityStats{
		LastCheck:    im.lastCheck,
		CheckCount:   im.checkCount,
		ErrorCount:   im.errorCount,
		IsMonitoring: im.monitoring,
		ErrorRate:    float64(im.errorCount) / float64(im.checkCount),
	}
}

// IntegrityStats represents integrity monitoring statistics
type IntegrityStats struct {
	LastCheck    time.Time `json:"last_check"`
	CheckCount   int64     `json:"check_count"`
	ErrorCount   int64     `json:"error_count"`
	IsMonitoring bool      `json:"is_monitoring"`
	ErrorRate    float64   `json:"error_rate"`
}

// ConcurrentOperationTracker tracks concurrent operations
type ConcurrentOperationTracker struct {
	activeOps    map[string]int
	completedOps int64
	failedOps    int64
	mu           sync.RWMutex
}

// NewConcurrentOperationTracker creates a new operation tracker
func NewConcurrentOperationTracker() *ConcurrentOperationTracker {
	return &ConcurrentOperationTracker{
		activeOps: make(map[string]int),
	}
}

// StartOperation marks the start of an operation
func (cot *ConcurrentOperationTracker) StartOperation(opType string) {
	cot.mu.Lock()
	defer cot.mu.Unlock()
	
	cot.activeOps[opType]++
}

// EndOperation marks the end of an operation
func (cot *ConcurrentOperationTracker) EndOperation(opType string, success bool) {
	cot.mu.Lock()
	defer cot.mu.Unlock()
	
	if cot.activeOps[opType] > 0 {
		cot.activeOps[opType]--
	}
	
	if success {
		cot.completedOps++
	} else {
		cot.failedOps++
	}
}

// GetStats returns operation tracking statistics
func (cot *ConcurrentOperationTracker) GetStats() OperationStats {
	cot.mu.RLock()
	defer cot.mu.RUnlock()
	
	activeCount := 0
	for _, count := range cot.activeOps {
		activeCount += count
	}
	
	return OperationStats{
		ActiveOperations:    cot.activeOps,
		CompletedOperations: cot.completedOps,
		FailedOperations:    cot.failedOps,
		TotalActive:         activeCount,
	}
}

// OperationStats represents operation tracking statistics
type OperationStats struct {
	ActiveOperations    map[string]int `json:"active_operations"`
	CompletedOperations int64          `json:"completed_operations"`
	FailedOperations    int64          `json:"failed_operations"`
	TotalActive         int             `json:"total_active"`
}

// DatabaseLocker provides database-level locking
type DatabaseLocker struct {
	db *Database
}

// NewDatabaseLocker creates a new database locker
func NewDatabaseLocker(db *Database) *DatabaseLocker {
	return &DatabaseLocker{db: db}
}

// WithLock executes a function with a database lock
func (dl *DatabaseLocker) WithLock(lockName string, timeout time.Duration, fn func() error) error {
	// Get exclusive lock
	query := "SELECT 1 FROM sqlite_master WHERE type='table' AND name='lock_table'"
	var exists int
	err := dl.db.db.QueryRow(query).Scan(&exists)
	if err != nil {
		// Create lock table if it doesn't exist
		_, err = dl.db.db.Exec(`
			CREATE TABLE IF NOT EXISTS lock_table (
				lock_name TEXT PRIMARY KEY,
				locked_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				locked_by TEXT
			)
		`)
		if err != nil {
			return fmt.Errorf("failed to create lock table: %w", err)
		}
	}
	
	// Try to acquire lock
	start := time.Now()
	for time.Since(start) < timeout {
		_, err = dl.db.db.Exec(`
			INSERT INTO lock_table (lock_name, locked_by) 
			VALUES (?, ?)
		`, lockName, fmt.Sprintf("process-%d", time.Now().UnixNano()))
		
		if err == nil {
			// Lock acquired successfully
			defer func() {
				// Release lock
				dl.db.db.Exec("DELETE FROM lock_table WHERE lock_name = ?", lockName)
			}()
			
			return fn()
		}
		
		// Wait a bit before retrying
		time.Sleep(10 * time.Millisecond)
	}
	
	return fmt.Errorf("failed to acquire lock %s within timeout %v", lockName, timeout)
}