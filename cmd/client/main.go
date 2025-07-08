package main

import "C"
import (
	"fmt"
	"time"

	"github.com/alexrjones/narc/idle"
)

func main() {
	fmt.Println("Starting sleep/wake watcher...")
	idle.StartSleepWatcher(func(awake bool) {
		if awake {
			fmt.Println("Awake at", time.Now())
		} else {
			fmt.Println("Asleep at", time.Now())
		}
	})
	//
	//ticker := time.NewTicker(10 * time.Second)
	//defer ticker.Stop()
	//for range ticker.C {
	//	idle := getIdleSeconds()
	//	if idle < 0 {
	//		fmt.Println("Error getting idle time")
	//	} else if idle > 300 {
	//		fmt.Printf("Inactive for %.1f seconds\n", idle)
	//	} else {
	//		fmt.Printf("Active (idle %.1f seconds)\n", idle)
	//	}
	//}
}
