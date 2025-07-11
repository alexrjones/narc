//go:build darwin

package idle

/*
#cgo LDFLAGS: -framework IOKit -framework CoreFoundation
#include "sleepwatcher_darwin.h"
extern void onSystemSleep();
extern void onSystemWake();
*/
import "C"

var _onSleepStateChange func(awake bool)

//export onSystemSleep
func onSystemSleep() {
	_onSleepStateChange(false)
}

//export onSystemWake
func onSystemWake() {
	_onSleepStateChange(true)
}

func StartSleepWatcher(onSleepStateChange func(bool)) {
	_onSleepStateChange = onSleepStateChange
	C.StartSleepWatcher((*[0]byte)(C.onSystemSleep), (*[0]byte)(C.onSystemWake))
}
