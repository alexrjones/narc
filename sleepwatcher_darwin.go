package main

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation
#include "sleepwatcher.h"
extern void onSystemSleep();
extern void onSystemWake();
*/
import "C"
import (
	"fmt"
)

//export onSystemSleep
func onSystemSleep() {
	fmt.Println("[macOS] System is going to sleep")
}

//export onSystemWake
func onSystemWake() {
	fmt.Println("[macOS] System just woke up")
}

func StartSleepWatcher() {
	C.StartSleepWatcher((*[0]byte)(C.onSystemSleep), (*[0]byte)(C.onSystemWake))
}
