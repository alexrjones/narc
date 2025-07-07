package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework ApplicationServices -framework CoreFoundation
#include "listener.h"
extern void activityCallbackBridge();
*/
import "C"
import (
	"context"
	"fmt"
	"os/signal"
	"syscall"
	"time"
)

var lastActive time.Time

//export activityCallbackBridge
func activityCallbackBridge() {
	lastActive = time.Now()
}

func main() {
	lastActive = time.Now()
	go func() {
		for {
			time.Sleep(10 * time.Second)
			inactiveFor := time.Since(lastActive)
			if inactiveFor > 5*time.Minute {
				fmt.Printf("You’ve been inactive for %v\n", inactiveFor)
			} else {
				fmt.Printf("Active (last input %v ago)\n", inactiveFor)
			}
		}
	}()
	C.StartEventTap((*[0]byte)(C.activityCallbackBridge)) // Cast to void*
	defer C.StopEventTap()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	fmt.Println("Quitting")
}
