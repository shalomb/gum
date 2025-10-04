package cache

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	// Create a temporary directory for test cache
	tempDir := t.TempDir()
	originalCacheHome := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if originalCacheHome != "" {
			os.Setenv("XDG_CACHE_HOME", originalCacheHome)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
	}()

	// Set test cache directory
	os.Setenv("XDG_CACHE_HOME", tempDir)

	cache := New()
	if cache == nil {
		t.Fatal("New() returned nil cache")
	}
}

func TestCacheOperations(t *testing.T) {
	// Create test cache
	tempDir := t.TempDir()
	originalCacheHome := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if originalCacheHome != "" {
			os.Setenv("XDG_CACHE_HOME", originalCacheHome)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
	}()

	os.Setenv("XDG_CACHE_HOME", tempDir)

	cache := New()

	t.Run("Set and Get", func(t *testing.T) {
		testData := []string{"test", "data", "for", "cache"}
		
		// Set data
		err := cache.Set("test-key", testData, 5*time.Minute)
		if err != nil {
			t.Errorf("Set failed: %v", err)
		}

		// Get data
		var retrievedData []string
		success := cache.Get("test-key", &retrievedData)
		if !success {
			t.Error("Get failed to retrieve data")
		}

		if len(retrievedData) != len(testData) {
			t.Errorf("Expected %d items, got %d", len(testData), len(retrievedData))
		}

		for i, item := range testData {
			if retrievedData[i] != item {
				t.Errorf("Expected item %d to be '%s', got '%s'", i, item, retrievedData[i])
			}
		}
	})

	t.Run("Get Non-existent Key", func(t *testing.T) {
		var data []string
		success := cache.Get("non-existent-key", &data)
		if success {
			t.Error("Get should return false for non-existent key")
		}
	})

	t.Run("TTL Expiration", func(t *testing.T) {
		testData := "expiring data"
		
		// Set data with very short TTL
		err := cache.Set("expiring-key", testData, 1*time.Millisecond)
		if err != nil {
			t.Errorf("Set failed: %v", err)
		}

		// Wait for expiration
		time.Sleep(10 * time.Millisecond)

		// Try to get expired data
		var retrievedData string
		success := cache.Get("expiring-key", &retrievedData)
		if success {
			t.Error("Get should return false for expired data")
		}
	})

	t.Run("Clear", func(t *testing.T) {
		testData := "data to clear"
		
		// Set data
		err := cache.Set("clear-key", testData, 5*time.Minute)
		if err != nil {
			t.Errorf("Set failed: %v", err)
		}

		// Clear specific key
		err = cache.Clear("clear-key")
		if err != nil {
			t.Errorf("Clear failed: %v", err)
		}

		// Verify data is cleared
		var retrievedData string
		success := cache.Get("clear-key", &retrievedData)
		if success {
			t.Error("Get should return false after clear")
		}
	})

	t.Run("ClearAll", func(t *testing.T) {
		// Set multiple keys
		keys := []string{"key1", "key2", "key3"}
		for _, key := range keys {
			err := cache.Set(key, "test data", 5*time.Minute)
			if err != nil {
				t.Errorf("Set failed for key %s: %v", key, err)
			}
		}

		// Clear all
		err := cache.ClearAll()
		if err != nil {
			t.Errorf("ClearAll failed: %v", err)
		}

		// Verify all data is cleared
		for _, key := range keys {
			var data string
			success := cache.Get(key, &data)
			if success {
				t.Errorf("Key %s should be cleared", key)
			}
		}
	})
}

func TestCacheFileOperations(t *testing.T) {
	// Create test cache
	tempDir := t.TempDir()
	originalCacheHome := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if originalCacheHome != "" {
			os.Setenv("XDG_CACHE_HOME", originalCacheHome)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
	}()

	os.Setenv("XDG_CACHE_HOME", tempDir)

	cache := New()

	t.Run("Cache Directory Creation", func(t *testing.T) {
		testData := "test data"
		err := cache.Set("test-key", testData, 5*time.Minute)
		if err != nil {
			t.Errorf("Set failed: %v", err)
		}

		// Verify cache directory was created
		cacheDir := filepath.Join(tempDir, "gum")
		if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
			t.Errorf("Cache directory was not created at %s", cacheDir)
		}

		// Verify cache file was created
		cacheFile := filepath.Join(cacheDir, "test-key.json")
		if _, err := os.Stat(cacheFile); os.IsNotExist(err) {
			t.Errorf("Cache file was not created at %s", cacheFile)
		}
	})

	t.Run("Corrupted Cache File", func(t *testing.T) {
		// Create a corrupted cache file
		cacheDir := filepath.Join(tempDir, "gum")
		cacheFile := filepath.Join(cacheDir, "corrupted.json")
		
		err := os.WriteFile(cacheFile, []byte("invalid json"), 0644)
		if err != nil {
			t.Errorf("Failed to create corrupted cache file: %v", err)
		}

		// Try to get from corrupted file
		var data string
		success := cache.Get("corrupted", &data)
		if success {
			t.Error("Get should return false for corrupted cache file")
		}
	})
}

func TestCacheConcurrency(t *testing.T) {
	// Create test cache
	tempDir := t.TempDir()
	originalCacheHome := os.Getenv("XDG_CACHE_HOME")
	defer func() {
		if originalCacheHome != "" {
			os.Setenv("XDG_CACHE_HOME", originalCacheHome)
		} else {
			os.Unsetenv("XDG_CACHE_HOME")
		}
	}()

	os.Setenv("XDG_CACHE_HOME", tempDir)

	cache := New()

	t.Run("Concurrent Set Operations", func(t *testing.T) {
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func(i int) {
				key := fmt.Sprintf("concurrent-key-%d", i)
				data := fmt.Sprintf("concurrent data %d", i)
				err := cache.Set(key, data, 5*time.Minute)
				if err != nil {
					t.Errorf("Concurrent Set failed for key %s: %v", key, err)
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify all data was set
		for i := 0; i < 10; i++ {
			key := fmt.Sprintf("concurrent-key-%d", i)
			var data string
			success := cache.Get(key, &data)
			if !success {
				t.Errorf("Failed to retrieve data for key %s", key)
			}
		}
	})

	t.Run("Concurrent Get Operations", func(t *testing.T) {
		// Set up test data
		testData := "concurrent get test"
		err := cache.Set("get-test-key", testData, 5*time.Minute)
		if err != nil {
			t.Errorf("Set failed: %v", err)
		}

		// Concurrent gets
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				var data string
				success := cache.Get("get-test-key", &data)
				if !success {
					t.Error("Concurrent Get failed")
				}
				if data != testData {
					t.Errorf("Expected '%s', got '%s'", testData, data)
				}
				done <- true
			}()
		}

		// Wait for all goroutines to complete
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}

func TestCacheTTLConstants(t *testing.T) {
	// Test that TTL constants are reasonable
	if ProjectsCacheTTL <= 0 {
		t.Error("ProjectsCacheTTL should be positive")
	}

	if DirsCacheTTL <= 0 {
		t.Error("DirsCacheTTL should be positive")
	}

	if ProjectDirsCacheTTL <= 0 {
		t.Error("ProjectDirsCacheTTL should be positive")
	}

	// Test relative TTL values make sense
	if DirsCacheTTL >= ProjectsCacheTTL {
		t.Error("DirsCacheTTL should be shorter than ProjectsCacheTTL")
	}

	if ProjectsCacheTTL >= ProjectDirsCacheTTL {
		t.Error("ProjectsCacheTTL should be shorter than ProjectDirsCacheTTL")
	}
}