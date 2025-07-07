package main

import (
	"fmt"
	"time"
)

func main() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		idle := getIdleSeconds()
		if idle < 0 {
			fmt.Println("Error getting idle time")
		} else if idle > 300 {
			fmt.Printf("Inactive for %.1f seconds\n", idle)
		} else {
			fmt.Printf("Active (idle %.1f seconds)\n", idle)
		}
	}
}
