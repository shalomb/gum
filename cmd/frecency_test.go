package cmd

import (
	"testing"
	"time"
)

func TestFrecencyScore(t *testing.T) {
	now := time.Now()
	
	tests := []struct {
		name      string
		frequency int
		age       time.Duration
		expectMin int64 // Minimum expected score (behavioral)
	}{
		{
			name:      "Recent high frequency",
			frequency: 100,
			age:       30 * time.Minute, // 30 minutes ago
			expectMin: 1000, // Should be high due to recent + frequent
		},
		{
			name:      "Recent low frequency",
			frequency: 5,
			age:       30 * time.Minute, // 30 minutes ago
			expectMin: 100, // Should be moderate due to recent but low frequency
		},
		{
			name:      "Today high frequency",
			frequency: 50,
			age:       12 * time.Hour, // 12 hours ago
			expectMin: 500, // Should be good due to high frequency but some decay
		},
		{
			name:      "This week",
			frequency: 20,
			age:       3 * 24 * time.Hour, // 3 days ago
			expectMin: 100, // Should be moderate due to age
		},
		{
			name:      "This month",
			frequency: 10,
			age:       15 * 24 * time.Hour, // 15 days ago
			expectMin: 20, // Should be lower due to age
		},
		{
			name:      "Old but frequent",
			frequency: 100,
			age:       2 * 30 * 24 * time.Hour, // 2 months ago
			expectMin: 10, // Should be low due to age but not zero
		},
		{
			name:      "Very old",
			frequency: 5,
			age:       6 * 30 * 24 * time.Hour, // 6 months ago
			expectMin: 1, // Should be very low but not zero
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lastSeen := now.Add(-tt.age)
			score := calculateFrecencyScore(tt.frequency, lastSeen, now)
			
			// Test behavioral properties
			if score < tt.expectMin {
				t.Errorf("calculateFrecencyScore() = %d, want >= %d (too low for %s)", 
					score, tt.expectMin, tt.name)
			}
			
			// Ensure score is never zero (key behavioral property)
			if score <= 0 {
				t.Errorf("calculateFrecencyScore() = %d, want > 0 (score should never be zero)", score)
			}
		})
	}
}

func TestFrecencyScoreProperties(t *testing.T) {
	now := time.Now()
	
	// Test that higher frequency gives higher score (for same age)
	score1 := calculateFrecencyScore(10, now.Add(-1*time.Hour), now)
	score2 := calculateFrecencyScore(20, now.Add(-1*time.Hour), now)
	
	if score2 <= score1 {
		t.Errorf("Higher frequency should give higher score: %d vs %d", score2, score1)
	}
	
	// Test that more recent gives higher score (for same frequency)
	score3 := calculateFrecencyScore(10, now.Add(-1*time.Hour), now)
	score4 := calculateFrecencyScore(10, now.Add(-24*time.Hour), now)
	
	if score4 >= score3 {
		t.Errorf("More recent should give higher score: %d vs %d", score3, score4)
	}
	
	// Test logarithmic scaling (diminishing returns)
	score5 := calculateFrecencyScore(100, now.Add(-1*time.Hour), now)
	score6 := calculateFrecencyScore(200, now.Add(-1*time.Hour), now)
	
	// Score6 should be higher but not double
	if score6 <= score5 {
		t.Errorf("Higher frequency should give higher score: %d vs %d", score6, score5)
	}
	
	// The ratio should be less than 2 (logarithmic scaling)
	ratio := float64(score6) / float64(score5)
	if ratio >= 2.0 {
		t.Errorf("Logarithmic scaling should give diminishing returns: ratio %f", ratio)
	}
}