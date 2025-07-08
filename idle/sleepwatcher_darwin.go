package idle

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

var _onSleepStateChange func(awake bool)

//export onSystemSleep
func onSystemSleep() {
	fmt.Println("[macOS] System is going to sleep")
	_onSleepStateChange(false)
}

//export onSystemWake
func onSystemWake() {
	fmt.Println("[macOS] System just woke up")
	_onSleepStateChange(true)
}

func StartSleepWatcher(onSleepStateChange func(bool)) {
	_onSleepStateChange = onSleepStateChange
	C.StartSleepWatcher((*[0]byte)(C.onSystemSleep), (*[0]byte)(C.onSystemWake))
}
