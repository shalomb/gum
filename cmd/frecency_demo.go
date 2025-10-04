package cmd

import (
	"fmt"
	"strings"
	"time"
)

// DemoFrecencyScores demonstrates the frecency algorithm
func DemoFrecencyScores() {
	now := time.Now()
	
	scenarios := []struct {
		name      string
		frequency int
		age       time.Duration
	}{
		{"Active project (now)", 50, 0},
		{"Active project (1 hour ago)", 50, 1 * time.Hour},
		{"Active project (1 day ago)", 50, 24 * time.Hour},
		{"Active project (1 week ago)", 50, 7 * 24 * time.Hour},
		{"Active project (1 month ago)", 50, 30 * 24 * time.Hour},
		{"Active project (6 months ago)", 50, 180 * 24 * time.Hour},
		{"Rare project (now)", 5, 0},
		{"Rare project (1 day ago)", 5, 24 * time.Hour},
		{"Rare project (1 week ago)", 5, 7 * 24 * time.Hour},
		{"Rare project (1 month ago)", 5, 30 * 24 * time.Hour},
		{"Rare project (6 months ago)", 5, 180 * 24 * time.Hour},
		{"Very frequent (now)", 200, 0},
		{"Very frequent (1 day ago)", 200, 24 * time.Hour},
		{"Very frequent (1 week ago)", 200, 7 * 24 * time.Hour},
		{"Very frequent (1 month ago)", 200, 30 * 24 * time.Hour},
		{"Very frequent (6 months ago)", 200, 180 * 24 * time.Hour},
	}
	
	fmt.Println("Frecency Score Demonstration")
	fmt.Println("============================")
	fmt.Printf("%-25s %-15s %-10s %-10s\n", "Scenario", "Age", "Frequency", "Score")
	fmt.Println(strings.Repeat("-", 60))
	
	for _, s := range scenarios {
		lastSeen := now.Add(-s.age)
		score := calculateFrecencyScore(s.frequency, lastSeen, now)
		
		ageStr := formatAge(s.age)
		fmt.Printf("%-25s %-15s %-10d %-10d\n", s.name, ageStr, s.frequency, score)
	}
}

func formatAge(age time.Duration) string {
	if age == 0 {
		return "now"
	}
	
	hours := age.Hours()
	if hours < 1 {
		return fmt.Sprintf("%.0fm ago", age.Minutes())
	} else if hours < 24 {
		return fmt.Sprintf("%.1fh ago", hours)
	} else if hours < 24*7 {
		return fmt.Sprintf("%.1fd ago", hours/24)
	} else if hours < 24*30 {
		return fmt.Sprintf("%.1fw ago", hours/(24*7))
	} else {
		return fmt.Sprintf("%.1fmo ago", hours/(24*30))
	}
}